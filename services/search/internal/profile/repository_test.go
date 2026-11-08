package profile

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"mybilibili/pkg/abstraction"
)

type fakeDocumentStore struct {
	profiles map[int64]*Profile
	lastID   int64

	// Optional error overrides for testing error paths.
	queryErr    error
	updateErr   error
	insertErr   error
	queryFail   bool
	updateFail  bool
	insertFail  bool

	// Observations
	queryCalls  int
	updateCalls int
	insertCalls int
}

func newFakeDocumentStore() *fakeDocumentStore {
	return &fakeDocumentStore{
		profiles: make(map[int64]*Profile),
		lastID:   0,
	}
}

func (f *fakeDocumentStore) allocateID() string {
	f.lastID++
	return "profile-" + itoa(f.lastID)
}

func itoa(v int64) string {
	if v == 0 {
		return "0"
	}
	neg := v < 0
	if neg {
		v = -v
	}
	var buf [20]byte
	i := len(buf)
	for v > 0 {
		i--
		buf[i] = byte('0' + v%10)
		v /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}

func (f *fakeDocumentStore) findProfile(p *Profile) *Profile {
	for _, prof := range f.profiles {
		if prof.ID == p.ID && prof.UserID == p.UserID {
			return prof
		}
	}
	return nil
}

func (f *fakeDocumentStore) Insert(ctx context.Context, collection string, doc any) (string, error) {
	f.insertCalls++
	if f.insertFail {
		return "", f.insertErr
	}
	p, ok := doc.(*Profile)
	if !ok {
		return "", errors.New("unexpected doc type")
	}
	id := f.allocateID()
	clone := *p
	clone.ID = id
	f.profiles[p.UserID] = &clone
	return id, nil
}

func (f *fakeDocumentStore) FindByID(ctx context.Context, collection, id string, result any) error {
	// Not used by profile package.
	return nil
}

func (f *fakeDocumentStore) Update(ctx context.Context, collection, id string, doc any) error {
	f.updateCalls++
	if f.updateFail {
		return f.updateErr
	}
	p, ok := doc.(*Profile)
	if !ok {
		return errors.New("unexpected doc type")
	}
	p.ID = id
	f.profiles[p.UserID] = p
	return nil
}

func (f *fakeDocumentStore) Delete(ctx context.Context, collection, id string) error {
	return nil
}

func (f *fakeDocumentStore) Query(ctx context.Context, collection string, filter abstraction.QueryFilter, result any) error {
	f.queryCalls++
	if f.queryFail {
		return f.queryErr
	}
	list, ok := result.(*[]*Profile)
	if !ok {
		return errors.New("unexpected result type")
	}
	userID, ok := filter.Filters["user_id"].(int64)
	if !ok {
		*list = nil
		return nil
	}
	if p, found := f.profiles[userID]; found {
		*list = []*Profile{p}
	} else {
		*list = nil
	}
	return nil
}

var _ abstraction.DocumentStore = (*fakeDocumentStore)(nil)

func TestNewRepository_WrapsStore(t *testing.T) {
	store := newFakeDocumentStore()
	repo := NewRepository(store)
	require.NotNil(t, repo)
	assert.Equal(t, store, repo.store)
}

func TestRepository_GetByUser_NotFound(t *testing.T) {
	store := newFakeDocumentStore()
	repo := NewRepository(store)
	p, err := repo.GetByUser(context.Background(), 999)
	require.NoError(t, err)
	assert.Nil(t, p)
	assert.Equal(t, 1, store.queryCalls)
}

func TestRepository_GetByUser_Found(t *testing.T) {
	store := newFakeDocumentStore()
	store.profiles[7] = &Profile{ID: "profile-1", UserID: 7, Tags: []string{"a"}}
	repo := NewRepository(store)

	p, err := repo.GetByUser(context.Background(), 7)
	require.NoError(t, err)
	require.NotNil(t, p)
	assert.Equal(t, "profile-1", p.ID)
	assert.Equal(t, []string{"a"}, p.Tags)
}

func TestRepository_GetByUser_QueryError(t *testing.T) {
	store := newFakeDocumentStore()
	store.queryFail = true
	store.queryErr = errors.New("db unavailable")
	repo := NewRepository(store)

	p, err := repo.GetByUser(context.Background(), 1)
	require.Error(t, err)
	assert.Nil(t, p)
	assert.ErrorContains(t, err, "db unavailable")
}

func TestRepository_Upsert_InsertNew(t *testing.T) {
	store := newFakeDocumentStore()
	repo := NewRepository(store)
	p := &Profile{UserID: 42, Tags: []string{"tech"}}

	require.NoError(t, repo.Upsert(context.Background(), p))
	assert.NotEmpty(t, p.ID)
	require.Equal(t, 1, store.insertCalls)
	require.Equal(t, 0, store.updateCalls)
	assert.Equal(t, p, store.profiles[42])
}

func TestRepository_Upsert_UpdateExisting(t *testing.T) {
	store := newFakeDocumentStore()
	store.profiles[5] = &Profile{ID: "existing-id", UserID: 5}
	repo := NewRepository(store)
	p := &Profile{UserID: 5, Tags: []string{"updated"}}

	require.NoError(t, repo.Upsert(context.Background(), p))
	assert.Equal(t, "existing-id", p.ID)
	assert.Equal(t, 1, store.updateCalls)
	assert.Equal(t, 0, store.insertCalls)
	assert.Equal(t, p, store.profiles[5])
}

func TestRepository_Upsert_InsertError(t *testing.T) {
	store := newFakeDocumentStore()
	store.insertFail = true
	store.insertErr = errors.New("insert failed")
	repo := NewRepository(store)
	p := &Profile{UserID: 42}

	err := repo.Upsert(context.Background(), p)
	require.Error(t, err)
	assert.ErrorContains(t, err, "insert failed")
	assert.Empty(t, p.ID)
}

func TestRepository_Upsert_UpdateError(t *testing.T) {
	store := newFakeDocumentStore()
	store.profiles[5] = &Profile{ID: "existing-id", UserID: 5}
	store.updateFail = true
	store.updateErr = errors.New("update failed")
	repo := NewRepository(store)
	p := &Profile{UserID: 5}

	err := repo.Upsert(context.Background(), p)
	require.Error(t, err)
	assert.ErrorContains(t, err, "update failed")
	assert.Equal(t, "existing-id", p.ID)
}

func TestRepository_Upsert_SwallowedGetByUserError(t *testing.T) {
	// GetByUser errors are intentionally ignored in Upsert (existing, _ := ...).
	store := newFakeDocumentStore()
	store.queryFail = true
	store.queryErr = errors.New("query failed")
	repo := NewRepository(store)
	p := &Profile{UserID: 1}

	require.NoError(t, repo.Upsert(context.Background(), p))
	assert.NotEmpty(t, p.ID)
}
