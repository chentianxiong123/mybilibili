module mybilibili/core-service

go 1.26.5

require (
	github.com/golang-jwt/jwt/v5 v5.3.1
	github.com/lib/pq v1.12.3
	github.com/redis/go-redis/v9 v9.22.0
	google.golang.org/grpc v1.83.0
	google.golang.org/protobuf v1.36.11
	mybilibili/pkg v0.0.0
)

require golang.org/x/net v0.55.0 // indirect

replace mybilibili/pkg => ../pkg
