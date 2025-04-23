package storage

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
)

// DBWrapper 封裝 bun.DB 並提供關閉方法
type DBWrapper struct {
	DB    *bun.DB
	sqlDB *sql.DB
}

// NewDB 建立並回傳 DBWrapper，包含連線池與錯誤偵測與資料表驗證
func NewDB(dsn string, checkSchema bool) *DBWrapper {
	sqlDB := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))

	// 連線設定 ...
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		log.Fatalf("Database ping failed: %v", err)
	}

	db := bun.NewDB(sqlDB, pgdialect.New())
	log.Println("Connected to database successfully")

	if checkSchema {
		verifyTables(db, ctx)
	}

	return &DBWrapper{DB: db, sqlDB: sqlDB}
}

// Close 關閉資料庫連線
func (w *DBWrapper) Close() {
	if err := w.sqlDB.Close(); err != nil {
		log.Printf("Failed to close database: %v", err)
	} else {
		log.Println("Database connection closed.")
	}
}

// TableDefinition 定義每個表格要驗證的 schema 與欄位
type TableDefinition struct {
	Schema  string
	Name    string
	Columns []string
}

// verifyTables 檢查多 schema 多表格與欄位是否存在
func verifyTables(db *bun.DB, ctx context.Context) {
	tables := []TableDefinition{
		{Schema: "public", Name: "cve_sources", Columns: []string{"id", "name", "type", "url"}},
		{Schema: "public", Name: "cves", Columns: []string{"id", "source_id", "source_uid", "title"}},
	}

	for _, t := range tables {
		fullTable := fmt.Sprintf("%s.%s", t.Schema, t.Name)
		query := fmt.Sprintf(`
			SELECT column_name
			FROM information_schema.columns
			WHERE table_schema = '%s' AND table_name = '%s'`, t.Schema, t.Name)

		var columns []string
		if err := db.NewRaw(query).Scan(ctx, &columns); err != nil {
			log.Fatalf("Failed to query columns for table %s: %v", fullTable, err)
		}

		cols := map[string]bool{}
		for _, col := range columns {
			cols[col] = true
		}

		for _, expected := range t.Columns {
			if !cols[expected] {
				log.Fatalf("Missing column '%s' in table %s", expected, fullTable)
			}
		}

		log.Printf("Table verified: %s (%d columns)\n", fullTable, len(cols))
	}

	log.Println("All required tables and columns verified.")
}
