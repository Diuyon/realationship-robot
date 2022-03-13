package client

import (
	"encoding/json"
	"fmt"
	"relationship-robot/dto"
	"relationship-robot/errs"
	"relationship-robot/log"
	"relationship-robot/util"
	"time"

	"go.uber.org/atomic"

	"github.com/pkg/errors"

	"github.com/gorilla/websocket"
)

const DefaultQueueSize = 1000

type WSClientCfg struct {
	Intent   dto.Intent
	ShardCfg dto.ShardConfig
	Handler  dto.MsgHandler
}

type WSClient struct {
	token           string
	intent          dto.Intent
	url             string
	handler         dto.MsgHandler
	shadCfg         dto.ShardConfig
	conn            *websocket.Conn
	session         *dto.Session
	msgCh           chan *dto.WSPayload // 数据源头关闭
	ErrCh           chan error          // 不需要关闭，WSClient 结束自动释放，无缓冲 chan，更高效的处理错误，减少双方资源浪费
	closeCh         chan struct{}
	isClosed        *atomic.Bool
	heartbeatTicker *time.Ticker
}

func NewWSClient(token, url string, cfg WSClientCfg) *WSClient {
	if cfg.Intent == 0 {
		cfg.Intent = dto.IntentGuilds
	}
	if cfg.Handler == nil {
		cfg.Handler = func(payload *dto.WSPayload) {} // TODO: 默认回调打印上下文信息更好
	}
	return &WSClient{
		token:           token,
		url:             url,
		handler:         cfg.Handler,
		intent:          cfg.Intent,
		shadCfg:         cfg.ShardCfg,
		msgCh:           make(chan *dto.WSPayload, DefaultQueueSize),
		ErrCh:           make(chan error),
		closeCh:         make(chan struct{}),
		isClosed:        atomic.NewBool(false),
		heartbeatTicker: time.NewTicker(60 * time.Second),
	}
}

func NewAndStartWSClient(token, url string, cfg WSClientCfg) (*WSClient, error) {
	client := NewWSClient(token, url, cfg)
	err := client.Connect()
	if err != nil {
		return nil, err
	}
	client.Identify()
	go client.Listen()
	return client, nil
}

func (c *WSClient) Connect() error {
	if c.url == "" {
		return errs.ErrURLInvalid
	}

	var err error
	c.conn, _, err = websocket.DefaultDialer.Dial(c.url, nil)
	if err != nil {
		log.Errorf("%s connect err: %v, url: %v", c.shadCfg, err, c.url)
		return err
	}
	log.Printf("%s url %v, connected", c.shadCfg, c.url)

	return nil
}

func (c *WSClient) ReadMsg() {
	defer close(c.msgCh) // 在哪儿产生数据，就得在哪儿关掉
	for !c.isClosed.Load() {
		_, msg, err := c.conn.ReadMessage()
		if err != nil {
			c.ErrCh <- errors.Wrap(err, fmt.Sprintf("%s %s session read message error", c.shadCfg, c.session))
		}

		event := new(dto.WSPayload)
		if err := json.Unmarshal(msg, event); err != nil {
			log.Errorf("%s %s session read msg unmarshal error: %s, msg: %s", c.shadCfg, c.session, err, string(msg))
			continue
		}
		event.RawMessage = msg
		event.D = util.Parse(msg, "d")

		log.Printf("%s %s receive %s message, %s", c.shadCfg, c.session, dto.OPMeans(event.OPCode), string(msg))
		c.processEventByOp(event)
	}
}

func (c *WSClient) CallBackMsg() {
	for data := range c.msgCh {
		c.saveSeq(data.Seq)
		func() {
			defer func() {
				if err := recover(); err != nil {
					log.Errorf("%s session callback msg handler error! err:", c.session, err)
				}
			}()
			c.handler(data)
		}()
	}
}

func (c *WSClient) Write(message *dto.WSPayload) {
	m, _ := json.Marshal(message)
	log.Printf("%s %s write %s message, %v", c.shadCfg, c.session, dto.OPMeans(message.OPCode), string(m))

	if err := c.conn.WriteMessage(websocket.TextMessage, m); err != nil {
		c.ErrCh <- errors.Wrap(err, fmt.Sprintf("%s WriteMessage failed", c.session))
	}
}

func (c *WSClient) processEventByOp(event *dto.WSPayload) {
	switch event.OPCode {
	case dto.WSHello: // 接收到 hello 后需要开始发心跳
		c.startHeartBeat(event.D)
	case dto.WSReconnect: // 达到连接时长，需要重新连接，此时可以通过 resume 续传原连接上的事件
		c.Resume()
	case dto.WSInvalidSession: // 无效的 sessionLog，需要重新鉴权
		c.Identify()
	default:
		c.processEventByType(event)
	}
}

func (c *WSClient) processEventByType(event *dto.WSPayload) {
	switch event.Type {
	case dto.EventReady:
		c.readySession(event)
		return
	default:
	}
	c.msgCh <- event
}

func (c *WSClient) startHeartBeat(data []byte) {
	helloData := new(dto.WSHelloData)
	if err := json.Unmarshal(data, helloData); err != nil {
		log.Errorf("%s %s msg data unmarshal hello_data error: %s", c.shadCfg, c.session, err)
	} else {
		c.heartbeatTicker.Reset(time.Duration(helloData.HeartbeatInterval) * time.Millisecond)
	}
}

func (c *WSClient) Listen() {
	defer c.Close()
	go c.ReadMsg()
	go c.CallBackMsg()

	for {
		select {
		case err := <-c.ErrCh:
			c.processErr(err)
		case <-c.heartbeatTicker.C:
			c.heartbeat()
		case <-c.closeCh:
			return
		}
	}
}

// Resume 重连
func (c *WSClient) Resume() {
	event := &dto.WSPayload{
		Data: &dto.WSResumeData{
			Token:     c.token,
			SessionID: c.session.ID,
			Seq:       c.session.LastSeqId,
		},
	}
	event.OPCode = dto.WSResume // 内嵌结构体字段，单独赋值
	c.Write(event)
}

// Identify 对一个连接进行鉴权，并声明监听的 shard 信息
func (c *WSClient) Identify() {
	event := &dto.WSPayload{
		Data: &dto.WSIdentityData{
			Token:   c.token,
			Intents: c.intent,
			Shard: []uint32{
				c.shadCfg.ShardID,
				c.shadCfg.ShardCount,
			},
		},
	}
	event.OPCode = dto.WSIdentity

	log.Println(c.shadCfg, "start Identify, post event:", util.ToString(event))
	c.Write(event)
}

func (c *WSClient) heartbeat() {
	log.Debugf("%s %s listened heartBeat", c.shadCfg, c.session)
	heartBeatEvent := &dto.WSPayload{
		WSPayloadBase: dto.WSPayloadBase{
			OPCode: dto.WSHeartbeat,
		},
		Data: c.session.LastSeqId,
	}
	c.Write(heartBeatEvent)
}

func (c *WSClient) readySession(event *dto.WSPayload) {
	readyData := &dto.WSReadyData{}
	if err := json.Unmarshal(event.D, &readyData); err != nil {
		log.Errorf("%s %s unmarshal ready_data failed! err: %s, message: %v", c.shadCfg, c.session, err, util.ToString(event))
	}
	c.session = new(dto.Session)
	c.session.Version = readyData.Version
	// 基于 ready 事件，更新 session 信息
	c.session.ID = readyData.SessionID
	c.session.LastSeqId = event.Seq
	c.shadCfg.ShardID = readyData.Shard[0]
	c.shadCfg.ShardCount = readyData.Shard[1]
}

func (c *WSClient) processErr(err error) {
	log.Errorf("%s %s session happened error! err: %s", c.shadCfg, c.session, err)
	err = errors.Cause(err)
	if websocket.IsCloseError(err, 4914, 4915) {
		c.shutdown()
	} else if websocket.IsCloseError(err, 4009) {
		c.Resume()
	} else {
		c.Identify()
	}
}

func (c *WSClient) saveSeq(seq uint32) {
	if seq > 0 {
		c.session.LastSeqId = seq
	}
}

func (c *WSClient) Close() {
	if err := c.conn.Close(); err != nil {
		log.Errorf("%s %s, close conn err: %v", c.shadCfg, c.session, err)
	}
	c.heartbeatTicker.Stop()
	c.shutdown()
}

func (c *WSClient) shutdown() {
	if c.isClosed.CAS(false, true) {
		close(c.closeCh)
	}
}
