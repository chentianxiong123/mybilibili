package subtitle

import (
	"context"
	"encoding/json"
	"time"

	"mybilibili/pkg/abstraction"
)

type Subtitle struct {
	ID           string    `json:"id"`
	VideoID      int64     `json:"video_id"`
	Language     string    `json:"language"`
	LanguageName string    `json:"language_name"`
	Format       string    `json:"format"`
	Content      string    `json:"content"`
	IsDefault    bool      `json:"is_default"`
	UploadedBy   int64     `json:"uploaded_by"`
	Status       int32     `json:"status"`
	Source       string    `json:"source"`
	UploadTime   time.Time `json:"upload_time"`
}

const collection = "subtitles"

type Repository struct {
	store abstraction.DocumentStore
}

func NewRepository(store abstraction.DocumentStore) *Repository {
	return &Repository{store: store}
}

func (r *Repository) Create(ctx context.Context, s *Subtitle) (string, error) {
	if s.Format == "" {
		s.Format = "srt"
	}
	if s.Status == 0 {
		s.Status = 0
	}
	if s.UploadTime.IsZero() {
		s.UploadTime = time.Now()
	}
	return r.store.Insert(ctx, collection, s)
}

func (r *Repository) GetByID(ctx context.Context, id string) (*Subtitle, error) {
	s := &Subtitle{}
	if err := r.store.FindByID(ctx, collection, id, s); err != nil {
		return nil, err
	}
	return s, nil
}

func (r *Repository) ListByVideo(ctx context.Context, videoID int64) ([]*Subtitle, error) {
	var result []*Subtitle
	filter := abstraction.QueryFilter{
		PageSize: 100,
		Filters:  map[string]any{"video_id": videoID},
	}
	if err := r.store.Query(ctx, collection, filter, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (r *Repository) GetByLanguage(ctx context.Context, videoID int64, language string) (*Subtitle, error) {
	list, err := r.ListByVideo(ctx, videoID)
	if err != nil {
		return nil, err
	}
	for _, s := range list {
		if s.Language == language {
			return s, nil
		}
	}
	return nil, nil
}

func (r *Repository) ListPending(ctx context.Context) ([]*Subtitle, error) {
	var result []*Subtitle
	filter := abstraction.QueryFilter{
		PageSize: 100,
		Filters:  map[string]any{"status": 0},
	}
	if err := r.store.Query(ctx, collection, filter, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (r *Repository) ListAll(ctx context.Context) ([]*Subtitle, error) {
	var result []*Subtitle
	filter := abstraction.QueryFilter{PageSize: 100}
	if err := r.store.Query(ctx, collection, filter, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (r *Repository) Update(ctx context.Context, id string, s *Subtitle) error {
	return r.store.Update(ctx, collection, id, s)
}

func (r *Repository) Delete(ctx context.Context, id string) error {
	return r.store.Delete(ctx, collection, id)
}

func (r *Repository) SetDefault(ctx context.Context, videoID int64, subtitleID string) error {
	list, err := r.ListByVideo(ctx, videoID)
	if err != nil {
		return err
	}
	for _, s := range list {
		s.IsDefault = s.ID == subtitleID
		if err := r.store.Update(ctx, collection, s.ID, s); err != nil {
			return err
		}
	}
	return nil
}

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Upload(ctx context.Context, videoID, uploadedBy int64, language, languageName, content string) (*Subtitle, error) {
	st := &Subtitle{
		VideoID: videoID, Language: language, LanguageName: languageName,
		Content: content, Format: "srt", UploadedBy: uploadedBy,
		Status: 0, Source: "user",
	}
	id, err := s.repo.Create(ctx, st)
	if err != nil {
		return nil, err
	}
	st.ID = id
	return st, nil
}

func (s *Service) ListByVideo(ctx context.Context, videoID int64) ([]*Subtitle, error) {
	return s.repo.ListByVideo(ctx, videoID)
}

func (s *Service) ListAll(ctx context.Context) ([]*Subtitle, error) {
	return s.repo.ListAll(ctx)
}

func (s *Service) ListByVideoForScan(ctx context.Context, videoID int64) ([]*Subtitle, error) {
	return s.repo.ListByVideo(ctx, videoID)
}

func (s *Service) GetByLanguage(ctx context.Context, videoID int64, language string) (*Subtitle, error) {
	return s.repo.GetByLanguage(ctx, videoID, language)
}

func (s *Service) ListPending(ctx context.Context) ([]*Subtitle, error) {
	return s.repo.ListPending(ctx)
}

func (s *Service) Approve(ctx context.Context, id string) error {
	st, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	st.Status = 1
	return s.repo.Update(ctx, id, st)
}

func (s *Service) Reject(ctx context.Context, id string) error {
	st, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	st.Status = 2
	return s.repo.Update(ctx, id, st)
}

func (s *Service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

func (s *Service) SetDefault(ctx context.Context, videoID int64, subtitleID string) error {
	return s.repo.SetDefault(ctx, videoID, subtitleID)
}

func (s *Service) Preview(ctx context.Context, id string) ([]map[string]interface{}, error) {
	st, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	var cues []map[string]interface{}
	if json.Unmarshal([]byte(st.Content), &cues) != nil {
		if parsed, err := ParseSRT(st.Content); err == nil {
			cues = make([]map[string]interface{}, 0, len(parsed))
			for _, c := range parsed {
				cues = append(cues, c.ToCueMap())
			}
		}
	}
	return cues, nil
}
