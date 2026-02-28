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

// NewDB 建立並回傳 DBWrapper，包含無限重試連線機制
func NewDB(dsn string, checkSchema bool) *DBWrapper {
	connector := pgdriver.NewConnector(pgdriver.WithDSN(dsn))
	sqlDB := sql.OpenDB(connector)

	// 設定連線池基本參數
	sqlDB.SetMaxOpenConns(10)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)

	// 無限重試連線，直到資料庫就緒
	for {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		err := sqlDB.PingContext(ctx)
		cancel()

		if err == nil {
			log.Println("✅ Successfully connected to database.")
			break
		}

		log.Printf("⏳ Database not ready, retrying in 5 seconds... (error: %v)", err)
		time.Sleep(5 * time.Second)
	}

	db := bun.NewDB(sqlDB, pgdialect.New())

	if checkSchema {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
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

// verifyTables 檢查資料表完整性
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
			log.Printf("⚠️ Warning: Table verification query failed for %s: %v", fullTable, err)
			continue
		}

		cols := map[string]bool{}
		for _, col := range columns {
			cols[col] = true
		}

		for _, expected := range t.Columns {
			if !cols[expected] {
				log.Printf("⚠️ Warning: Missing column '%s' in table %s", expected, fullTable)
			}
		}
	}
}
