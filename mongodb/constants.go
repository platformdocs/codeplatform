package mongodb

import (
	"time"
)

const (
	HEART_BEAT_INTERVAL       = time.Second * 10
	CONNECT_TIMEOUT           = time.Second * 10
	MAX_CONNIDLE_TIME         = time.Second * 60 * 5
	QUERY_TIME_OUT            = time.Second * 5
	STATIC_COUNT_REDIS_EXPIRE = time.Minute * 60
)

const (
	DB_DATABASES  = "codeplatform"
	DB_COLLECTION = "codeList"
)
