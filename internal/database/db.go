package database

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type DB struct {
	Name string
	Spec DriverSpec
	SQL  *sql.DB
}

func Open(name string, conn ConnectionConfig) (*DB, error) {
	spec, err := BuildDriverSpec(conn)
	if err != nil {
		return nil, err
	}
	db, err := sql.Open(spec.DriverName, spec.DSN)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(4)
	db.SetMaxIdleConns(2)
	db.SetConnMaxLifetime(3 * time.Minute)
	return &DB{Name: name, Spec: spec, SQL: db}, nil
}

func (db *DB) Close() error {
	if db == nil || db.SQL == nil {
		return nil
	}
	return db.SQL.Close()
}

func (db *DB) Ping(ctx context.Context) error {
	if db == nil || db.SQL == nil {
		return fmt.Errorf("数据库未打开")
	}
	ctx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()
	return db.SQL.PingContext(ctx)
}

func (db *DB) Query(ctx context.Context, sqlText string, maxRows int) (QueryResult, error) {
	start := time.Now()
	if maxRows <= 0 {
		maxRows = 100
	}
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	rows, err := db.SQL.QueryContext(ctx, sqlText)
	if err != nil {
		return QueryResult{}, err
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return QueryResult{}, err
	}
	result := QueryResult{Columns: cols}
	for rows.Next() {
		values := make([]any, len(cols))
		ptrs := make([]any, len(cols))
		for i := range values {
			ptrs[i] = &values[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return QueryResult{}, err
		}
		if len(result.Rows) < maxRows {
			result.Rows = append(result.Rows, stringifyRow(values))
		} else {
			result.Truncated = true
		}
	}
	if err := rows.Err(); err != nil {
		return QueryResult{}, err
	}
	result.DurationMS = time.Since(start).Milliseconds()
	return result, nil
}

func (db *DB) Exec(ctx context.Context, sqlText string) (QueryResult, error) {
	start := time.Now()
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	res, err := db.SQL.ExecContext(ctx, sqlText)
	if err != nil {
		return QueryResult{}, err
	}
	affected, _ := res.RowsAffected()
	return QueryResult{RowsAffected: affected, DurationMS: time.Since(start).Milliseconds()}, nil
}

func stringifyRow(values []any) []string {
	out := make([]string, len(values))
	for i, v := range values {
		switch x := v.(type) {
		case nil:
			out[i] = "NULL"
		case []byte:
			out[i] = string(x)
		case time.Time:
			out[i] = x.Format("2006-01-02 15:04:05")
		default:
			out[i] = strings.TrimSpace(strconv.Quote(fmt.Sprint(x)))
			out[i] = strings.Trim(out[i], "\"")
		}
	}
	return out
}
