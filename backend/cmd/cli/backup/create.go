package backup

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"github.com/xitongsys/parquet-go-source/local"
	"github.com/xitongsys/parquet-go/writer"

	"gitlab.com/jacky850509/secra/internal/config"
	"gitlab.com/jacky850509/secra/internal/model"
	"gitlab.com/jacky850509/secra/internal/storage"
)

var (
	outputFile string
	outputDir  string
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a Parquet-based backup with auto-generated filename",
	Long: `Create a complete database backup in Parquet format.

The backup filename is automatically generated as: secra_<version>_<timestamp>.tar.gz
You can specify an output directory with -d or a full path with -o.`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()
		db := storage.NewDB(cfg.PostgresDSN, false)
		defer db.Close()

		// Auto-generate filename with version and timestamp
		timestamp := time.Now().Format("20060102_150405")
		version := config.Version
		if version == "" {
			version = "dev"
		}

		if outputFile == "" {
			// Auto-generate filename
			filename := fmt.Sprintf("secra_%s_%s.tar.gz", version, timestamp)
			if outputDir != "" {
				// Use specified directory
				outputFile = filepath.Join(outputDir, filename)
			} else {
				// Use current directory
				outputFile = filename
			}
		}

		// Ensure output directory exists
		outputDir := filepath.Dir(outputFile)
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			log.Fatalf("❌ Failed to create output directory: %v", err)
		}

		tmpDir, _ := os.MkdirTemp("", "secra_backup_*")
		defer os.RemoveAll(tmpDir)

		log.Printf("📦 Starting backup process to %s...", outputFile)

		// Backup all tables in dependency order
		tables := []string{
			"cve_sources", "vendors", "products", "cves",
			"cve_products", "cve_references", "cve_weaknesses",
			"roles", "users", "user_roles",
			"subscriptions", "subscription_targets",
			"notification_preferences", "oauth_accounts",
			"severity_levels", "target_types",
			"daily_cve_counts",
		}
		for _, table := range tables {
			parquetFile := filepath.Join(tmpDir, table+".parquet")
			log.Printf("📄 Exporting table [%s]...", table)
			exportTableToParquet(cmd.Context(), db, table, parquetFile)
		}

		createTarGz(outputFile, tmpDir)
		log.Printf("✅ Backup successfully created: %s", outputFile)
	},
}

func init() {
	createCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Full output path (overrides -d). If not specified, auto-generates filename.")
	createCmd.Flags().StringVarP(&outputDir, "dir", "d", ".", "Output directory (default: current directory)")
}

type SourceDTO struct {
	ID   string `parquet:"name=id, type=BYTE_ARRAY, convertedtype=UTF8"`
	Name string `parquet:"name=name, type=BYTE_ARRAY, convertedtype=UTF8"`
	Type string `parquet:"name=type, type=BYTE_ARRAY, convertedtype=UTF8"`
	URL  string `parquet:"name=url, type=BYTE_ARRAY, convertedtype=UTF8"`
}

type VendorDTO struct {
	ID   string `parquet:"name=id, type=BYTE_ARRAY, convertedtype=UTF8"`
	Name string `parquet:"name=name, type=BYTE_ARRAY, convertedtype=UTF8"`
}

type ProductDTO struct {
	ID       string `parquet:"name=id, type=BYTE_ARRAY, convertedtype=UTF8"`
	VendorID string `parquet:"name=vendor_id, type=BYTE_ARRAY, convertedtype=UTF8"`
	Name     string `parquet:"name=name, type=BYTE_ARRAY, convertedtype=UTF8"`
}

type CVEDTO struct {
	ID          string  `parquet:"name=id, type=BYTE_ARRAY, convertedtype=UTF8"`
	SourceID    string  `parquet:"name=source_id, type=BYTE_ARRAY, convertedtype=UTF8"`
	SourceUID   string  `parquet:"name=source_uid, type=BYTE_ARRAY, convertedtype=UTF8"`
	Title       string  `parquet:"name=title, type=BYTE_ARRAY, convertedtype=UTF8"`
	Description string  `parquet:"name=description, type=BYTE_ARRAY, convertedtype=UTF8"`
	Severity    string  `parquet:"name=severity, type=BYTE_ARRAY, convertedtype=UTF8"`
	CVSSScore   float64 `parquet:"name=cvss_score, type=DOUBLE"`
	PublishedAt int64   `parquet:"name=published_at, type=INT64, convertedtype=TIMESTAMP_MILLIS"`
}

type CVEProductDTO struct {
	CVEID     string `parquet:"name=cve_id, type=BYTE_ARRAY, convertedtype=UTF8"`
	ProductID string `parquet:"name=product_id, type=BYTE_ARRAY, convertedtype=UTF8"`
}

type CVEReferenceDTO struct {
	ID     string `parquet:"name=id, type=BYTE_ARRAY, convertedtype=UTF8"`
	CVEID  string `parquet:"name=cve_id, type=BYTE_ARRAY, convertedtype=UTF8"`
	URL    string `parquet:"name=url, type=BYTE_ARRAY, convertedtype=UTF8"`
	Source string `parquet:"name=source, type=BYTE_ARRAY, convertedtype=UTF8"`
	Tags   string `parquet:"name=tags, type=BYTE_ARRAY, convertedtype=UTF8"` // JSON array as string
}

type CVEWeaknessDTO struct {
	ID       string `parquet:"name=id, type=BYTE_ARRAY, convertedtype=UTF8"`
	CVEID    string `parquet:"name=cve_id, type=BYTE_ARRAY, convertedtype=UTF8"`
	Weakness string `parquet:"name=weakness, type=BYTE_ARRAY, convertedtype=UTF8"`
}

type RoleDTO struct {
	ID   string `parquet:"name=id, type=BYTE_ARRAY, convertedtype=UTF8"`
	Name string `parquet:"name=name, type=BYTE_ARRAY, convertedtype=UTF8"`
}

type UserRoleDTO struct {
	UserID string `parquet:"name=user_id, type=BYTE_ARRAY, convertedtype=UTF8"`
	RoleID string `parquet:"name=role_id, type=BYTE_ARRAY, convertedtype=UTF8"`
}

type NotificationPreferenceDTO struct {
	ID      string `parquet:"name=id, type=BYTE_ARRAY, convertedtype=UTF8"`
	UserID  string `parquet:"name=user_id, type=BYTE_ARRAY, convertedtype=UTF8"`
	Channel string `parquet:"name=channel, type=BYTE_ARRAY, convertedtype=UTF8"`
	Enabled bool   `parquet:"name=enabled, type=BOOLEAN"`
	Config  string `parquet:"name=config, type=BYTE_ARRAY, convertedtype=UTF8"` // JSONB as string
}

type OAuthAccountDTO struct {
	ID             string `parquet:"name=id, type=BYTE_ARRAY, convertedtype=UTF8"`
	UserID         string `parquet:"name=user_id, type=BYTE_ARRAY, convertedtype=UTF8"`
	Provider       string `parquet:"name=provider, type=BYTE_ARRAY, convertedtype=UTF8"`
	ProviderUserID string `parquet:"name=provider_user_id, type=BYTE_ARRAY, convertedtype=UTF8"`
	AccessToken    string `parquet:"name=access_token, type=BYTE_ARRAY, convertedtype=UTF8"`
	RefreshToken   string `parquet:"name=refresh_token, type=BYTE_ARRAY, convertedtype=UTF8"`
	TokenExpiry    int64  `parquet:"name=token_expiry, type=INT64, convertedtype=TIMESTAMP_MILLIS"`
}

type SeverityLevelDTO struct {
	ID   int32  `parquet:"name=id, type=INT32"`
	Name string `parquet:"name=name, type=BYTE_ARRAY, convertedtype=UTF8"`
}

type SubscriptionTargetDTO struct {
	ID             string `parquet:"name=id, type=BYTE_ARRAY, convertedtype=UTF8"`
	SubscriptionID string `parquet:"name=subscription_id, type=BYTE_ARRAY, convertedtype=UTF8"`
	TargetTypeID   int32  `parquet:"name=target_type_id, type=INT32"`
	TargetID       string `parquet:"name=target_id, type=BYTE_ARRAY, convertedtype=UTF8"`
}

type TargetTypeDTO struct {
	ID   int32  `parquet:"name=id, type=INT32"`
	Name string `parquet:"name=name, type=BYTE_ARRAY, convertedtype=UTF8"`
}

type DailyCVECountDTO struct {
	Day   int64 `parquet:"name=day, type=INT64, convertedtype=DATE"`
	Count int32 `parquet:"name=count, type=INT32"`
}

func exportTableToParquet(ctx context.Context, db *storage.DBWrapper, tableName string, filePath string) {
	fw, _ := local.NewLocalFileWriter(filePath)
	defer fw.Close()

	switch tableName {
	case "cve_sources":
		var items []model.CVESource
		db.DB.NewSelect().Model(&items).Scan(ctx)
		pw, _ := writer.NewParquetWriter(fw, new(SourceDTO), 4)
		for _, item := range items {
			pw.Write(SourceDTO{ID: item.ID, Name: item.Name, Type: item.Type, URL: item.URL})
		}
		pw.WriteStop()
	case "vendors":
		var items []model.Vendor
		db.DB.NewSelect().Model(&items).Scan(ctx)
		pw, _ := writer.NewParquetWriter(fw, new(VendorDTO), 4)
		for _, item := range items {
			pw.Write(VendorDTO{ID: item.ID, Name: item.Name})
		}
		pw.WriteStop()
	case "products":
		var items []model.Product
		db.DB.NewSelect().Model(&items).Scan(ctx)
		pw, _ := writer.NewParquetWriter(fw, new(ProductDTO), 4)
		for _, item := range items {
			pw.Write(ProductDTO{ID: item.ID, VendorID: item.VendorID, Name: item.Name})
		}
		pw.WriteStop()
	case "cves":
		var items []model.CVE
		db.DB.NewSelect().Model(&items).Scan(ctx)
		pw, _ := writer.NewParquetWriter(fw, new(CVEDTO), 4)
		for _, item := range items {
			dto := CVEDTO{ID: item.ID, SourceID: item.SourceID, SourceUID: item.SourceUID, Title: item.Title, Description: item.Description, PublishedAt: item.PublishedAt.UnixMilli()}
			if item.Severity != nil {
				dto.Severity = *item.Severity
			}
			if item.CVSSScore != nil {
				dto.CVSSScore = *item.CVSSScore
			}
			pw.Write(dto)
		}
		pw.WriteStop()
	case "cve_products":
		var items []model.CVEProduct
		db.DB.NewSelect().Model(&items).Scan(ctx)
		pw, _ := writer.NewParquetWriter(fw, new(CVEProductDTO), 4)
		for _, item := range items {
			pw.Write(CVEProductDTO{CVEID: item.CVEID, ProductID: item.ProductID})
		}
		pw.WriteStop()
	case "cve_references":
		type CVERef struct {
			ID     string
			CVEID  string `bun:"cve_id"`
			URL    string
			Source *string
			Tags   []string `bun:",array"`
		}
		var items []CVERef
		db.DB.NewSelect().Table("cve_references").Scan(ctx, &items)
		pw, _ := writer.NewParquetWriter(fw, new(CVEReferenceDTO), 4)
		for _, item := range items {
			src := ""
			if item.Source != nil {
				src = *item.Source
			}
			tags := ""
			if len(item.Tags) > 0 {
				tags = fmt.Sprintf("[%s]", item.Tags[0])
				for i := 1; i < len(item.Tags); i++ {
					tags = tags[:len(tags)-1] + "," + item.Tags[i] + "]"
				}
			}
			pw.Write(CVEReferenceDTO{ID: item.ID, CVEID: item.CVEID, URL: item.URL, Source: src, Tags: tags})
		}
		pw.WriteStop()
	case "cve_weaknesses":
		type CVEWeak struct {
			ID       string
			CVEID    string `bun:"cve_id"`
			Weakness string
		}
		var items []CVEWeak
		db.DB.NewSelect().Table("cve_weaknesses").Scan(ctx, &items)
		pw, _ := writer.NewParquetWriter(fw, new(CVEWeaknessDTO), 4)
		for _, item := range items {
			pw.Write(CVEWeaknessDTO{ID: item.ID, CVEID: item.CVEID, Weakness: item.Weakness})
		}
		pw.WriteStop()
	case "roles":
		type Role struct {
			ID   string
			Name string
		}
		var items []Role
		db.DB.NewSelect().Table("roles").Scan(ctx, &items)
		pw, _ := writer.NewParquetWriter(fw, new(RoleDTO), 4)
		for _, item := range items {
			pw.Write(RoleDTO{ID: item.ID, Name: item.Name})
		}
		pw.WriteStop()
	case "user_roles":
		type UserRole struct {
			UserID string `bun:"user_id"`
			RoleID string `bun:"role_id"`
		}
		var items []UserRole
		db.DB.NewSelect().Table("user_roles").Scan(ctx, &items)
		pw, _ := writer.NewParquetWriter(fw, new(UserRoleDTO), 4)
		for _, item := range items {
			pw.Write(UserRoleDTO{UserID: item.UserID, RoleID: item.RoleID})
		}
		pw.WriteStop()
	case "notification_preferences":
		type NotifPref struct {
			ID      string
			UserID  string `bun:"user_id"`
			Channel string
			Enabled bool
			Config  string // JSONB stored as string
		}
		var items []NotifPref
		db.DB.NewSelect().Table("notification_preferences").Scan(ctx, &items)
		pw, _ := writer.NewParquetWriter(fw, new(NotificationPreferenceDTO), 4)
		for _, item := range items {
			pw.Write(NotificationPreferenceDTO{
				ID: item.ID, UserID: item.UserID, Channel: item.Channel,
				Enabled: item.Enabled, Config: item.Config,
			})
		}
		pw.WriteStop()
	case "oauth_accounts":
		type OAuthAcc struct {
			ID             string
			UserID         string `bun:"user_id"`
			Provider       string
			ProviderUserID string     `bun:"provider_user_id"`
			AccessToken    *string    `bun:"access_token"`
			RefreshToken   *string    `bun:"refresh_token"`
			TokenExpiry    *time.Time `bun:"token_expiry"`
		}
		var items []OAuthAcc
		db.DB.NewSelect().Table("oauth_accounts").Scan(ctx, &items)
		pw, _ := writer.NewParquetWriter(fw, new(OAuthAccountDTO), 4)
		for _, item := range items {
			accessToken := ""
			if item.AccessToken != nil {
				accessToken = *item.AccessToken
			}
			refreshToken := ""
			if item.RefreshToken != nil {
				refreshToken = *item.RefreshToken
			}
			var expiry int64
			if item.TokenExpiry != nil {
				expiry = item.TokenExpiry.UnixMilli()
			}
			pw.Write(OAuthAccountDTO{
				ID: item.ID, UserID: item.UserID, Provider: item.Provider,
				ProviderUserID: item.ProviderUserID, AccessToken: accessToken,
				RefreshToken: refreshToken, TokenExpiry: expiry,
			})
		}
		pw.WriteStop()
	case "severity_levels":
		type SevLevel struct {
			ID   int16
			Name string
		}
		var items []SevLevel
		db.DB.NewSelect().Table("severity_levels").Scan(ctx, &items)
		pw, _ := writer.NewParquetWriter(fw, new(SeverityLevelDTO), 4)
		for _, item := range items {
			pw.Write(SeverityLevelDTO{ID: int32(item.ID), Name: item.Name})
		}
		pw.WriteStop()
	case "subscription_targets":
		type SubTarget struct {
			ID             string
			SubscriptionID string `bun:"subscription_id"`
			TargetTypeID   int32  `bun:"target_type_id"`
			TargetID       string `bun:"target_id"`
		}
		var items []SubTarget
		db.DB.NewSelect().Table("subscription_targets").Scan(ctx, &items)
		pw, _ := writer.NewParquetWriter(fw, new(SubscriptionTargetDTO), 4)
		for _, item := range items {
			pw.Write(SubscriptionTargetDTO{
				ID: item.ID, SubscriptionID: item.SubscriptionID,
				TargetTypeID: item.TargetTypeID, TargetID: item.TargetID,
			})
		}
		pw.WriteStop()
	case "target_types":
		type TgtType struct {
			ID   int32
			Name string
		}
		var items []TgtType
		db.DB.NewSelect().Table("target_types").Scan(ctx, &items)
		pw, _ := writer.NewParquetWriter(fw, new(TargetTypeDTO), 4)
		for _, item := range items {
			pw.Write(TargetTypeDTO{ID: item.ID, Name: item.Name})
		}
		pw.WriteStop()
	case "daily_cve_counts":
		type DailyCount struct {
			Day   time.Time
			Count int32
		}
		var items []DailyCount
		db.DB.NewSelect().Table("daily_cve_counts").Scan(ctx, &items)
		pw, _ := writer.NewParquetWriter(fw, new(DailyCVECountDTO), 4)
		for _, item := range items {
			dayMillis := item.Day.UnixMilli()
			pw.Write(DailyCVECountDTO{Day: dayMillis, Count: item.Count})
		}
		pw.WriteStop()
	}
}

func createTarGz(outputFile string, srcDir string) {
	out, _ := os.Create(outputFile)
	defer out.Close()
	gw := gzip.NewWriter(out)
	defer gw.Close()
	tw := tar.NewWriter(gw)
	defer tw.Close()
	files, _ := os.ReadDir(srcDir)
	for _, f := range files {
		info, _ := f.Info()
		header, _ := tar.FileInfoHeader(info, "")
		header.Name = f.Name()
		tw.WriteHeader(header)
		file, _ := os.Open(filepath.Join(srcDir, f.Name()))
		io.Copy(tw, file)
		file.Close()
	}
}
