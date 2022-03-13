package client

import (
	"encoding/json"
	"relationship-robot/dto"
	"relationship-robot/log"
	"relationship-robot/util"
)

type (
	AtMessageHandler       = func(*ApiClient, *dto.WSATMessageData)
	ReactionMessageHandler = func(*ApiClient, *dto.MessageReaction)
)

var (
	atMessageFn       = func(*ApiClient, *dto.WSATMessageData) {}
	reactionMessageFn = func(*ApiClient, *dto.MessageReaction) {}

	defaultHandlers = []interface{}{atMessageFn, reactionMessageFn}
)

func RegisterHandler(handlers ...interface{}) dto.Intent {
	if len(handlers) == 0 {
		handlers = defaultHandlers
	}
	var i dto.Intent
	for _, handler := range handlers {
		switch h := handler.(type) {
		case AtMessageHandler:
			i |= dto.EventToIntent(dto.EventAtMessageCreate)
			atMessageFn = h
		case ReactionMessageHandler:
			i |= dto.EventToIntent(dto.EventMessageReactionAdd, dto.EventMessageReactionRemove)
			reactionMessageFn = h
		}
	}

	return i
}

func DistributeMsg(c *ApiClient, msg *dto.WSPayload) {
	switch msg.OPCode {
	case dto.WSDispatchEvent:
		DistributeDispatchEvent(c, msg)
	}
}

func DistributeDispatchEvent(c *ApiClient, msg *dto.WSPayload) {
	switch msg.Type {
	case dto.EventAtMessageCreate:
		data := new(dto.WSATMessageData)
		if unmarshalData(msg, data) {
			atMessageFn(c, data)
		}
	case dto.EventMessageReactionAdd, dto.EventMessageReactionRemove:
		data := new(dto.MessageReaction)
		if unmarshalData(msg, data) {
			reactionMessageFn(c, data)
		}
	}
}

func unmarshalData(msg *dto.WSPayload, data interface{}) bool {
	if err := json.Unmarshal(msg.D, data); err != nil {
		log.Errorln("DistributeDispatchEvent Unmarshal msg err:", err, "msg:", util.ToString(msg))
		return false
	}
	return true
}
