package util

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/tidwall/gjson"
)

func Parse(msg []byte, path string) []byte {
	return []byte(gjson.Get(string(msg), "d").String())
}

func parseData(message []byte, target interface{}) error {
	data := gjson.Get(string(message), "d")
	return json.Unmarshal([]byte(data.String()), target)
}

func Emoji(ID int) string {
	return fmt.Sprintf("<emoji:%d>", ID)
}

var removeAtRe = regexp.MustCompile(`<@!\d+>`)

const spaceCharSet = " \u00A0"

func ETLInput(input string) string {
	etlData := string(removeAtRe.ReplaceAll([]byte(input), []byte("")))
	etlData = strings.Trim(etlData, spaceCharSet)
	return etlData
}
