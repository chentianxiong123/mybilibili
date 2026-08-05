package core

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"
	"time"

	pb "mybilibili/internal/core/pb"
)

type Service struct {
	repo *Repository
	jwt  *JWT
}

func NewService(repo *Repository, jwtSecret string) *Service {
	return &Service{
		repo: repo,
		jwt:  NewJWT(jwtSecret),
	}
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
			return nil, ErrNotFound("user not found")
		}
		return nil, ErrInternal("database error")
	}

	hash := sha256.Sum256([]byte(req.Password))
	if fmt.Sprintf("%x", hash) != user.Password {
		return nil, ErrUnauthenticated("invalid password")
	}

	token, err := s.jwt.Generate(user.ID)
	if err != nil {
		return nil, ErrInternal("token generation failed")
	}

	return &pb.LoginResponse{Token: token, UserId: user.ID, Nickname: user.Nickname}, nil
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

type JWT struct {
	secret    string
	duration  time.Duration
}

func NewJWT(secret string) *JWT {
	return &JWT{secret: secret, duration: 24 * time.Hour}
}

func (j *JWT) Generate(userID int64) (string, error) {
	payload := fmt.Sprintf("%d:%d:%s", userID, time.Now().Add(j.duration).Unix(), j.secret)
	hash := sha256.Sum256([]byte(payload))
	token := fmt.Sprintf("%x", hash)
	return token, nil
}