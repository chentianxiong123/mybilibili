package subtitle

// Pinning tests for confirmed subtitle repository / import defects.
// Each test asserts CURRENT (buggy) behavior; it fails once the bug is fixed.

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"mybilibili/pkg/abstraction"
)

// recordingStore records inserts/updates and serves a fixed in-memory list.
type recordingStore struct {
	mu      sync.Mutex
	subs    []*Subtitle
	inserts int
	updates []*Subtitle
	nextID  int
}

func (m *recordingStore) Insert(_ context.Context, _ string, doc any) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.inserts++
	m.nextID++
	if s, ok := doc.(*Subtitle); ok {
		s.ID = "rec-" + string(rune('0'+m.nextID))
		m.subs = append(m.subs, s)
	}
	return "rec-" + string(rune('0'+m.nextID)), nil
}

func (m *recordingStore) FindByID(_ context.Context, _ string, id string, result any) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, s := range m.subs {
		if s.ID == id {
			*(result.(*Subtitle)) = *s
			return nil
		}
	}
	return nil
}

func (m *recordingStore) Update(_ context.Context, _ string, _ string, doc any) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if s, ok := doc.(*Subtitle); ok {
		cp := *s
		m.updates = append(m.updates, &cp)
	}
	return nil
}

func (m *recordingStore) Delete(_ context.Context, _ string, _ string) error { return nil }

func (m *recordingStore) Query(_ context.Context, _ string, _ abstraction.QueryFilter, result any) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if r, ok := result.(*[]*Subtitle); ok {
		cp := make([]*Subtitle, len(m.subs))
		for i, s := range m.subs {
			cc := *s
			cp[i] = &cc
		}
		*r = cp
	}
	return nil
}

// BUG: SetDefault(videoID, unknownID) sets IsDefault=false on every subtitle
// and leaves NO default (clears the existing default instead of rejecting the
// unknown id). No DB-level uniqueness exists either, so two defaults can
// coexist when data is inserted outside SetDefault. File: subtitle_repository.go:110-122.
func TestSetDefault_InvalidIDClearsAllDefaults_Pinned(t *testing.T) {
	store := &recordingStore{
		subs: []*Subtitle{
			{ID: "a", VideoID: 10, IsDefault: true},
			{ID: "b", VideoID: 10, IsDefault: false},
		},
	}
	repo := NewRepository(store)
	err := repo.SetDefault(context.Background(), 10, "does-not-exist")
	require.NoError(t, err)
	// Pinned: after SetDefault with an unknown id, no subtitle is default.
	anyDefault := false
	for _, u := range store.updates {
		if u.IsDefault {
			anyDefault = true
		}
	}
	assert.False(t, anyDefault, "pinned: SetDefault(unknown id) left zero defaults (should keep one)")
}

// BUG: Upload never checks whether a subtitle already exists for the
// video+language, so repeated imports of the same video/language create
// duplicate records. No dedup in Create/Upload. File: subtitle_repository.go:132-144.
func TestUpload_NoDuplicateDedup_Pinned(t *testing.T) {
	store := &recordingStore{}
	svc := NewService(NewRepository(store))
	_, err1 := svc.Upload(context.Background(), 10, 1, "zh-CN", "中文", "[]")
	_, err2 := svc.Upload(context.Background(), 10, 1, "zh-CN", "中文", "[]")
	require.NoError(t, err1)
	require.NoError(t, err2)
	assert.Equal(t, 2, store.inserts, "pinned: duplicate video+language imports not deduped")
}

// BUG: Create inserts a subtitle with IsDefault=true without clearing any
// existing default for the video (GenerateFromAudio sets IsDefault:true
// directly at generator.go:64 and never calls SetDefault). A video that
// already has a default ends up with TWO defaults — the "default subtitle
// uniqueness" invariant is not enforced. File: subtitle_repository.go:35-46;
// generator.go:58-67.
func TestCreate_AllowsTwoDefaults_Pinned(t *testing.T) {
	store := &recordingStore{
		subs: []*Subtitle{{ID: "a", VideoID: 10, IsDefault: true}},
	}
	repo := NewRepository(store)
	_, err := repo.Create(context.Background(), &Subtitle{VideoID: 10, IsDefault: true, Format: "srt"})
	require.NoError(t, err)
	defaultCount := 0
	for _, s := range store.subs {
		if s.IsDefault {
			defaultCount++
		}
	}
	assert.Equal(t, 2, defaultCount, "pinned: Create allows two defaults for one video")
}

// BUG: Upload also never validates video/manuscript ownership (no check that
// the caller owns the video), so an authenticated user can attach subtitles to
// any video_id. The import handler (handleImport) does not even require auth.
// File: subtitle_handler.go:264-290; subtitle_repository.go:132-144.
func TestUpload_NoOwnershipCheck_Pinned(t *testing.T) {
	store := &recordingStore{}
	svc := NewService(NewRepository(store))
	// Any caller id can attach to any video; the service never consults a
	// video/manuscript ownership store.
	_, err := svc.Upload(context.Background(), 999999, 1, "zh-CN", "中文", "[]")
	require.NoError(t, err)
	assert.Equal(t, 1, store.inserts, "pinned: upload accepted without ownership validation")
}
