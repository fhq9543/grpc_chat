package rabbitmq

import (
	"encoding/json"
	"fmt"
	chat "grpcChat/pb"
	"grpcChat/utils"
	"strconv"
	"testing"
	"time"

	"github.com/golang/protobuf/ptypes/timestamp"
)

func TestRMQ(t *testing.T) {
	InitRabbitMQ()
	go Consume("cc", func(res *chat.ChatResponse) {
		utils.Debug(fmt.Sprintf("receive %s", res.Message))
	})
	time.Sleep(time.Second)
	for i := 0; i < 5; i++ {
		msg := &chat.ChatResponse{
			Username: strconv.Itoa(i),
			Message:  strconv.Itoa(i),
			Time:     &timestamp.Timestamp{Seconds: time.Now().Unix()},
		}
		msgData, _ := json.Marshal(msg)
		Broadcast(msgData)
		utils.Debug("send: ", i)
		time.Sleep(time.Millisecond * 500)
	}
}
