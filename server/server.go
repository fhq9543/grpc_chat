package main

import (
	"fmt"
	chat "grpcChat/pb"
	"grpcChat/utils"
	"net"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/golang/protobuf/ptypes/timestamp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type Service struct{}

type ConnectPool struct {
	sync.Map
}

var connectPool *ConnectPool

func (p *ConnectPool) Get(name string) chat.Chat_SayServer {
	if stream, ok := p.Load(name); ok {
		return stream.(chat.Chat_SayServer)
	} else {
		return nil
	}
}

func (p *ConnectPool) Add(name string, stream chat.Chat_SayServer) {
	p.Store(name, stream)
}

func (p *ConnectPool) Del(name string) {
	p.Delete(name)
}

func (p *ConnectPool) BroadCast(from, message string) {
	utils.Debug(message)
	p.Range(func(username_i, stream_i interface{}) bool {
		username := username_i.(string)
		stream := stream_i.(chat.Chat_SayServer)
		if username == from {
			return true
		} else {
			stream.Send(&chat.SayResponse{
				Message: message,
				SayTime: &timestamp.Timestamp{Seconds: time.Now().Unix()},
			})
		}
		return true
	})
}

func (p *ConnectPool) PrivateChat(from, to, message string) {
	// 判断用户是否在线
	steam := connectPool.Get(to)
	if steam != nil {
		utils.Debug(fmt.Sprintf("%s 私聊 %s：%s", from, to, message))
		steam.Send(&chat.SayResponse{
			Message: fmt.Sprintf("%s 私聊：%s", from, message),
			SayTime: &timestamp.Timestamp{Seconds: time.Now().Unix()},
		})
	} else {
		utils.Debug(fmt.Sprintf("%s 不在线，已转为离线消息！", to))
	}
}

func (s *Service) Say(stream chat.Chat_SayServer) error {
	md, _ := metadata.FromIncomingContext(stream.Context())
	username := md["name"][0]
	if connectPool.Get(username) != nil {
		stream.Send(&chat.SayResponse{
			Message: fmt.Sprintf("用户名 %s 已存在！", username),
		})
		return nil
	} else { // 连接成功
		connectPool.Add(username, stream)
		stream.Send(&chat.SayResponse{
			Message: fmt.Sprintf("连接成功！"),
		})
		// 检查是否有离线消息

	}
	go func() {
		<-stream.Context().Done()
		connectPool.Del(username)
		connectPool.BroadCast(username, fmt.Sprintf("%s 离线了！", username))
	}()
	connectPool.BroadCast(username, fmt.Sprintf("欢迎 %s！", username))
	for {
		req, err := stream.Recv()
		if err != nil {
			return err
		}

		// 判断是否为私聊。私聊格式   @用户名 消息
		reg := regexp.MustCompile(`^@(.+)\s`)
		if reg != nil {
			list := reg.FindAllStringSubmatch(req.Message, -1)
			if len(list) > 0 && len(list[0]) >= 2 {
				toUser := list[0][1]
				connectPool.PrivateChat(username, toUser, strings.TrimPrefix(req.Message, list[0][0]))
				continue
			}
		}

		connectPool.BroadCast(username, fmt.Sprintf("%s：%s", username, req.Message))
	}
	return nil
}

func main() {
	connectPool = &ConnectPool{}

	rpcServer := grpc.NewServer()
	chat.RegisterChatServer(rpcServer, &Service{})

	addr := ":18881"
	if len(os.Args) >= 2 {
		addr = os.Args[1]
	}
	listen, err := net.Listen("tcp", addr)
	if !utils.Check(err) {
		panic(err)
	}
	utils.Debug(fmt.Sprintf("开始监听： %s", addr))

	utils.Check(rpcServer.Serve(listen))
}
