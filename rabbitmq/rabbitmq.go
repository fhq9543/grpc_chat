package rabbitmq

import (
	"encoding/json"
	"grpcChat/config"
	chat "grpcChat/pb"
	"grpcChat/utils"

	"github.com/pkg/errors"

	"github.com/streadway/amqp"
)

var broadcastChannel *amqp.Channel
var rabbitConn *amqp.Connection

func InitRabbitMQ() error {
	var err error
	// 创建链接
	rabbitConn, err = amqp.Dial(config.RMQAddr)
	if err != nil {
		return errors.Wrap(err, "Failed to Dial")
	}

	// 打开一个通道
	broadcastChannel, err = rabbitConn.Channel()
	if err != nil {
		return errors.Wrap(err, "Failed to openChannel")
	}

	// 生成一个交换机（交换机不存在的情况下）
	err = broadcastChannel.ExchangeDeclare(
		config.RMQExchangeName,
		amqp.ExchangeFanout,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return errors.Wrap(err, "Failed to declare an exchange")
	}
	return nil
}

func GetRabbitMQChannel() *amqp.Channel {
	return broadcastChannel
}

func CloseRabbitMQ() {
	broadcastChannel.Close()
	rabbitConn.Close()
}

func Broadcast(msgData []byte) error {
	//utils.Debug(string(msgData))

	// 指定交换机发布消息
	err := broadcastChannel.Publish(config.RMQExchangeName, "", false, false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        msgData,
		})
	if err != nil {
		return errors.Wrap(err, "Failed to publish message")
	}
	return nil
}

func Consume(username string, precess func(*chat.ChatResponse)) {
	q, err := broadcastChannel.QueueDeclare(
		"",    // name
		false, // durable
		false, // delete when usused
		true,  // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if !utils.Check(err) {
		return
	}

	err = broadcastChannel.QueueBind(
		q.Name,                 // queue name
		"",                     // routing key
		config.RMQExchangeName, // exchange
		false,
		nil,
	)
	if !utils.Check(err) {
		return
	}

	msgs, err := broadcastChannel.Consume(
		q.Name,
		"",
		true, //Auto Ack
		false,
		false,
		false,
		nil,
	)
	if !utils.Check(err) {
		return
	}

	for msg := range msgs {
		res := &chat.ChatResponse{}
		err = json.Unmarshal(msg.Body, res)
		if !utils.Check(err) {
			return
		}

		if res.Username != username {
			precess(res)
		}
	}
}
