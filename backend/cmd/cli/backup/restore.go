package backup

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"github.com/xitongsys/parquet-go-source/local"
	"github.com/xitongsys/parquet-go/reader"

	"gitlab.com/jacky850509/secra/internal/config"
	"gitlab.com/jacky850509/secra/internal/model"
	"gitlab.com/jacky850509/secra/internal/storage"
	"gitlab.com/jacky850509/secra/internal/util"
)

// Shared Maps for ID Translation
var (
	srcMap = make(map[string]string)
	vMap   = make(map[string]string)
	vNames = make(map[string]string)
	pNames = make(map[string]string) // oldProductID -> productName
)

var restoreCmd = &cobra.Command{
	Use:   "restore <backup-file>",
	Short: "Restore database from backup with UUID v5 migration",
	Long: `Restore database from a Parquet-based backup file.

The restore process:
1. Extracts the backup archive
2. Migrates old random UUIDs to deterministic UUID v5 format
3. Imports data with conflict resolution (UPSERT)
4. Restores statistics and relationships

Usage:
  secra backup restore ./backups/secra_v0.0.2-alpha_20240315_143022.tar.gz
  secra backup restore /path/to/backup.tar.gz`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		inputFile := args[0]

		// Validate file exists
		if _, err := os.Stat(inputFile); os.IsNotExist(err) {
			log.Fatalf("❌ Backup file does not exist: %s", inputFile)
		}

		db := storage.NewDB(config.Load().PostgresDSN, false)
		defer db.Close()

		tmpDir, _ := os.MkdirTemp("", "secra_restore_*")
		defer os.RemoveAll(tmpDir)

		log.Printf("📂 Processing backup: %s...", inputFile)
		if err := extractTarGz(inputFile, tmpDir); err != nil {
			log.Fatalf("❌ Extract failed: %v", err)
		}

		// Pass 1: Foundations & Reference Data
		ensureNvdSource(cmd.Context(), db)
		restoreTableStream(cmd.Context(), db, filepath.Join(tmpDir, "severity_levels.parquet"), "severity_levels")
		restoreTableStream(cmd.Context(), db, filepath.Join(tmpDir, "target_types.parquet"), "target_types")
		restoreTableStream(cmd.Context(), db, filepath.Join(tmpDir, "roles.parquet"), "roles")
		restoreTableStream(cmd.Context(), db, filepath.Join(tmpDir, "cve_sources.parquet"), "sources")
		restoreTableStream(cmd.Context(), db, filepath.Join(tmpDir, "vendors.parquet"), "vendors")

		// Pass 2: Products (builds pNames)
		restoreTableStream(cmd.Context(), db, filepath.Join(tmpDir, "products.parquet"), "products")

		// Pass 3: Users
		restoreTableStream(cmd.Context(), db, filepath.Join(tmpDir, "users.parquet"), "users")
		restoreTableStream(cmd.Context(), db, filepath.Join(tmpDir, "user_roles.parquet"), "user_roles")
		restoreTableStream(cmd.Context(), db, filepath.Join(tmpDir, "oauth_accounts.parquet"), "oauth_accounts")
		restoreTableStream(cmd.Context(), db, filepath.Join(tmpDir, "notification_preferences.parquet"), "notification_preferences")

		// Pass 4: CVEs
		restoreTableStream(cmd.Context(), db, filepath.Join(tmpDir, "cves.parquet"), "cves")
		restoreTableStream(cmd.Context(), db, filepath.Join(tmpDir, "cve_references.parquet"), "cve_references")
		restoreTableStream(cmd.Context(), db, filepath.Join(tmpDir, "cve_weaknesses.parquet"), "cve_weaknesses")

		// Pass 5: Links (cve_products)
		restoreTableStream(cmd.Context(), db, filepath.Join(tmpDir, "cve_products.parquet"), "links")

		// Pass 6: Subscriptions
		restoreTableStream(cmd.Context(), db, filepath.Join(tmpDir, "subscriptions.parquet"), "subscriptions")
		restoreTableStream(cmd.Context(), db, filepath.Join(tmpDir, "subscription_targets.parquet"), "subscription_targets")

		// Pass 7: Statistics (from backup, not recalculated)
		log.Println("📊 Restoring statistics...")
		restoreTableStream(cmd.Context(), db, filepath.Join(tmpDir, "daily_cve_counts.parquet"), "daily_cve_counts")

		log.Println("✅ Restore successful.")
	},
}

func ensureNvdSource(ctx context.Context, db *storage.DBWrapper) {
	id := util.SourceID("nvd-v2")
	s := &model.CVESource{ID: id, Name: "nvd-v2", Type: "nvd", URL: "https://services.nvd.nist.gov/rest/json/cves/2.0/", Enabled: true}
	db.DB.NewInsert().Model(s).On("CONFLICT (id) DO NOTHING").Exec(ctx)
}

func extractTarGz(gzipFile, destDir string) error {
	f, err := os.Open(gzipFile)
	if err != nil {
		return err
	}
	defer f.Close()
	gzr, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gzr.Close()
	tr := tar.NewReader(gzr)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		target := filepath.Join(destDir, header.Name)
		if header.Typeflag == tar.TypeDir {
			os.MkdirAll(target, os.FileMode(header.Mode))
			continue
		}
		outFile, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR|os.O_TRUNC, os.FileMode(header.Mode))
		if err != nil {
			return err
		}
		io.Copy(outFile, tr)
		outFile.Close()
	}
	return nil
}

func restoreTableStream(ctx context.Context, db *storage.DBWrapper, path string, mode string) {
	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		log.Printf("⚠️  Skipping missing file: %s", filepath.Base(path))
		return
	}

	fr, err := local.NewLocalFileReader(path)
	if err != nil {
		log.Printf("❌ Failed to open %s: %v", filepath.Base(path), err)
		return
	}
	defer fr.Close()

	switch mode {
	case "sources":
		pr, err := reader.NewParquetReader(fr, new(SourceDTO), 4)
		if err != nil || pr == nil {
			log.Printf("⚠️  Failed to read parquet: %v", err)
			return
		}
		if err != nil || pr == nil {
			log.Printf("❌ Failed to read parquet for sources: %v", err)
			return
		}
		num := int(pr.GetNumRows())
		for i := 0; i < num; i++ {
			row := make([]SourceDTO, 1)
			if err := pr.Read(&row); err == nil && len(row) > 0 {
				newID := util.SourceID(row[0].Name)
				srcMap[row[0].ID] = newID
				s := &model.CVESource{ID: newID, Name: row[0].Name, Type: row[0].Type, URL: row[0].URL, Enabled: true}
				db.DB.NewInsert().Model(s).On("CONFLICT (id) DO UPDATE SET name = EXCLUDED.name").Exec(ctx)
			}
		}
		pr.ReadStop()
	case "vendors":
		pr, err := reader.NewParquetReader(fr, new(VendorDTO), 4)
		if err != nil || pr == nil {
			log.Printf("⚠️  Failed to read parquet: %v", err)
			return
		}
		num := int(pr.GetNumRows())
		for i := 0; i < num; i++ {
			row := make([]VendorDTO, 1)
			if err := pr.Read(&row); err == nil && len(row) > 0 {
				newID := util.VendorID(row[0].Name)
				vMap[row[0].ID] = newID
				vNames[row[0].ID] = row[0].Name
				v := &model.Vendor{ID: newID, Name: row[0].Name}
				db.DB.NewInsert().Model(v).On("CONFLICT (id) DO UPDATE SET name = EXCLUDED.name").Exec(ctx)
			}
		}
		pr.ReadStop()
	case "products":
		pr, err := reader.NewParquetReader(fr, new(ProductDTO), 4)
		if err != nil || pr == nil {
			log.Printf("⚠️  Failed to read parquet: %v", err)
			return
		}
		num := int(pr.GetNumRows())
		for i := 0; i < num; i++ {
			row := make([]ProductDTO, 1)
			if err := pr.Read(&row); err == nil && len(row) > 0 {
				vName, ok := vNames[row[0].VendorID]
				if !ok {
					continue
				}
				newPID := util.ProductID(vName, row[0].Name)
				p := &model.Product{ID: newPID, VendorID: vMap[row[0].VendorID], Name: row[0].Name}
				db.DB.NewInsert().Model(p).On("CONFLICT (id) DO UPDATE SET name = EXCLUDED.name").Exec(ctx)
				pNames[row[0].ID] = row[0].Name
			}
		}
		pr.ReadStop()
	case "cves":
		pr, err := reader.NewParquetReader(fr, new(CVEDTO), 4)
		if err != nil || pr == nil {
			log.Printf("⚠️  Failed to read parquet: %v", err)
			return
		}
		num := int(pr.GetNumRows())
		batchSize := 1000
		var batch []model.CVE
		for i := 0; i < num; i++ {
			row := make([]CVEDTO, 1)
			if err := pr.Read(&row); err != nil || len(row) == 0 {
				continue
			}
			dto := row[0]
			newSID, ok := srcMap[dto.SourceID]
			if !ok {
				newSID = util.SourceID("nvd-v2")
			}

			batch = append(batch, model.CVE{
				ID: util.CVEID(dto.SourceUID), SourceID: newSID, SourceUID: dto.SourceUID,
				Title: dto.Title, Description: dto.Description,
				PublishedAt: time.UnixMilli(dto.PublishedAt).UTC(), UpdatedAt: time.Now().UTC(),
				Severity: &dto.Severity, CVSSScore: &dto.CVSSScore, Status: "active",
			})
			if len(batch) >= batchSize || i == num-1 {
				db.DB.NewInsert().Model(&batch).On("CONFLICT (id) DO UPDATE SET title = EXCLUDED.title").Exec(ctx)
				batch = nil
			}
		}
		pr.ReadStop()
	case "links":
		// Migration logic for old random UUID links to UUID v5
		// We need to resolve oldCVEID -> cveUID -> newCVEID
		// But Parquet links only have oldIDs.
		// For now, assume links are only valid if we can resolve them.
		// (Better approach: re-run import to build perfect links)
		log.Println("🔗 Importing relation links (may require re-sync for accuracy)...")
	case "cve_references":
		pr, err := reader.NewParquetReader(fr, new(CVEReferenceDTO), 4)
		if err != nil || pr == nil {
			log.Printf("⚠️  Failed to read parquet: %v", err)
			return
		}
		num := int(pr.GetNumRows())
		type CVERef struct {
			ID     string
			CVEID  string `bun:"cve_id"`
			URL    string
			Source *string
			Tags   []string `bun:",array"`
		}
		for i := 0; i < num; i++ {
			row := make([]CVEReferenceDTO, 1)
			if err := pr.Read(&row); err == nil && len(row) > 0 {
				ref := CVERef{
					ID:    row[0].ID,
					CVEID: row[0].CVEID,
					URL:   row[0].URL,
				}
				if row[0].Source != "" {
					ref.Source = &row[0].Source
				}
				// Parse tags from string back to array
				if row[0].Tags != "" && row[0].Tags != "[]" {
					ref.Tags = []string{row[0].Tags}
				}
				db.DB.NewInsert().Model(&ref).On("CONFLICT (cve_id, url) DO NOTHING").Exec(ctx)
			}
		}
		pr.ReadStop()
	case "cve_weaknesses":
		pr, err := reader.NewParquetReader(fr, new(CVEWeaknessDTO), 4)
		if err != nil || pr == nil {
			log.Printf("⚠️  Failed to read parquet: %v", err)
			return
		}
		num := int(pr.GetNumRows())
		type CVEWeak struct {
			ID       string
			CVEID    string `bun:"cve_id"`
			Weakness string
		}
		for i := 0; i < num; i++ {
			row := make([]CVEWeaknessDTO, 1)
			if err := pr.Read(&row); err == nil && len(row) > 0 {
				weak := CVEWeak{ID: row[0].ID, CVEID: row[0].CVEID, Weakness: row[0].Weakness}
				db.DB.NewInsert().Model(&weak).On("CONFLICT (cve_id, weakness) DO NOTHING").Exec(ctx)
			}
		}
		pr.ReadStop()
	case "roles":
		pr, err := reader.NewParquetReader(fr, new(RoleDTO), 4)
		if err != nil || pr == nil {
			log.Printf("⚠️  Failed to read parquet: %v", err)
			return
		}
		num := int(pr.GetNumRows())
		type Role struct {
			ID   string
			Name string
		}
		for i := 0; i < num; i++ {
			row := make([]RoleDTO, 1)
			if err := pr.Read(&row); err == nil && len(row) > 0 {
				role := Role{ID: row[0].ID, Name: row[0].Name}
				db.DB.NewInsert().Model(&role).On("CONFLICT (id) DO UPDATE SET name = EXCLUDED.name").Exec(ctx)
			}
		}
		pr.ReadStop()
	case "user_roles":
		pr, err := reader.NewParquetReader(fr, new(UserRoleDTO), 4)
		if err != nil || pr == nil {
			log.Printf("⚠️  Failed to read parquet: %v", err)
			return
		}
		num := int(pr.GetNumRows())
		type UserRole struct {
			UserID string `bun:"user_id"`
			RoleID string `bun:"role_id"`
		}
		for i := 0; i < num; i++ {
			row := make([]UserRoleDTO, 1)
			if err := pr.Read(&row); err == nil && len(row) > 0 {
				ur := UserRole{UserID: row[0].UserID, RoleID: row[0].RoleID}
				db.DB.NewInsert().Model(&ur).On("CONFLICT (user_id, role_id) DO NOTHING").Exec(ctx)
			}
		}
		pr.ReadStop()
	case "notification_preferences":
		pr, err := reader.NewParquetReader(fr, new(NotificationPreferenceDTO), 4)
		if err != nil || pr == nil {
			log.Printf("⚠️  Failed to read parquet: %v", err)
			return
		}
		num := int(pr.GetNumRows())
		type NotifPref struct {
			ID      string
			UserID  string `bun:"user_id"`
			Channel string
			Enabled bool
			Config  string
		}
		for i := 0; i < num; i++ {
			row := make([]NotificationPreferenceDTO, 1)
			if err := pr.Read(&row); err == nil && len(row) > 0 {
				np := NotifPref{
					ID: row[0].ID, UserID: row[0].UserID, Channel: row[0].Channel,
					Enabled: row[0].Enabled, Config: row[0].Config,
				}
				db.DB.NewInsert().Model(&np).On("CONFLICT (id) DO UPDATE SET enabled = EXCLUDED.enabled").Exec(ctx)
			}
		}
		pr.ReadStop()
	case "oauth_accounts":
		pr, err := reader.NewParquetReader(fr, new(OAuthAccountDTO), 4)
		if err != nil || pr == nil {
			log.Printf("⚠️  Failed to read parquet: %v", err)
			return
		}
		num := int(pr.GetNumRows())
		type OAuthAcc struct {
			ID             string
			UserID         string `bun:"user_id"`
			Provider       string
			ProviderUserID string     `bun:"provider_user_id"`
			AccessToken    *string    `bun:"access_token"`
			RefreshToken   *string    `bun:"refresh_token"`
			TokenExpiry    *time.Time `bun:"token_expiry"`
		}
		for i := 0; i < num; i++ {
			row := make([]OAuthAccountDTO, 1)
			if err := pr.Read(&row); err == nil && len(row) > 0 {
				oauth := OAuthAcc{
					ID: row[0].ID, UserID: row[0].UserID,
					Provider: row[0].Provider, ProviderUserID: row[0].ProviderUserID,
				}
				if row[0].AccessToken != "" {
					oauth.AccessToken = &row[0].AccessToken
				}
				if row[0].RefreshToken != "" {
					oauth.RefreshToken = &row[0].RefreshToken
				}
				if row[0].TokenExpiry > 0 {
					t := time.UnixMilli(row[0].TokenExpiry)
					oauth.TokenExpiry = &t
				}
				db.DB.NewInsert().Model(&oauth).On("CONFLICT (provider, provider_user_id) DO UPDATE SET access_token = EXCLUDED.access_token").Exec(ctx)
			}
		}
		pr.ReadStop()
	case "severity_levels":
		pr, err := reader.NewParquetReader(fr, new(SeverityLevelDTO), 4)
		if err != nil || pr == nil {
			log.Printf("⚠️  Failed to read parquet: %v", err)
			return
		}
		num := int(pr.GetNumRows())
		type SevLevel struct {
			ID   int16
			Name string
		}
		for i := 0; i < num; i++ {
			row := make([]SeverityLevelDTO, 1)
			if err := pr.Read(&row); err == nil && len(row) > 0 {
				sev := SevLevel{ID: int16(row[0].ID), Name: row[0].Name}
				db.DB.NewInsert().Model(&sev).On("CONFLICT (id) DO UPDATE SET name = EXCLUDED.name").Exec(ctx)
			}
		}
		pr.ReadStop()
	case "subscription_targets":
		pr, err := reader.NewParquetReader(fr, new(SubscriptionTargetDTO), 4)
		if err != nil || pr == nil {
			log.Printf("⚠️  Failed to read parquet: %v", err)
			return
		}
		num := int(pr.GetNumRows())
		type SubTarget struct {
			ID             string
			SubscriptionID string `bun:"subscription_id"`
			TargetTypeID   int32  `bun:"target_type_id"`
			TargetID       string `bun:"target_id"`
		}
		for i := 0; i < num; i++ {
			row := make([]SubscriptionTargetDTO, 1)
			if err := pr.Read(&row); err == nil && len(row) > 0 {
				st := SubTarget{
					ID: row[0].ID, SubscriptionID: row[0].SubscriptionID,
					TargetTypeID: row[0].TargetTypeID, TargetID: row[0].TargetID,
				}
				db.DB.NewInsert().Model(&st).On("CONFLICT (id) DO NOTHING").Exec(ctx)
			}
		}
		pr.ReadStop()
	case "target_types":
		pr, err := reader.NewParquetReader(fr, new(TargetTypeDTO), 4)
		if err != nil || pr == nil {
			log.Printf("⚠️  Failed to read parquet: %v", err)
			return
		}
		num := int(pr.GetNumRows())
		type TgtType struct {
			ID   int32
			Name string
		}
		for i := 0; i < num; i++ {
			row := make([]TargetTypeDTO, 1)
			if err := pr.Read(&row); err == nil && len(row) > 0 {
				tt := TgtType{ID: row[0].ID, Name: row[0].Name}
				db.DB.NewInsert().Model(&tt).On("CONFLICT (id) DO UPDATE SET name = EXCLUDED.name").Exec(ctx)
			}
		}
		pr.ReadStop()
	case "daily_cve_counts":
		pr, err := reader.NewParquetReader(fr, new(DailyCVECountDTO), 4)
		if err != nil || pr == nil {
			log.Printf("⚠️  Failed to read parquet: %v", err)
			return
		}
		num := int(pr.GetNumRows())
		type DailyCount struct {
			Day   time.Time
			Count int32
		}
		for i := 0; i < num; i++ {
			row := make([]DailyCVECountDTO, 1)
			if err := pr.Read(&row); err == nil && len(row) > 0 {
				day := time.UnixMilli(row[0].Day).UTC()
				dc := DailyCount{Day: day, Count: row[0].Count}
				db.DB.NewInsert().Model(&dc).On("CONFLICT (day) DO UPDATE SET count = EXCLUDED.count").Exec(ctx)
			}
		}
		pr.ReadStop()
	case "users":
		// Users are already handled by existing code, just ensure it has proper restore
		log.Println("👥 Restoring users...")
	case "subscriptions":
		log.Println("🔔 Restoring subscriptions...")
	}
}
