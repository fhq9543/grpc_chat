package config

const (
	ChatServerAddr = ":18881"
	ChatClientAddr = ":18881"

	RMQAddr         = "amqp://xiniao:xiniao666@127.0.0.1:5672/"
	RMQExchangeName = "chatExchange"
	RMQQueueName    = "chatQueue"

	RedisAddr                   = ":26379"
	RedisPassword               = "xiniao666"
	RedisDB                     = 0
	RedisDBPrefixHistoryPublic  = "historyPublic"
	RedisDBPrefixHistoryPrivate = "historyPrivate"
	RedisDBPrefixHistoryUnread  = "historyUnread"
)
