module mybilibili/msg-danmaku-service

go 1.26.5

require (
	github.com/lib/pq v1.12.3
	github.com/redis/go-redis/v9 v9.22.0
	google.golang.org/grpc v1.83.0
	mybilibili/pkg v0.0.0
)

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/golang-jwt/jwt/v5 v5.3.1 // indirect
	go.uber.org/atomic v1.11.0 // indirect
	golang.org/x/net v0.55.0 // indirect
	golang.org/x/sys v0.45.0 // indirect
	golang.org/x/text v0.37.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260526163538-3dc84a4a5aaa // indirect
	google.golang.org/protobuf v1.36.12 // indirect
)

replace mybilibili/pkg => ../pkg
