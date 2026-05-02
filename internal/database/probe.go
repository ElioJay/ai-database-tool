package database

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

func ProbeSchema(ctx context.Context, db *sql.DB, spec DriverSpec, conn ConnectionConfig) (Schema, error) {
	switch spec.Dialect {
	case "mysql":
		return probeMySQL(ctx, db, conn.Database)
	case "oracle":
		schema := conn.Schema
		if schema == "" {
			schema = strings.ToUpper(conn.Username)
		}
		return probeOracleLike(ctx, db, schema, false)
	case "dm":
		schema := conn.Schema
		if schema == "" {
			schema = strings.ToUpper(conn.Username)
		}
		return probeOracleLike(ctx, db, schema, true)
	default:
		return Schema{}, fmt.Errorf("不支持 schema 探测的数据库类型: %s", spec.Dialect)
	}
}

func probeMySQL(ctx context.Context, db *sql.DB, database string) (Schema, error) {
	rows, err := db.QueryContext(ctx, `
select table_name, column_name, column_type, is_nullable, column_key
from information_schema.columns
where table_schema = ?
order by table_name, ordinal_position`, database)
	if err != nil {
		return Schema{}, err
	}
	defer rows.Close()
	return scanSchemaRows(rows)
}

func probeOracleLike(ctx context.Context, db *sql.DB, schema string, dm bool) (Schema, error) {
	query := `
select c.table_name, c.column_name, c.data_type, c.nullable,
       case when pk.column_name is null then '' else 'PRI' end as column_key
from all_tab_columns c
left join (
  select acc.table_name, acc.column_name
  from all_constraints ac
  join all_cons_columns acc on ac.owner = acc.owner and ac.constraint_name = acc.constraint_name
  where ac.constraint_type = 'P' and ac.owner = :1
) pk on pk.table_name = c.table_name and pk.column_name = c.column_name
where c.owner = :2
order by c.table_name, c.column_id`
	if dm {
		query = strings.ReplaceAll(query, ":1", "?")
		query = strings.ReplaceAll(query, ":2", "?")
	}
	rows, err := db.QueryContext(ctx, query, strings.ToUpper(schema), strings.ToUpper(schema))
	if err != nil {
		return Schema{}, err
	}
	defer rows.Close()
	return scanSchemaRows(rows)
}

func scanSchemaRows(rows *sql.Rows) (Schema, error) {
	tableMap := map[string]*Table{}
	var order []string
	for rows.Next() {
		var tableName, columnName, dataType, nullable, key string
		if err := rows.Scan(&tableName, &columnName, &dataType, &nullable, &key); err != nil {
			return Schema{}, err
		}
		t, ok := tableMap[tableName]
		if !ok {
			t = &Table{Name: tableName}
			tableMap[tableName] = t
			order = append(order, tableName)
		}
		t.Columns = append(t.Columns, Column{
			Name:       columnName,
			Type:       dataType,
			Nullable:   strings.EqualFold(nullable, "YES") || strings.EqualFold(nullable, "Y"),
			PrimaryKey: strings.EqualFold(key, "PRI"),
		})
	}
	if err := rows.Err(); err != nil {
		return Schema{}, err
	}
	var schema Schema
	for _, name := range order {
		schema.Tables = append(schema.Tables, *tableMap[name])
	}
	return schema, nil
}
