package errs

import "fmt"

const (
	ShardLimitCode = 10000 + iota
	URLInvalidCode
)

var (
	ErrShardLimit = New(ShardLimitCode, "shard over limit")
	ErrURLInvalid = New(ShardLimitCode, "session url is empty")
)

type Err struct {
	code int
	text string
}

func (e Err) Error() string {
	return fmt.Sprintf("code:%v, text:%v", e.code, e.text)
}

func New(code int, text string) error {
	return &Err{
		code: code,
		text: text,
	}
}
