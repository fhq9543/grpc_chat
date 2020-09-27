docker:
	curl -fsSL https://get.docker.com | bash -s docker --mirror Aliyun

prepare: docker
	docker pull redis
	docker run -p 26379:6379 --restart always --name myredis -v `pwd`/dockerData:/data -d redis redis-server --requirepass "xiniao666" --appendonly yes
	docker pull rabbitmq:management
	docker run -p 15672:15672 -p 5672:5672 --restart always --name myrabbit -d --hostname myrabbit -e RABBITMQ_DEFAULT_USER=xiniao -e RABBITMQ_DEFAULT_PASS=xiniao666 rabbitmq:management

runserver:
	GO111MODULE=on go run server/server.go

runclient:
	GO111MODULE=on go run client/client.go

build:
	GO111MODULE=on go build -o chatServer server/server.go
	GO111MODULE=on go build -o chatClient client/client.go

test:
	go test -v ./...

