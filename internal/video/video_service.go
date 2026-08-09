package video

import (
	"context"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetVideo(ctx context.Context, id int64) (*Video, error) {
	return s.repo.GetVideoByID(ctx, id)
}

func (s *Service) ListByManuscript(ctx context.Context, manuscriptID int64) ([]*Video, error) {
	return s.repo.ListByManuscript(ctx, manuscriptID)
}

func (s *Service) ListCategories(ctx context.Context) ([]*Category, error) {
	return s.repo.ListCategories(ctx)
}

func (s *Service) CreateCategory(ctx context.Context, name, icon string, sortOrder int32) error {
	_, err := s.repo.CreateCategory(ctx, name, icon, sortOrder)
	return err
}

func (s *Service) UpdateCategory(ctx context.Context, id int64, name, icon string, sortOrder int32) error {
	return s.repo.UpdateCategory(ctx, id, name, icon, sortOrder)
}

func (s *Service) DeleteCategory(ctx context.Context, id int64) error {
	return s.repo.DeleteCategory(ctx, id)
}

func (s *Service) ListBanners(ctx context.Context, bannerType int32) ([]*BannerImage, error) {
	return s.repo.ListBanners(ctx, bannerType)
}

func (s *Service) ListBannersByCategory(ctx context.Context, bannerType int32, categoryID int64) ([]*BannerImage, error) {
	return s.repo.ListBannersByCategory(ctx, bannerType, categoryID)
}

func (s *Service) CreateBanner(ctx context.Context, b *BannerImage) error {
	_, err := s.repo.CreateBanner(ctx, b)
	return err
}

func (s *Service) UpdateBanner(ctx context.Context, id int64, b *BannerImage) error {
	return s.repo.UpdateBanner(ctx, id, b)
}

func (s *Service) DeleteBanner(ctx context.Context, id int64) error {
	return s.repo.DeleteBanner(ctx, id)
}

func (s *Service) Statistics(ctx context.Context) (map[string]interface{}, error) {
	return s.repo.GetStatistics(ctx)
}
