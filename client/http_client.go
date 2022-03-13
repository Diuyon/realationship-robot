package client

import (
	"context"
	"relationship-robot/dto"
	"relationship-robot/errs"
	"relationship-robot/log"
	"relationship-robot/util"
	"relationship-robot/version"
	"time"

	"github.com/go-resty/resty/v2"
)

type ApiClientCfg struct {
	Sandbox bool
	Timeout time.Duration
}

type ApiClient struct {
	token      string
	sandbox    bool
	httpclient *resty.Client

	cfg ApiClientCfg
}

func NewApiClient(token string, cfg ApiClientCfg) *ApiClient {
	if cfg.Timeout == 0 {
		cfg.Timeout = 3 * time.Second
	}
	return &ApiClient{
		token:   token,
		sandbox: cfg.Sandbox,
		cfg:     cfg,
		httpclient: resty.New().
			SetTimeout(cfg.Timeout).
			SetAuthScheme("Bot").SetAuthToken(token).
			SetHeader("User-Agent", version.String()).
			OnAfterResponse(
				func(client *resty.Client, resp *resty.Response) error {
					log.Printf("%v", util.FormatResp(resp))
					// 非成功含义的状态码，需要返回 error 供调用方识别
					if !IsSuccessStatus(resp.StatusCode()) {
						return errs.New(resp.StatusCode(), string(resp.Body()))
					}
					return nil
				},
			),
	}
}

func (o *ApiClient) request(ctx context.Context) *resty.Request {
	return o.httpclient.R().SetContext(ctx)
}

func (o *ApiClient) PostMessage(ctx context.Context, channelID string, msg *dto.MessageToCreate) (*dto.Message, error) {
	resp, err := o.request(ctx).
		SetResult(dto.Message{}).
		SetPathParam("channel_id", channelID).
		SetBody(msg).
		Post(o.getURL(messagesURI))
	if err != nil {
		return nil, err
	}

	return resp.Result().(*dto.Message), nil
}

func (o *ApiClient) WS(ctx context.Context) (*dto.WebsocketAP, error) {
	resp, err := o.request(ctx).
		SetResult(dto.WebsocketAP{}).
		Get(o.getURL(gatewayBotURI))
	if err != nil {
		return nil, err
	}

	return resp.Result().(*dto.WebsocketAP), nil
}
