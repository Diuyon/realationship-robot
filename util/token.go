package util

import "fmt"

func GetToken(appId int64, appToken string) string {
	return fmt.Sprintf("%v.%s", appId, appToken)
}
