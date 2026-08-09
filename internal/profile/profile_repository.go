package profile

import (
	"context"
	"time"

	"mybilibili/internal/abstraction"
)

type Profile struct {
	ID                 string    `json:"id"`
	UserID             int64     `json:"user_id"`
	Tags               []string  `json:"tags"`
	FavoriteCategories []int64   `json:"favorite_categories"`
	WatchCount         int64     `json:"watch_count"`
	LikeCount          int64     `json:"like_count"`
	CollectCount       int64     `json:"collect_count"`
	UpdatedAt          time.Time `json:"updated_at"`
}

const collection = "user_profiles"

type Repository struct {
	store abstraction.DocumentStore
}

func NewRepository(store abstraction.DocumentStore) *Repository {
	return &Repository{store: store}
}

func (r *Repository) GetByUser(ctx context.Context, userID int64) (*Profile, error) {
	var profiles []*Profile
	filter := abstraction.QueryFilter{
		PageSize: 1,
		Filters:  map[string]any{"user_id": userID},
	}
	if err := r.store.Query(ctx, collection, filter, &profiles); err != nil {
		return nil, err
	}
	if len(profiles) == 0 {
		return nil, nil
	}
	return profiles[0], nil
}

func (r *Repository) Upsert(ctx context.Context, p *Profile) error {
	existing, _ := r.GetByUser(ctx, p.UserID)
	if existing != nil {
		p.ID = existing.ID
		return r.store.Update(ctx, collection, p.ID, p)
	}
	id, err := r.store.Insert(ctx, collection, p)
	if err != nil {
		return err
	}
	p.ID = id
	return nil
}

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Get(ctx context.Context, userID int64) (*Profile, error) {
	return s.repo.GetByUser(ctx, userID)
}

func (s *Service) Init(ctx context.Context, userID int64, tags []string) (*Profile, error) {
	existing, _ := s.repo.GetByUser(ctx, userID)
	if existing != nil {
		existing.Tags = tags
		existing.UpdatedAt = time.Now()
		if err := s.repo.Upsert(ctx, existing); err != nil {
			return nil, err
		}
		return existing, nil
	}
	p := &Profile{UserID: userID, Tags: tags, UpdatedAt: time.Now()}
	if err := s.repo.Upsert(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *Service) RecordWatch(ctx context.Context, userID int64) error {
	p, _ := s.repo.GetByUser(ctx, userID)
	if p == nil {
		p = &Profile{UserID: userID, UpdatedAt: time.Now()}
	}
	p.WatchCount++
	p.UpdatedAt = time.Now()
	return s.repo.Upsert(ctx, p)
}

var _ = context.Background
