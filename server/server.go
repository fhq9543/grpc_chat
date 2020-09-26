package main

import (
	"encoding/json"
	"fmt"
	"grpcChat/config"
	chat "grpcChat/pb"
	"grpcChat/rabbitmq"
	"grpcChat/redis"
	"grpcChat/utils"
	"net"
	"regexp"
	"strings"
	"sync"
	"time"

	"google.golang.org/grpc"

	"github.com/golang/protobuf/ptypes/timestamp"
	"google.golang.org/grpc/metadata"
)

type Service struct{}

type ConnectPool struct {
	sync.Map
}

var connectPool *ConnectPool

func main() {
	rabbitmq.InitRabbitMQ()
	redis.InitRedis()
	connectPool = &ConnectPool{}

	rpcServer := grpc.NewServer()
	chat.RegisterChatServer(rpcServer, &Service{})

	addr := config.ChatServerAddr
	listen, err := net.Listen("tcp", addr)
	if !utils.Check(err) {
		panic(err)
	}
	utils.Debug(fmt.Sprintf("开始监听： %s", addr))

	utils.Check(rpcServer.Serve(listen))
}

func (s *Service) GoChat(stream chat.Chat_GoChatServer) error {
	md, _ := metadata.FromIncomingContext(stream.Context())
	username := md["name"][0]
	if connectPool.Get(username) != nil {
		stream.Send(&chat.ChatResponse{
			Message: fmt.Sprintf("用户名 %s 已存在！", username),
			Time:    &timestamp.Timestamp{Seconds: time.Now().Unix()},
		})
		return nil
	} else { // 连接成功
		connectPool.Add(username, stream)

		// 检查是否有离线消息
		resList := redis.RedisPopList(config.RedisDBPrefixHistoryUnread + username)
		utils.Debug(fmt.Sprintf("RedisPopList	key:xiniaoList, value:%s", resList))
		for _, v := range resList {
			sendRes := &chat.ChatResponse{}
			err := json.Unmarshal([]byte(v), sendRes)
			if !utils.Check(err) {
				return err
			}

			stream.Send(sendRes)
		}
	}
	go func() {
		<-stream.Context().Done()
		connectPool.Del(username)
		msg := &chat.ChatResponse{
			Username: username,
			Message:  fmt.Sprintf("%s 离线了！", username),
			Time:     &timestamp.Timestamp{Seconds: time.Now().Unix()},
		}
		msgData, _ := json.Marshal(msg)
		rabbitmq.Broadcast(msgData)
	}()

	// 保存，发送消息
	msg := &chat.ChatResponse{
		Username: username,
		Message:  fmt.Sprintf("欢迎 %s！", username),
		Time:     &timestamp.Timestamp{Seconds: time.Now().Unix()},
	}
	msgData, _ := json.Marshal(msg)
	rabbitmq.Broadcast(msgData)
	for {
		req, err := stream.Recv()
		if err != nil {
			return err
		}

		// history
		switch req.Message {
		case config.RedisDBPrefixHistoryPublic:
			resList, err := redis.RedisGetList(config.RedisDBPrefixHistoryPublic)
			if utils.Check(err) {
				utils.Debug(fmt.Sprintf("RedisGetList	key:%s, value:%s", config.RedisDBPrefixHistoryPublic, resList))
				for _, v := range resList {
					sendRes := &chat.ChatResponse{}
					err = json.Unmarshal([]byte(v), sendRes)
					if !utils.Check(err) {
						return err
					}
					stream.Send(sendRes)
				}
			}
			continue
		case config.RedisDBPrefixHistoryPrivate:
			resList, err := redis.RedisGetList(config.RedisDBPrefixHistoryPrivate + username)
			if utils.Check(err) {
				utils.Debug(fmt.Sprintf("RedisGetList	key:%s, value:%s", config.RedisDBPrefixHistoryPrivate+username, resList))
				for _, v := range resList {
					sendRes := &chat.ChatResponse{}
					err = json.Unmarshal([]byte(v), sendRes)
					if !utils.Check(err) {
						return err
					}
					stream.Send(sendRes)
				}
			}
			continue
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

		msg = &chat.ChatResponse{
			Username: username,
			Message:  fmt.Sprintf("%s：%s", username, req.Message),
			Time:     &timestamp.Timestamp{Seconds: time.Now().Unix()},
		}
		msgData, _ = json.Marshal(msg)
		// 保存到redis
		err = redis.RedisRPush(config.RedisDBPrefixHistoryPublic, string(msgData))
		if !utils.Check(err) {
			return err
		}
		rabbitmq.Broadcast(msgData)
	}
	return nil
}

func (p *ConnectPool) Get(name string) chat.Chat_GoChatServer {
	if stream, ok := p.Load(name); ok {
		return stream.(chat.Chat_GoChatServer)
	} else {
		return nil
	}
}

func (p *ConnectPool) Add(name string, stream chat.Chat_GoChatServer) {
	p.Store(name, stream)
}

func (p *ConnectPool) Del(name string) {
	p.Delete(name)
}

func (p *ConnectPool) BroadCast(from, message string) {
	utils.Debug(message)
	p.Range(func(username_i, stream_i interface{}) bool {
		username := username_i.(string)
		stream := stream_i.(chat.Chat_GoChatServer)
		if username == from {
			return true
		} else {
			stream.Send(&chat.ChatResponse{
				Message: message,
				Time:    &timestamp.Timestamp{Seconds: time.Now().Unix()},
			})
		}
		return true
	})
}

func (p *ConnectPool) PrivateChat(from, to, message string) {
	msg := &chat.ChatResponse{
		Username: from,
		Message:  fmt.Sprintf("私聊 %s：%s", from, message),
		Time:     &timestamp.Timestamp{Seconds: time.Now().Unix()},
	}
	msgData, _ := json.Marshal(msg)
	// 保存到redis
	err := redis.RedisRPush(config.RedisDBPrefixHistoryPrivate+to, string(msgData))
	if !utils.Check(err) {
		return
	}
	err = redis.RedisRPush(config.RedisDBPrefixHistoryPrivate+from, string(msgData))
	if !utils.Check(err) {
		return
	}

	// 判断用户是否在线
	steam := connectPool.Get(to)
	if steam != nil {
		utils.Debug(fmt.Sprintf("%s 私聊 %s：%s", from, to, message))
		steam.Send(msg)
	} else {
		utils.Debug(fmt.Sprintf("%s 不在线，已转为离线消息！", to))
		// 保存到redis
		err := redis.RedisRPush(config.RedisDBPrefixHistoryUnread+to, string(msgData))
		if !utils.Check(err) {
			return
		}
	}
}
