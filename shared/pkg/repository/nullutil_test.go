package repository

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNullInt64_Zero(t *testing.T) {
	assert.Nil(t, NullInt64(0))
}

func TestNullInt64_NonZero(t *testing.T) {
	assert.Equal(t, int64(5), NullInt64(5))
}

func TestNullInt64_Negative(t *testing.T) {
	assert.Equal(t, int64(-3), NullInt64(-3))
}

func TestNullInt64FromSQL_Valid(t *testing.T) {
	v := int64(42)
	assert.Equal(t, &v, NullInt64FromSQL(sql.NullInt64{Int64: 42, Valid: true}))
}

func TestNullInt64FromSQL_Invalid(t *testing.T) {
	assert.Nil(t, NullInt64FromSQL(sql.NullInt64{Valid: false}))
}
