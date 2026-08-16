module mybilibili/msg-danmaku-service

go 1.26.5

require (
	github.com/lib/pq v1.12.3
	google.golang.org/grpc v1.83.0
	google.golang.org/protobuf v1.36.12
	mybilibili/pkg v0.0.0
)

replace mybilibili/pkg => ../pkg