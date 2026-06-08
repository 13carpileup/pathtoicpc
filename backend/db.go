package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	mysql "github.com/go-sql-driver/mysql"
)

func openDatabase() (*sql.DB, error) {
	dsn := strings.TrimSpace(os.Getenv("MYSQL_DSN"))
	if dsn == "" {
		dsn = buildMySQLDSNFromEnv()
	}

	if dsn == "" {
		log.Print("mysql database is not configured; account endpoints are disabled")
		return nil, nil
	}

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("open mysql connection: %w", err)
	}

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping mysql database: %w", err)
	}

	return db, nil
}

func buildMySQLDSNFromEnv() string {
	dbName := strings.TrimSpace(os.Getenv("DB_NAME"))
	if dbName == "" {
		return ""
	}

	config := mysql.NewConfig()
	config.User = getEnv("DB_USER", "root")
	config.Passwd = os.Getenv("DB_PASSWORD")
	config.Net = "tcp"
	config.Addr = fmt.Sprintf("%s:%s", getEnv("DB_HOST", "127.0.0.1"), getEnv("DB_PORT", "3306"))
	config.DBName = dbName
	config.ParseTime = true
	config.Params = map[string]string{
		"charset":   "utf8mb4",
		"collation": "utf8mb4_unicode_ci",
	}

	return config.FormatDSN()
}
