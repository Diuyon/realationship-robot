package dto

import (
	"fmt"
)

// WebsocketAP wss 接入点信息
type WebsocketAP struct {
	URL               string            `json:"url"`
	Shards            uint32            `json:"shards"`
	SessionStartLimit SessionStartLimit `json:"session_start_limit"`
}

// SessionStartLimit 链接频控信息
type SessionStartLimit struct {
	Total          uint32 `json:"total"`
	Remaining      uint32 `json:"remaining"`
	ResetAfter     uint32 `json:"reset_after"`
	MaxConcurrency uint32 `json:"max_concurrency"`
}

type ShardConfig struct {
	ShardID    uint32
	ShardCount uint32
}

func (s ShardConfig) String() string {
	return fmt.Sprintf("[shard_id:%d-%d]", s.ShardID, s.ShardCount)
}

type Session struct {
	Version   int    `json:"version"`
	ID        string `json:"id"`
	LastSeqId uint32 `json:"last_seq_id"`
}

func (s *Session) String() string {
	if s == nil {
		return ""
	}
	return fmt.Sprintf("[session_id:%s, last_seq_id:%d]", s.ID, s.LastSeqId)
}

type MsgHandler func(payload *WSPayload)
