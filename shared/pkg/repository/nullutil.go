package repository

import "database/sql"

func NullInt64(v int64) interface{} {
	if v == 0 {
		return nil
	}
	return v
}

func NullInt64FromSQL(n sql.NullInt64) *int64 {
	if n.Valid {
		return &n.Int64
	}
	return nil
}