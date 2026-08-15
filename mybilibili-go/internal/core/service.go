package core

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"mybilibili/internal/abstraction"
	pb "mybilibili/internal/core/pb"
)

type Service struct {
	repo       *Repository
	jwt        *JWT
	cacheStore abstraction.CacheStore
}

func NewService(repo *Repository, jwtSecret string) *Service {
	return &Service{
		repo: repo,
		jwt:  NewJWT(jwtSecret),
	}
}

func (s *Service) SetCacheStore(cs abstraction.CacheStore) {
	s.cacheStore = cs
}

func (s *Service) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	if req.Username == "" || req.Password == "" {
		return nil, ErrInvalidArgument("username and password required")
	}

	hash := sha256.Sum256([]byte(req.Password))
	user := &User{
		Username: req.Username,
		Password: fmt.Sprintf("%x", hash),
		Nickname: req.Nickname,
		Email:    req.Email,
	}
	if user.Nickname == "" {
		user.Nickname = user.Username
	}

	id, err := s.repo.Create(ctx, user)
	if err != nil {
		return nil, ErrAlreadyExists("username already exists")
	}

	token, err := s.jwt.Generate(id)
	if err != nil {
		return nil, ErrInternal("token generation failed")
	}

	return &pb.RegisterResponse{UserId: id, Token: token}, nil
}

func (s *Service) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	if req.Username == "" || req.Password == "" {
		return nil, ErrInvalidArgument("username and password required")
	}

	user, err := s.repo.FindByUsername(ctx, req.Username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			_ = s.recordLoginFailure(ctx, 0)
			return nil, ErrNotFound("user not found")
		}
		return nil, ErrInternal("database error")
	}

	if s.cacheStore != nil {
		lockKey := fmt.Sprintf("login:fail:%d:lock", user.ID)
		locked, _ := s.cacheStore.Exists(ctx, lockKey)
		if locked {
			return nil, ErrUnauthenticated("account locked due to too many login attempts, try again later")
		}
	}

	hash := sha256.Sum256([]byte(req.Password))
	if fmt.Sprintf("%x", hash) != user.Password {
		_ = s.recordLoginFailure(ctx, user.ID)
		return nil, ErrUnauthenticated("invalid password")
	}

	if s.cacheStore != nil {
		s.cacheStore.Delete(ctx, fmt.Sprintf("login:fail:%d", user.ID))
		s.cacheStore.Delete(ctx, fmt.Sprintf("login:fail:%d:lock", user.ID))
	}

	token, err := s.jwt.Generate(user.ID)
	if err != nil {
		return nil, ErrInternal("token generation failed")
	}

	return &pb.LoginResponse{Token: token, UserId: user.ID, Nickname: user.Nickname}, nil
}

func (s *Service) recordLoginFailure(ctx context.Context, userID int64) error {
	if s.cacheStore == nil {
		return nil
	}
	key := fmt.Sprintf("login:fail:%d", userID)
	n, err := s.cacheStore.Incr(ctx, key, 15*time.Minute)
	if err != nil {
		return err
	}
	if n >= 5 && userID > 0 {
		lockKey := fmt.Sprintf("login:fail:%d:lock", userID)
		s.cacheStore.Set(ctx, lockKey, []byte("locked"), 15*time.Minute)
	}
	return nil
}

func (s *Service) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	user, err := s.repo.FindByID(ctx, req.UserId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound("user not found")
		}
		return nil, ErrInternal("database error")
	}

	return s.repo.ToPB(user), nil
}

type JWTClaims struct {
	UserId int64 `json:"user_id"`
	jwt.RegisteredClaims
}

type JWT struct {
	secret   string
	duration time.Duration
}

func NewJWT(secret string) *JWT {
	return &JWT{secret: secret, duration: 24 * time.Hour}
}

func (j *JWT) Generate(userID int64) (string, error) {
	claims := JWTClaims{
		UserId: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.duration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.secret))
}

func (j *JWT) GenerateRefresh(userID int64) (string, error) {
	claims := JWTClaims{
		UserId: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.secret))
}

func (j *JWT) Parse(tokenStr string) (int64, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &JWTClaims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(j.secret), nil
	})
	if err != nil {
		return 0, ErrUnauthenticated("invalid or expired token")
	}
	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return 0, ErrUnauthenticated("invalid token")
	}
	return claims.UserId, nil
}
