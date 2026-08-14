package core

import (
	"context"
	"database/sql"
	"errors"

	pb "mybilibili/internal/core/pb"
)

type ManuscriptService struct {
	repo     *ManuscriptRepository
	userRepo *Repository
}

func NewManuscriptService(repo *ManuscriptRepository, userRepo *Repository) *ManuscriptService {
	return &ManuscriptService{repo: repo, userRepo: userRepo}
}

func (s *ManuscriptService) GetManuscript(ctx context.Context, req *pb.GetManuscriptRequest) (*pb.GetManuscriptResponse, error) {
	m, err := s.repo.FindByID(ctx, req.Id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound("manuscript not found")
		}
		return nil, ErrInternal("database error")
	}
	return &pb.GetManuscriptResponse{Manuscript: s.buildManuscriptInfo(ctx, m, 0, false)}, nil
}

func (s *ManuscriptService) GetManuscriptWithVideos(ctx context.Context, req *pb.GetManuscriptWithVideosRequest) (*pb.GetManuscriptResponse, error) {
	m, err := s.repo.FindByID(ctx, req.Id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound("manuscript not found")
		}
		return nil, ErrInternal("database error")
	}
	s.repo.IncrementViewCount(ctx, req.Id)
	s.repo.UpsertDailyMetric(ctx, req.Id, m.UserID, "view_count", 1)
	return &pb.GetManuscriptResponse{Manuscript: s.buildManuscriptInfo(ctx, m, req.CurrentUserId, true)}, nil
}

func (s *ManuscriptService) ListUserManuscripts(ctx context.Context, req *pb.ListUserManuscriptsRequest) (*pb.ListUserManuscriptsResponse, error) {
	list, total, err := s.repo.ListByUser(ctx, req.UserId, req.Status, req.Page, req.PageSize)
	if err != nil {
		return nil, ErrInternal("database error")
	}

	var infos []*pb.ManuscriptInfo
	for _, m := range list {
		infos = append(infos, s.buildManuscriptInfo(ctx, m, 0, false))
	}

	return &pb.ListUserManuscriptsResponse{
		Manuscripts: infos,
		Total:       total,
		Page:        req.Page,
		PageSize:    req.PageSize,
	}, nil
}

func (s *ManuscriptService) ListRecommended(ctx context.Context, req *pb.ListRecommendedRequest) (*pb.ListRecommendedResponse, error) {
	list, err := s.repo.ListRecommended(ctx)
	if err != nil {
		return nil, ErrInternal("database error")
	}

	var infos []*pb.ManuscriptInfo
	for _, m := range list {
		infos = append(infos, s.buildManuscriptInfo(ctx, m, req.UserId, false))
	}
	return &pb.ListRecommendedResponse{Manuscripts: infos}, nil
}

func (s *ManuscriptService) ListHot(ctx context.Context, req *pb.ListHotRequest) (*pb.ListHotResponse, error) {
	list, err := s.repo.ListHot(ctx)
	if err != nil {
		return nil, ErrInternal("database error")
	}

	var infos []*pb.ManuscriptInfo
	for _, m := range list {
		infos = append(infos, s.buildManuscriptInfo(ctx, m, req.UserId, false))
	}
	return &pb.ListHotResponse{Manuscripts: infos}, nil
}

func (s *ManuscriptService) ListByCategory(ctx context.Context, req *pb.ListByCategoryRequest) (*pb.ListByCategoryResponse, error) {
	list, total, err := s.repo.ListByCategory(ctx, req.CategoryId, req.Page, req.PageSize)
	if err != nil {
		return nil, ErrInternal("database error")
	}

	var infos []*pb.ManuscriptInfo
	for _, m := range list {
		infos = append(infos, s.buildManuscriptInfo(ctx, m, 0, false))
	}
	return &pb.ListByCategoryResponse{Manuscripts: infos, Total: total}, nil
}

func (s *ManuscriptService) ListCategories(ctx context.Context, req *pb.ListCategoriesRequest) (*pb.ListCategoriesResponse, error) {
	cats, err := s.repo.ListCategories(ctx)
	if err != nil {
		return nil, ErrInternal("database error")
	}

	var infos []*pb.CategoryInfo
	for _, c := range cats {
		infos = append(infos, &pb.CategoryInfo{
			Id: c.ID, Name: c.Name, Icon: c.Icon, SortOrder: c.SortOrder,
		})
	}
	return &pb.ListCategoriesResponse{Categories: infos}, nil
}

func (s *ManuscriptService) DeleteManuscript(ctx context.Context, req *pb.DeleteManuscriptRequest) (*pb.DeleteManuscriptResponse, error) {
	if err := s.repo.Delete(ctx, req.Id, req.UserId); err != nil {
		return nil, ErrPermissionDenied("cannot delete")
	}
	return &pb.DeleteManuscriptResponse{}, nil
}

func (s *ManuscriptService) PublishManuscript(ctx context.Context, req *pb.PublishManuscriptRequest) (*pb.PublishManuscriptResponse, error) {
	if err := s.repo.UpdateStatus(ctx, req.Id, req.UserId, 3); err != nil {
		return nil, ErrPermissionDenied("cannot publish")
	}
	return &pb.PublishManuscriptResponse{}, nil
}

func (s *ManuscriptService) UnpublishManuscript(ctx context.Context, req *pb.UnpublishManuscriptRequest) (*pb.UnpublishManuscriptResponse, error) {
	if err := s.repo.UpdateStatus(ctx, req.Id, req.UserId, -1); err != nil {
		return nil, ErrPermissionDenied("cannot unpublish")
	}
	return &pb.UnpublishManuscriptResponse{}, nil
}

func (s *ManuscriptService) SearchUserManuscripts(ctx context.Context, req *pb.SearchUserManuscriptsRequest) (*pb.SearchUserManuscriptsResponse, error) {
	list, err := s.repo.SearchUser(ctx, req.UserId, req.Keyword, req.Sort)
	if err != nil {
		return nil, ErrInternal("database error")
	}

	var infos []*pb.ManuscriptInfo
	for _, m := range list {
		infos = append(infos, s.buildManuscriptInfo(ctx, m, 0, false))
	}
	return &pb.SearchUserManuscriptsResponse{Manuscripts: infos}, nil
}

func (s *ManuscriptService) buildManuscriptInfo(ctx context.Context, m *Manuscript, currentUserID int64, loadVideos bool) *pb.ManuscriptInfo {
	catName := ""
	cat, err := s.repo.FindCategoryByID(ctx, m.CategoryID)
	if err == nil {
		catName = cat.Name
	}

	var uploader *pb.UserInfo
	user, err := s.userRepo.FindByID(ctx, m.UserID)
	if err == nil {
		uploader = &pb.UserInfo{
			Id:     user.ID,
			Name:   user.Nickname,
			Avatar: user.Avatar,
			Level:  user.Level,
		}
	}

	var videoPBs []*pb.VideoItem
	var firstVideoID int64
	var firstVideoPlayURL string

	if loadVideos {
		videos, _ := s.repo.FindVideosByManuscriptID(ctx, m.ID)
		for _, v := range videos {
			videoPBs = append(videoPBs, videoToPB(v))
		}
		if len(videos) > 0 {
			firstVideoID = videos[0].ID
			firstVideoPlayURL = videos[0].PlayURLHd
		}
	}

	var tags []string
	if loadVideos && len(videoPBs) > 0 {
		for _, videoPB := range videoPBs {
			videoTags, _ := s.repo.FindTagsByVideoID(ctx, videoPB.Id)
			tags = append(tags, videoTags...)
		}
		tags = unique(tags)
	}

	return s.repo.ToPB(m, catName, uploader, videoPBs, tags, firstVideoID, firstVideoPlayURL)
}

func unique(s []string) []string {
	seen := make(map[string]bool)
	var r []string
	for _, v := range s {
		if !seen[v] {
			seen[v] = true
			r = append(r, v)
		}
	}
	return r
}
