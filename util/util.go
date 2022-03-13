package util

import "encoding/json"

func ToString(v interface{}) string {
	c, _ := json.Marshal(v)
	return string(c)
}
