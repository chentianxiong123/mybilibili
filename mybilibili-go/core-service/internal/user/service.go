package user

import (
	"context"
	"crypto/sha256"
	"database/sql"
	stderrors "errors"
	"fmt"
	"time"

	"mybilibili/pkg/abstraction"
	"mybilibili/pkg/auth"
	"mybilibili/pkg/errors"
	pb "mybilibili/pkg/pb"
)

type Service struct {
	repo       *Repository
	jwt        *auth.JWT
	cacheStore abstraction.CacheStore
}

func NewService(repo *Repository, jwtSecret string) *Service {
	return &Service{
		repo: repo,
		jwt:  auth.NewJWT(jwtSecret),
	}
}

func (s *Service) SetCacheStore(cs abstraction.CacheStore) {
	s.cacheStore = cs
}

func (s *Service) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	if req.Username == "" || req.Password == "" {
		return nil, errors.ErrInvalidArgument("username and password required")
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

	if _, err := s.repo.FindByNickname(ctx, user.Nickname); err == nil {
		return nil, errors.ErrAlreadyExists("nickname already exists")
	}

	id, err := s.repo.Create(ctx, user)
	if err != nil {
		return nil, errors.ErrAlreadyExists("username already exists")
	}

	token, err := s.jwt.Generate(id)
	if err != nil {
		return nil, errors.ErrInternal("token generation failed")
	}

	return &pb.RegisterResponse{UserId: id, Token: token}, nil
}

func (s *Service) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	if req.Username == "" || req.Password == "" {
		return nil, errors.ErrInvalidArgument("username and password required")
	}

	user, err := s.repo.FindByUsername(ctx, req.Username)
	if err != nil {
		if stderrors.Is(err, sql.ErrNoRows) {
			_ = s.recordLoginFailure(ctx, 0)
			return nil, errors.ErrNotFound("user not found")
		}
		return nil, errors.ErrInternal("database error")
	}

	if user.Status != 1 {
		return nil, errors.ErrUnauthenticated("account is disabled")
	}

	if s.cacheStore != nil {
		lockKey := fmt.Sprintf("login:fail:%d:lock", user.ID)
		locked, _ := s.cacheStore.Exists(ctx, lockKey)
		if locked {
			return nil, errors.ErrUnauthenticated("account locked due to too many login attempts, try again later")
		}
	}

	hash := sha256.Sum256([]byte(req.Password))
	if fmt.Sprintf("%x", hash) != user.Password {
		_ = s.recordLoginFailure(ctx, user.ID)
		return nil, errors.ErrUnauthenticated("invalid password")
	}

	if s.cacheStore != nil {
		s.cacheStore.Delete(ctx, fmt.Sprintf("login:fail:%d", user.ID))
		s.cacheStore.Delete(ctx, fmt.Sprintf("login:fail:%d:lock", user.ID))
	}

	token, err := s.jwt.Generate(user.ID)
	if err != nil {
		return nil, errors.ErrInternal("token generation failed")
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
		if stderrors.Is(err, sql.ErrNoRows) {
			return nil, errors.ErrNotFound("user not found")
		}
		return nil, errors.ErrInternal("database error")
	}

	return s.repo.ToPB(user), nil
}
