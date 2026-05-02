package database

import (
	"strings"
	"testing"
)

func TestSchemaSummaryFiltersAndLimits(t *testing.T) {
	schema := Schema{
		Tables: []Table{
			{Name: "users", Columns: []Column{{Name: "id", Type: "int", PrimaryKey: true}, {Name: "name", Type: "varchar"}}},
			{Name: "audit_log", Columns: []Column{{Name: "id", Type: "int"}, {Name: "payload", Type: "text"}}},
		},
	}
	opts := SchemaOptions{Include: []string{"users"}, MaxTables: 10, MaxColumnsPerTable: 5}

	got := schema.Summary(opts)
	if !strings.Contains(got, "users(id int PK, name varchar)") {
		t.Fatalf("summary missing users table: %s", got)
	}
	if strings.Contains(got, "audit_log") {
		t.Fatalf("summary should exclude audit_log: %s", got)
	}
}

func TestBuildDSNDoesNotExposePasswordInSafeString(t *testing.T) {
	conn := ConnectionConfig{Type: "mysql", Host: "127.0.0.1", Port: 3306, Username: "root", Password: "secret", Database: "app"}
	spec, err := BuildDriverSpec(conn)
	if err != nil {
		t.Fatalf("BuildDriverSpec() error = %v", err)
	}
	if strings.Contains(spec.SafeDSN, "secret") {
		t.Fatalf("SafeDSN leaked password: %s", spec.SafeDSN)
	}
	if spec.DriverName != "mysql" {
		t.Fatalf("DriverName = %q", spec.DriverName)
	}
}
