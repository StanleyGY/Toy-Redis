package cmdexec

type RedisDb struct {
	KvStore map[string]string
}

var db *RedisDb

func InitRedisDb() {
	db = &RedisDb{
		KvStore: make(map[string]string),
	}
}
