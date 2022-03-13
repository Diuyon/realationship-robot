package business

import (
	"context"
	"log"
	"relationship-robot/client"
	"relationship-robot/dto"
	"relationship-robot/util"
)

func ReturnAtMessage(c *client.ApiClient, data *dto.WSATMessageData) {
	toCreate := &dto.MessageToCreate{
		MessageReference: &dto.MessageReference{
			// 引用这条消息
			MessageID:             data.ID,
			IgnoreGetMessageError: true,
		},
	}

	content := util.ETLInput(data.Content)
	if content == "" {
		toCreate.Content = welcomeText + defaultReply
		goto POST
	}
	for _, e := range []rune(content) {
		if humanMap[string(e)] == "" {
			toCreate.Content = WarnText + defaultReply
			goto POST
		}
	}

	toCreate.Content = "答案为: "
	if humanMap[content] != "" {
		toCreate.Content += humanMap[content]
	} else if relationMap[content] != "" {
		toCreate.Content += relationMap[content]
	} else {
		toCreate.Content += dontKnow
	}

POST:
	if _, err := c.PostMessage(context.Background(), data.ChannelID, toCreate); err != nil {
		log.Println("ReturnAtMessage PostMessage err:", err)
	}
}
