package imageutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCompressToWebP(t *testing.T) {
	out, err := CompressToWebP("/data/a.jpg")
	require.NoError(t, err)
	assert.Equal(t, "/data/a.jpg", out)
}

func TestCompressToWebP_EmptyPath(t *testing.T) {
	_, err := CompressToWebP("")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "empty path")
}

func TestCompressAndReplace(t *testing.T) {
	out, err := CompressAndReplace("/data/dir/file.webp")
	require.NoError(t, err)
	assert.Equal(t, "file.webp", out)
}

func TestCompressAndReplace_RootPath(t *testing.T) {
	out, err := CompressAndReplace("file.webp")
	require.NoError(t, err)
	assert.Equal(t, "file.webp", out)
}

func TestCompressAndReplace_EmptyPath(t *testing.T) {
	_, err := CompressAndReplace("")
	assert.Error(t, err)
}
