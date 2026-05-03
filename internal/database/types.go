package database

type ConnectionConfig struct {
	Type     string
	Host     string
	Port     int
	Username string
	Password string
	Database string
	Schema   string
	DSN      string
	Include  []string
	Exclude  []string
}

type DriverSpec struct {
	DriverName string
	DSN        string
	SafeDSN    string
	Dialect    string
}

type QueryResult struct {
	Columns      []string
	Rows         [][]string
	RowsAffected int64
	DurationMS   int64
	Truncated    bool
}

type Column struct {
	Name       string
	Type       string
	Nullable   bool
	PrimaryKey bool
}

type Table struct {
	Name    string
	Comment string
	Columns []Column
}

type Schema struct {
	Tables []Table
}

type SchemaOptions struct {
	Include            []string
	Exclude            []string
	MaxTables          int
	MaxColumnsPerTable int
}
