package main

import (
	"bufio"
	"context"
	"fmt"
	"grpcChat/config"
	chat "grpcChat/pb"
	"grpcChat/rabbitmq"
	"grpcChat/utils"
	"io"
	"os"
	"sync"
	"time"

	"github.com/golang/protobuf/ptypes/timestamp"

	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

var mutex sync.Mutex

func ConsoleLog(res *chat.ChatResponse) {
	mutex.Lock()
	defer mutex.Unlock()
	fmt.Printf("\n------ %s -----\n%s\n> ", res.Time.AsTime().Format("2006-01-02 15:04:05"), res.Message)
}

func Input(prompt string) string {
	fmt.Print(prompt)
	reader := bufio.NewReader(os.Stdin)
	line, _, err := reader.ReadLine()
	if err != nil {
		if err == io.EOF {
			return ""
		} else {
			panic(errors.Wrap(err, "Input"))
		}
	}
	return string(line)
}

func main() {
	rabbitmq.InitRabbitMQ()

	conn, err := grpc.Dial(config.ChatClientAddr, grpc.WithInsecure())
	if !utils.Check(err) {
		panic(err)
	}
	defer conn.Close()

	client := chat.NewChatClient(conn)
	ctx, cancel := context.WithCancel(context.Background())

	username := Input("请输入用户名：")
	ctx = metadata.NewOutgoingContext(ctx, metadata.Pairs("name", username))
	stream, err := client.GoChat(ctx)
	if !utils.Check(err) {
		panic(err)
	}
	fmt.Print("> ")

	// 监听消息队列
	go rabbitmq.Consume(username, ConsoleLog)

	// 监听服务端通知
	go func() {
		var (
			reply *chat.ChatResponse
			err   error
		)
		for {
			reply, err = stream.Recv()
			if !utils.Check(err) {
				cancel()
				break
			}
			ConsoleLog(reply)
		}
	}()

	go func() {
		var (
			line string
			err  error
		)
		for {
			line = Input("")
			if line == "exit" {
				cancel()
				break
			}
			err = stream.Send(&chat.ChatRequest{
				Username: username,
				Message:  line,
				Time:     &timestamp.Timestamp{Seconds: time.Now().Unix()},
			})
			if !utils.Check(err) {
				cancel()
				break
			}
			fmt.Print("> ")
		}
	}()

	<-ctx.Done()
}
