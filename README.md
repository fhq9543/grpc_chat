###A grpc chat demo.

#### Usage：
##### 开发环境准备，处理docker安装、启动mariadb、redis、rabbitmq，挂载数据卷、完成初始化配置等
> make prepare

##### 单元测试
> make test

##### 运行程序
> make runserver
>
> make runclient

##### 编译程序
> make build

#### 查看聊天记录：
##### historyPublic：在聊天窗口中输入命令，获取公共聊天室聊天记录
##### historyPrivate：在聊天窗口中输入命令，获取私聊聊天记录
