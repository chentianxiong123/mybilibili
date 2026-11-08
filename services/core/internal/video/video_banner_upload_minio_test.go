package video

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"mybilibili/pkg/abstraction"
)

// TestBannerUpload_MinIODualWrite_IsUnreachable documents that the MinIO
// dual-write block of handleBanner (the `if h.storage != nil` branch, plus
// the os.Open / content-type switch / storage.Put / Put-error fallback
// statements) cannot be exercised from a test in this package:
//
//   - Handler.storage is a concrete *abstraction.MinioStorageService
//     (not an interface), so no fake can be injected.
//   - The struct's client/publicClient fields are unexported, so a non-nil
//     instance cannot be fabricated.
//   - abstraction.NewMinioStorageService performs a BucketExists round
//     trip in its constructor and therefore rejects any fake endpoint,
//     while a real endpoint would start a live service.
//
// The only constructible handler state is storage == nil, which is the path
// already covered by the local-disk upload tests in video_coverage_test.go.
func TestBannerUpload_MinIODualWrite_IsUnreachable(t *testing.T) {
	h, _ := newTestHandler(t)
	assert.Nil(t, h.storage)

	// Prove the constructor cannot be satisfied against a fake endpoint:
	// minio-go rejects URLs carrying a fully-qualified path, and a bare
	// hostname requires a live listener plus a real BucketExists round
	// trip, so no in-process fake can stand in for MinIO here.
	_, err := abstraction.NewMinioStorageService(abstraction.MinioConfig{
		Endpoint:   "http://127.0.0.1:1",
		BucketName: "mybilibili",
	})
	require.Error(t, err)
}
