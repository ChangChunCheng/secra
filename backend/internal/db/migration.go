package db

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/migrate"
)

// NewMigrator 初始化 migrator，載入指定路徑下的 migration
func NewMigrator(db *bun.DB) *migrate.Migrator {
	migrations := migrate.NewMigrations()

	log.Println("📦 Discovering migrations from: migrations/v0")
	if err := migrations.Discover(os.DirFS("migrations/v0")); err != nil {
		log.Fatalf("❌ failed to discover migrations: %v", err)
	}

	m := migrate.NewMigrator(db, migrations)

	if err := m.Init(context.Background()); err != nil {
		log.Fatalf("❌ failed to init migrator: %v", err)
	}
	log.Println("✅ Migrations discovered successfully.")

	return m
}

// RunUp 執行所有尚未執行的 migration
func RunUp(db *bun.DB) {
	ctx := context.Background()
	m := NewMigrator(db)

	if err := m.Init(ctx); err != nil {
		log.Fatalf("migration init failed: %v", err)
	}

	// 檢查是否有待執行的 migrations
	migrations, err := m.MigrationsWithStatus(ctx)
	if err != nil {
		log.Fatalf("failed to get migration status: %v", err)
	}

	// 計算待執行的 migration 數量
	pendingCount := 0
	for _, mig := range migrations {
		if !mig.IsApplied() {
			pendingCount++
		}
	}

	if pendingCount == 0 {
		log.Println("✅ All migrations already applied, skipping.")
		return
	}

	log.Printf("📦 Found %d pending migration(s), applying...", pendingCount)

	if err := m.Lock(ctx); err != nil {
		log.Fatalf("migration lock failed: %v", err)
	}
	defer func() {
		if err := m.Unlock(ctx); err != nil {
			log.Fatalf("migration unlock failed: %v", err)
		}
	}()

	group, err := m.Migrate(ctx)
	if err != nil {
		log.Fatalf("migration up failed: %v", err)
	}

	if group.IsZero() {
		log.Println("✅ No new migrations to apply.")
	} else {
		log.Printf("✅ Successfully applied %d migration(s).", len(group.Migrations))
	}
}

// RunStatus 顯示 migration 狀態
func RunStatus(db *bun.DB) {
	ctx := context.Background()
	m := NewMigrator(db)

	if err := m.Init(ctx); err != nil {
		log.Fatalf("migration init failed: %v", err)
	}

	migrations, err := m.MigrationsWithStatus(ctx)
	if err != nil {
		log.Fatalf("failed to get migration status: %v", err)
	}

	fmt.Println("Migration Status:")
	for _, mig := range migrations {
		status := "Pending"
		if mig.IsApplied() {
			status = "Applied"
		}
		fmt.Printf("- %s: %s\n", mig.Name, status)
	}
}
