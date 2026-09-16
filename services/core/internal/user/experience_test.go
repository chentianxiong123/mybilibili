package user

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLevelThreshold_Basics(t *testing.T) {
	assert.Equal(t, int64(100), LevelThreshold(1))
	assert.Equal(t, int64(348), LevelThreshold(2))
	assert.Equal(t, int64(722), LevelThreshold(3))
}

func TestLevelThreshold_Monotonic(t *testing.T) {
	var prev int64
	for lvl := int32(1); lvl <= 100; lvl++ {
		cur := LevelThreshold(lvl)
		assert.Greater(t, cur, prev, "level %d", lvl)
		prev = cur
	}
}

func TestLevelThreshold_InvalidLevelsClampTo1(t *testing.T) {
	assert.Equal(t, LevelThreshold(1), LevelThreshold(0))
	assert.Equal(t, LevelThreshold(1), LevelThreshold(-5))
}

func TestLevelThreshold_LargeValues(t *testing.T) {
	// floor(100 * 100^1.8) = 398107
	assert.Equal(t, int64(398107), LevelThreshold(100))
}

func TestAwardExperience_Guards(t *testing.T) {
	ctx := context.Background()
	cases := []struct {
		name     string
		userID   int64
		amount   int32
	}{
		{"zero user", 0, 100},
		{"zero amount", 1, 0},
		{"negative amount", 1, -10},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			awarded, lvl := AwardExperience(ctx, nil, c.userID, c.amount)
			assert.False(t, awarded)
			assert.Zero(t, lvl)
		})
	}
}
