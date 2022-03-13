package client

import (
	"context"
	"math"
	"relationship-robot/dto"
	"relationship-robot/errs"
	"relationship-robot/log"
	"relationship-robot/util"
	"time"
)

type Cfg struct {
	WSCfg  WSClientCfg  // 其中的Handler是数据的钩子，不影响数据的分发 TODO：抽象成 interface
	ApiCfg ApiClientCfg // TODO：抽象成 interface
}

type Client struct {
	Ctx       context.Context // TODO：传递控制
	Cfg       Cfg
	handler   dto.MsgHandler
	token     string
	apiClient *ApiClient
	wsClients []*WSClient
}

func New(ctx context.Context, token string, cfg Cfg) *Client {
	apiClient := NewApiClient(token, cfg.ApiCfg)
	client := &Client{
		Ctx:       ctx,
		token:     token,
		apiClient: apiClient,
		handler:   cfg.WSCfg.Handler,
		Cfg:       cfg,
	}
	client.Cfg.WSCfg.Handler = client.DistributeMsg

	return client
}

func (c *Client) Start(ctx context.Context, handlers ...interface{}) error {
	c.Cfg.WSCfg.Intent = RegisterHandler(handlers...)
	wsInfo, err := c.apiClient.WS(ctx) // TODO: 重置计数时间到后，动态关闭或者申请 wsClient
	if err != nil {
		return err
	}
	log.Println("client will create shard info:", util.ToString(wsInfo))
	if err = checkShardInfo(wsInfo); err != nil {
		return err
	}

	interval := calcInterval(wsInfo.SessionStartLimit.MaxConcurrency)
	wsCfg := c.Cfg.WSCfg
	for i := uint32(0); i < wsInfo.Shards; i++ {
		wsCfg.ShardCfg = dto.ShardConfig{
			ShardID:    i,
			ShardCount: wsInfo.Shards,
		}
		var wsClient *WSClient
		wsClient, err = NewAndStartWSClient(c.token, wsInfo.URL, wsCfg)
		if err != nil {
			log.Errorf("new and start WSClient failed! token: %s, url: %s, cfg: %s, err: %s",
				c.token, wsInfo.URL, util.ToString(wsCfg), err)
			continue
		}
		c.wsClients = append(c.wsClients, wsClient)
		time.Sleep(interval)
	}

	if len(c.wsClients) == 0 {
		return err
	}

	return nil
}

func (c *Client) DistributeMsg(msg *dto.WSPayload) {
	if c.handler != nil {
		func(handler dto.MsgHandler) {
			defer func() {
				if err := recover(); err != nil {
					log.Errorf("custom handler process stream msg failed! err: %s, msg %s", err, util.ToString(msg))
				}
			}()
			handler(msg)
		}(c.handler)
	}
	DistributeMsg(c.apiClient, msg)
}

func checkShardInfo(wsInfo *dto.WebsocketAP) error {
	if wsInfo.Shards > wsInfo.SessionStartLimit.Remaining {
		return errs.ErrShardLimit
	}
	return nil
}

func calcInterval(maxConcurrency uint32) time.Duration {
	if maxConcurrency == 0 {
		maxConcurrency = 1
	}
	f := math.Round(concurrencyTimeWindowSec / float64(maxConcurrency))
	if f == 0 {
		f = 1
	}
	return time.Duration(f) * time.Second
}
