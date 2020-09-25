package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	chat "grpcChat/pb"
	"grpcChat/utils"
	"io"
	"os"
	"sync"
	"time"

	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

var name *string = flag.String("name", "guess", "what's your name?")
var mutex sync.Mutex

func ConsoleLog(message string) {
	mutex.Lock()
	defer mutex.Unlock()
	fmt.Printf("\n------ %s -----\n%s\n> ", time.Now().Format("2006-01-02 15:04:05"), message)
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
	conn, err := grpc.Dial(":18881", grpc.WithInsecure())
	if !utils.Check(err) {
		panic(err)
	}
	defer conn.Close()

	client := chat.NewChatClient(conn)
	ctx, cancel := context.WithCancel(context.Background())

	username := Input("请输入用户名：")
	ctx = metadata.NewOutgoingContext(ctx, metadata.Pairs("name", username))
	stream, err := client.Say(ctx)
	if !utils.Check(err) {
		panic(err)
	}

	// 监听服务端通知
	go func() {
		var (
			reply *chat.SayResponse
			err   error
		)
		for {
			reply, err = stream.Recv()
			if !utils.Check(err) {
				cancel()
				break
			}
			ConsoleLog(reply.Message)
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
			err = stream.Send(&chat.SayRequest{
				Message: line,
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
