package service

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/smtp"
	"strings"

	"github.com/uptrace/bun"
	"gitlab.com/jacky850509/secra/internal/config"
	"gitlab.com/jacky850509/secra/internal/model"
)

type NotificationService interface {
	SendEmail(ctx context.Context, to, subject, body string) error
	NotifyCVEBatch(ctx context.Context, email string, cves []model.CVE) error
	ProcessBatch(ctx context.Context, cves []model.CVE) error
}

type smtpNotificationService struct {
	cfg config.SMTPConfig
	db  *bun.DB
}

func NewNotificationService(cfg config.SMTPConfig, db *bun.DB) NotificationService {
	return &smtpNotificationService{cfg: cfg, db: db}
}

func (s *smtpNotificationService) SendEmail(ctx context.Context, to, subject, body string) error {
	if s.cfg.Host == "" { return fmt.Errorf("SMTP host not configured") }
	addr := fmt.Sprintf("%s:%d", s.cfg.Host, s.cfg.Port)
	auth := smtp.PlainAuth("", s.cfg.User, s.cfg.Password, s.cfg.Host)
	msg := []byte(fmt.Sprintf("To: %s\r\nSubject: %s\r\n\r\n%s\r\n", to, subject, body))
	encryption := strings.ToUpper(s.cfg.Encryption)

	if encryption == "SSL" || encryption == "TLS" {
		tlsConfig := &tls.Config{InsecureSkipVerify: false, ServerName: s.cfg.Host}
		conn, err := tls.Dial("tcp", addr, tlsConfig)
		if err != nil { return fmt.Errorf("SSL connection failed: %v", err) }
		defer conn.Close()
		client, err := smtp.NewClient(conn, s.cfg.Host)
		if err != nil { return err }
		defer client.Quit()
		return s.executeSMTP(client, auth, to, msg)
	}

	client, err := smtp.Dial(addr)
	if err != nil { return fmt.Errorf("SMTP connection failed: %v", err) }
	defer client.Quit()
	if encryption == "STARTTLS" {
		tlsConfig := &tls.Config{InsecureSkipVerify: false, ServerName: s.cfg.Host}
		if err = client.StartTLS(tlsConfig); err != nil { return fmt.Errorf("STARTTLS handshake failed: %v", err) }
	}
	return s.executeSMTP(client, auth, to, msg)
}

func (s *smtpNotificationService) executeSMTP(client *smtp.Client, auth smtp.Auth, to string, msg []byte) error {
	if auth != nil {
		if ok, _ := client.Extension("AUTH"); ok {
			if err := client.Auth(auth); err != nil { return fmt.Errorf("SMTP authentication failed: %v", err) }
		}
	}
	if err := client.Mail(s.cfg.From); err != nil { return err }
	if err := client.Rcpt(to); err != nil { return err }
	w, err := client.Data()
	if err != nil { return err }
	_, err = w.Write(msg)
	if err != nil { return err }
	return w.Close()
}

func (s *smtpNotificationService) NotifyCVEBatch(ctx context.Context, email string, cves []model.CVE) error {
	count := len(cves)
	subject := fmt.Sprintf("[SECRA Alert] %d New Vulnerabilities Detected", count)
	body := fmt.Sprintf("Hello,\n\nWe found %d new vulnerabilities affecting your monitored assets:\n\n", count)
	for _, cve := range cves {
		body += fmt.Sprintf("- [%s] %s (Severity: %s)\n", cve.SourceUID, cve.Title, derefStr(cve.Severity))
	}
	body += "\nCheck details at SECRA Dashboard: http://localhost:8081\n"
	return s.SendEmail(ctx, email, subject, body)
}

func (s *smtpNotificationService) ProcessBatch(ctx context.Context, cves []model.CVE) error {
	if len(cves) == 0 { return nil }

	// 1. Get all CVE IDs in this batch
	var cveIDs []string
	for _, c := range cves { cveIDs = append(cveIDs, c.ID) } // ID is string

	// 2. Find all 'immediate' users who match ANY of these CVEs
	type match struct {
		Email    string `bun:"email"`
		Username string `bun:"username"`
		CVEID    string `bun:"cve_id"`
	}
	var matches []match

	err := s.db.NewSelect().
		TableExpr("users AS u").
		ColumnExpr("u.email, u.username, cp.cve_id").
		Join("JOIN subscriptions s ON s.user_id = u.id").
		Join("JOIN subscription_targets st ON st.subscription_id = s.id").
		Join("JOIN cve_products cp ON (st.target_type_id = 2 AND st.target_id = (SELECT vendor_id FROM products WHERE id = cp.product_id)) OR (st.target_type_id = 3 AND st.target_id = cp.product_id)").
		Where("u.notification_frequency = 'immediate' AND cp.cve_id IN (?)", bun.In(cveIDs)).
		Scan(ctx, &matches)

	if err != nil { return err }

	// 3. Group matches by User Email
	userMap := make(map[string][]model.CVE)
	cveLookup := make(map[string]model.CVE)
	for _, c := range cves { cveLookup[c.ID] = c }

	for _, m := range matches {
		userMap[m.Email] = append(userMap[m.Email], cveLookup[m.CVEID])
	}

	// 4. Send aggregated emails
	for email, userCVEs := range userMap {
		log.Printf("📧 Aggregating %d CVEs for %s", len(userCVEs), email)
		s.NotifyCVEBatch(ctx, email, userCVEs)
	}

	return nil
}

func derefStr(s *string) string {
	if s == nil { return "UNKNOWN" }
	return *s
}
