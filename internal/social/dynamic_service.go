package social

import (
	"context"
	"database/sql"
)

type DynamicService struct {
	repo *DynamicRepository
}

func NewDynamicService(repo *DynamicRepository) *DynamicService {
	return &DynamicService{repo: repo}
}

func (s *DynamicService) Publish(ctx context.Context, userID int64, content string, dynamicType int32, imageURL string, refManuscriptID int64) (*Dynamic, error) {
	d := &Dynamic{UserID: userID, Content: content, DynamicType: dynamicType, ImageURL: imageURL, RefManuscriptID: refManuscriptID}
	id, err := s.repo.Create(ctx, d)
	if err != nil {
		return nil, err
	}
	return s.repo.GetByID(ctx, id)
}

func (s *DynamicService) GetByID(ctx context.Context, id int64) (*Dynamic, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *DynamicService) ListByUser(ctx context.Context, userID int64, page, limit int32) ([]*Dynamic, error) {
	return s.repo.ListByUser(ctx, userID, page, limit)
}

func (s *DynamicService) ListFollowing(ctx context.Context, userID int64, page, limit int32) ([]*Dynamic, error) {
	return s.repo.ListFollowing(ctx, userID, page, limit)
}

func (s *DynamicService) Delete(ctx context.Context, id, userID int64) error {
	return s.repo.Delete(ctx, id, userID)
}

func (s *DynamicService) Like(ctx context.Context, id, userID int64) error {
	return s.repo.IncrLikeCount(ctx, id, 1)
}

func (s *DynamicService) Unlike(ctx context.Context, id, userID int64) error {
	return s.repo.IncrLikeCount(ctx, id, -1)
}

func (s *DynamicService) AddComment(ctx context.Context, dynamicID, userID int64, content string, parentID, replyUserID int64) (*DynamicComment, error) {
	dc := &DynamicComment{DynamicID: dynamicID, UserID: userID, Content: content, ParentID: parentID, ReplyUserID: replyUserID}
	_, err := s.repo.CreateComment(ctx, dc)
	if err != nil {
		return nil, err
	}
	s.repo.IncrCommentCount(ctx, dynamicID, 1)
	return dc, nil
}

func (s *DynamicService) ListComments(ctx context.Context, dynamicID int64, page, limit int32) ([]*DynamicComment, error) {
	return s.repo.ListComments(ctx, dynamicID, page, limit)
}

type CollectionRepository struct {
	db *sql.DB
}

func NewCollectionRepository(db *sql.DB) *CollectionRepository {
	return &CollectionRepository{db: db}
}

func (r *CollectionRepository) Create(ctx context.Context, userID int64, title, desc, cover string, status int32) (int64, error) {
	var id int64
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO manuscript_collections (title, description, cover_url, user_id, status) VALUES ($1,$2,$3,$4,$5) RETURNING id`,
		title, desc, cover, userID, status).Scan(&id)
	return id, err
}

func (r *CollectionRepository) GetByID(ctx context.Context, id int64) (map[string]interface{}, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, title, description, cover_url, user_id, manuscript_count, view_count, status, created_at, updated_at
		 FROM manuscript_collections WHERE id = $1`, id)
	return scanCollection(row)
}

func (r *CollectionRepository) ListByUser(ctx context.Context, userID int64) ([]map[string]interface{}, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, title, description, cover_url, user_id, manuscript_count, view_count, status, created_at, updated_at
		 FROM manuscript_collections WHERE user_id = $1 ORDER BY updated_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanCollections(rows)
}

func (r *CollectionRepository) Update(ctx context.Context, id, userID int64, title, desc string, status int32) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE manuscript_collections SET title=$1, description=$2, status=$3, updated_at=NOW() WHERE id=$4 AND user_id=$5`,
		title, desc, status, id, userID)
	return err
}

func (r *CollectionRepository) Delete(ctx context.Context, id, userID int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM manuscript_collections WHERE id=$1 AND user_id=$2`, id, userID)
	return err
}

func (r *CollectionRepository) AddManuscript(ctx context.Context, collectionID, manuscriptID int64, order int) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO manuscript_collection_relations (manuscript_id, collection_id, collection_order) VALUES ($1,$2,$3) ON CONFLICT DO NOTHING`,
		manuscriptID, collectionID, order)
	if err != nil {
		return err
	}
	_, err = r.db.ExecContext(ctx,
		`UPDATE manuscript_collections SET manuscript_count = (SELECT COUNT(*) FROM manuscript_collection_relations WHERE collection_id = $1) WHERE id = $1`,
		collectionID)
	return err
}

func (r *CollectionRepository) RemoveManuscript(ctx context.Context, collectionID, manuscriptID int64) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM manuscript_collection_relations WHERE collection_id=$1 AND manuscript_id=$2`,
		collectionID, manuscriptID)
	if err != nil {
		return err
	}
	_, err = r.db.ExecContext(ctx,
		`UPDATE manuscript_collections SET manuscript_count = (SELECT COUNT(*) FROM manuscript_collection_relations WHERE collection_id = $1) WHERE id = $1`,
		collectionID)
	return err
}

func (r *CollectionRepository) ListManuscripts(ctx context.Context, collectionID int64, page, limit int32) ([]int64, error) {
	offset := (page - 1) * limit
	rows, err := r.db.QueryContext(ctx,
		`SELECT manuscript_id FROM manuscript_collection_relations WHERE collection_id = $1 ORDER BY collection_order LIMIT $2 OFFSET $3`,
		collectionID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []int64
	for rows.Next() {
		var id int64
		rows.Scan(&id)
		ids = append(ids, id)
	}
	return ids, nil
}

func scanCollection(row interface {
	Scan(dest ...interface{}) error
}) (map[string]interface{}, error) {
	var id, userID, manuscriptCount, viewCount, status int64
	var title, desc, cover, createdAt, updatedAt string
	err := row.Scan(&id, &title, &desc, &cover, &userID, &manuscriptCount, &viewCount, &status, &createdAt, &updatedAt)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"id": id, "title": title, "description": desc, "cover_url": cover,
		"user_id": userID, "manuscript_count": manuscriptCount, "view_count": viewCount,
		"status": status, "created_at": createdAt, "updated_at": updatedAt,
	}, nil
}

func scanCollections(rows *sql.Rows) ([]map[string]interface{}, error) {
	var list []map[string]interface{}
	for rows.Next() {
		m, err := scanCollection(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, m)
	}
	return list, nil
}

type CollectionService struct {
	repo *CollectionRepository
}

func NewCollectionService(repo *CollectionRepository) *CollectionService {
	return &CollectionService{repo: repo}
}

func (s *CollectionService) Create(ctx context.Context, userID int64, title, desc, cover string, status int32) (map[string]interface{}, error) {
	id, err := s.repo.Create(ctx, userID, title, desc, cover, status)
	if err != nil {
		return nil, err
	}
	return s.repo.GetByID(ctx, id)
}

func (s *CollectionService) GetByID(ctx context.Context, id int64) (map[string]interface{}, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *CollectionService) ListByUser(ctx context.Context, userID int64) ([]map[string]interface{}, error) {
	return s.repo.ListByUser(ctx, userID)
}

func (s *CollectionService) Update(ctx context.Context, id, userID int64, title, desc string, status int32) error {
	return s.repo.Update(ctx, id, userID, title, desc, status)
}

func (s *CollectionService) Delete(ctx context.Context, id, userID int64) error {
	return s.repo.Delete(ctx, id, userID)
}

func (s *CollectionService) AddManuscript(ctx context.Context, collectionID, manuscriptID, userID int64) error {
	return s.repo.AddManuscript(ctx, collectionID, manuscriptID, 0)
}

func (s *CollectionService) RemoveManuscript(ctx context.Context, collectionID, manuscriptID, userID int64) error {
	return s.repo.RemoveManuscript(ctx, collectionID, manuscriptID)
}

func (s *CollectionService) ListManuscripts(ctx context.Context, collectionID int64, page, limit int32) ([]int64, error) {
	return s.repo.ListManuscripts(ctx, collectionID, page, limit)
}

type ShareRepository struct {
	db *sql.DB
}

func NewShareRepository(db *sql.DB) *ShareRepository {
	return &ShareRepository{db: db}
}

func (r *ShareRepository) Record(ctx context.Context, userID, manuscriptID int64, channel, ip string) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO shares (user_id, manuscript_id, channel, ip_address) VALUES ($1,$2,$3,$4)`,
		userID, manuscriptID, channel, ip)
	if err != nil {
		return err
	}
	_, err = r.db.ExecContext(ctx,
		`UPDATE manuscripts SET share_count = share_count + 1 WHERE id = $1`, manuscriptID)
	return err
}
