package database

import (
	"fmt"
	"net/url"
	"strings"
)

func BuildDriverSpec(conn ConnectionConfig) (DriverSpec, error) {
	dbType := strings.ToLower(strings.TrimSpace(conn.Type))
	switch dbType {
	case "mysql":
		return mysqlSpec(conn), nil
	case "oracle":
		return oracleSpec(conn), nil
	case "dm", "dameng":
		return dmSpec(conn), nil
	default:
		return DriverSpec{}, fmt.Errorf("不支持的数据库类型: %s", conn.Type)
	}
}

func mysqlSpec(conn ConnectionConfig) DriverSpec {
	if conn.DSN != "" {
		return DriverSpec{DriverName: "mysql", DSN: conn.DSN, SafeDSN: maskDSN(conn.DSN), Dialect: "mysql"}
	}
	port := conn.Port
	if port == 0 {
		port = 3306
	}
	params := "parseTime=true&timeout=5s&readTimeout=30s&writeTimeout=30s&multiStatements=false"
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?%s",
		conn.Username, conn.Password, conn.Host, port, conn.Database, params)
	safe := fmt.Sprintf("%s:***@tcp(%s:%d)/%s?%s",
		conn.Username, conn.Host, port, conn.Database, params)
	return DriverSpec{DriverName: "mysql", DSN: dsn, SafeDSN: safe, Dialect: "mysql"}
}

func oracleSpec(conn ConnectionConfig) DriverSpec {
	if conn.DSN != "" {
		return DriverSpec{DriverName: "oracle", DSN: conn.DSN, SafeDSN: maskDSN(conn.DSN), Dialect: "oracle"}
	}
	port := conn.Port
	if port == 0 {
		port = 1521
	}
	user := url.PathEscape(conn.Username)
	pass := url.PathEscape(conn.Password)
	service := strings.TrimPrefix(conn.Database, "/")
	dsn := fmt.Sprintf("oracle://%s:%s@%s:%d/%s", user, pass, conn.Host, port, service)
	safe := fmt.Sprintf("oracle://%s:***@%s:%d/%s", user, conn.Host, port, service)
	return DriverSpec{DriverName: "oracle", DSN: dsn, SafeDSN: safe, Dialect: "oracle"}
}

func dmSpec(conn ConnectionConfig) DriverSpec {
	if conn.DSN != "" {
		return DriverSpec{DriverName: "dm", DSN: conn.DSN, SafeDSN: maskDSN(conn.DSN), Dialect: "dm"}
	}
	port := conn.Port
	if port == 0 {
		port = 5236
	}
	user := url.PathEscape(conn.Username)
	pass := url.PathEscape(conn.Password)
	dsn := fmt.Sprintf("dm://%s:%s@%s:%d", user, pass, conn.Host, port)
	if conn.Schema != "" {
		dsn += "?schema=" + url.QueryEscape(conn.Schema)
	}
	safe := fmt.Sprintf("dm://%s:***@%s:%d", user, conn.Host, port)
	return DriverSpec{DriverName: "dm", DSN: dsn, SafeDSN: safe, Dialect: "dm"}
}

func maskDSN(dsn string) string {
	at := strings.LastIndex(dsn, "@")
	if at <= 0 {
		return dsn
	}
	colon := strings.LastIndex(dsn[:at], ":")
	if colon <= 0 {
		return "***"
	}
	return dsn[:colon+1] + "***" + dsn[at:]
}
