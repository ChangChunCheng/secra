package http_server

import (
	"html/template"
	"log"
	"math"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"

	"gitlab.com/jacky850509/secra/internal/config"
	"gitlab.com/jacky850509/secra/internal/model"
	"gitlab.com/jacky850509/secra/internal/repo"
	"gitlab.com/jacky850509/secra/internal/service"
	"gitlab.com/jacky850509/secra/internal/util"
)

type Server struct {
	mux *http.ServeMux
	db  *bun.DB
	cfg *config.AppConfig

	cveSvc          service.CveServicer
	userSvc         service.UserServicer
	subscriptionSvc service.SubscriptionServicer
}

func NewServer(db *bun.DB) *Server {
	cfg := config.Load()
	cveRepo := repo.NewCVERepo(db)
	userRepo := repo.NewUserRepository(db)
	subRepo := repo.NewSubscriptionRepository(db)

	s := &Server{
		mux:             http.NewServeMux(),
		db:              db,
		cfg:             cfg,
		cveSvc:          service.NewCveService(cveRepo),
		userSvc:         service.NewUserService(userRepo),
		subscriptionSvc: service.NewSubscriptionService(subRepo),
	}

	s.setupRoutes()
	return s
}

func derefString(s *string) string {
	if s == nil {
		return "UNKNOWN"
	}
	return *s
}

func (s *Server) render(w http.ResponseWriter, r *http.Request, tmpl string, data map[string]interface{}) {
	if data == nil {
		data = make(map[string]interface{})
	}
	if user, ok := s.getUserFromSession(r); ok {
		data["User"] = user
	} else {
		data["User"] = nil
	}

	t, err := template.New("layout.html").Funcs(template.FuncMap{
		"derefString": derefString,
		"formatDate": func(t time.Time) string {
			return t.UTC().Format("2006-01-02T15:04:05Z")
		},
		"add": func(a, b int) int { return a + b },
		"sub": func(a, b int) int { return a - b },
	}).ParseFiles(
		filepath.Join("web", "templates", "layout.html"),
		filepath.Join("web", "templates", tmpl),
	)
	if err != nil {
		log.Printf("Template error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = t.Execute(w, data)
	if err != nil {
		log.Printf("Execute error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) setupRoutes() {
	fs := http.FileServer(http.Dir("web/static"))
	s.mux.Handle("/static/", http.StripPrefix("/static/", fs))

	s.mux.HandleFunc("/health", s.handleHealth)
	s.mux.HandleFunc("/", s.handleIndex)
	s.mux.HandleFunc("/login", s.handleLogin)
	s.mux.HandleFunc("/register", s.handleRegister)
	s.mux.HandleFunc("/logout", s.handleLogout)
	s.mux.HandleFunc("/profile", s.requireAuth(s.handleProfile))
	
	s.mux.HandleFunc("/cves", s.handleCVEList)
	s.mux.HandleFunc("/cves/", s.handleCVEDetail)
	s.mux.HandleFunc("/cves/new", s.requireAuth(s.handleCVENew))

	s.mux.HandleFunc("/vendors", s.handleVendorList)
	s.mux.HandleFunc("/products", s.handleProductList)

	s.mux.HandleFunc("/my/dashboard", s.requireAuth(s.handleMyDashboard))
	s.mux.HandleFunc("/subscribe", s.requireAuth(s.handleSubscribe))
	s.mux.HandleFunc("/unsubscribe", s.requireAuth(s.handleUnsubscribe))
	s.mux.HandleFunc("/my/subscriptions/threshold", s.requireAuth(s.handleUpdateSubscriptionThreshold))

	// Admin Routes
	s.mux.HandleFunc("/admin/users", s.requireAdmin(s.handleAdminUsers))
	s.mux.HandleFunc("/admin/users/role", s.requireAdmin(s.handleUpdateUserRole))
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (s *Server) getUserFromSession(r *http.Request) (*model.User, bool) {
	cookie, err := r.Cookie("session_token")
	if err != nil {
		return nil, false
	}
	user, err := s.userSvc.GetProfile(r.Context(), cookie.Value)
	if err != nil {
		return nil, false
	}
	return user, true
}

func (s *Server) requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := s.getUserFromSession(r); !ok {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		next(w, r)
	}
}

func (s *Server) requireAdmin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := s.getUserFromSession(r)
		if !ok || user.Role != "admin" {
			http.Error(w, "Forbidden: Admin access required", http.StatusForbidden)
			return
		}
		next(w, r)
	}
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		s.render(w, r, "auth/login.html", nil)
		return
	}
	username := r.FormValue("username")
	password := r.FormValue("password")
	
	log.Printf("🔑 Attempting login for user: %s", username)
	token, _, err := s.userSvc.Login(r.Context(), username, password)
	if err != nil {
		log.Printf("❌ Login failed for %s: %v", username, err)
		s.render(w, r, "auth/login.html", map[string]interface{}{"Error": "Invalid credentials"})
		return
	}
	
	log.Printf("✅ Login successful for %s, setting cookie", username)
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    token,
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
		Path:     "/",
	})
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		s.render(w, r, "auth/register.html", nil)
		return
	}
	username := r.FormValue("username")
	email := r.FormValue("email")
	password := r.FormValue("password")
	_, err := s.userSvc.Register(r.Context(), username, email, password, password)
	if err != nil {
		s.render(w, r, "auth/register.html", map[string]interface{}{"Error": err.Error()})
		return
	}
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Expires:  time.Now().Add(-1 * time.Hour),
		HttpOnly: true,
		Path:     "/",
	})
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	timeRange := r.URL.Query().Get("range")
	if timeRange == "" { timeRange = "1y" }

	now := time.Now().UTC()
	end := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	var start time.Time
	
	switch timeRange {
	case "1w": start = end.AddDate(0, 0, -7)
	case "1m": start = end.AddDate(0, -1, 0)
	case "1y": start = end.AddDate(-1, 0, 0)
	case "5y": start = end.AddDate(-5, 0, 0)
	case "custom":
		start, _ = time.Parse("2006-01-02", r.URL.Query().Get("start"))
		customEnd, _ := time.Parse("2006-01-02", r.URL.Query().Get("end"))
		if !customEnd.IsZero() { end = customEnd }
	default: start = end.AddDate(-1, 0, 0)
	}

	type ChartRow struct {
		Period time.Time `bun:"period"`
		Count  int       `bun:"count"`
	}
	var rows []ChartRow
	dateMap := make(map[string]int)
	var labels []string
	var data []int

	if timeRange == "5y" {
		err := s.db.NewSelect().Table("daily_cve_counts").
			ColumnExpr("DATE_TRUNC('month', day) as period, sum(count) as count").
			Where("day >= ? AND day <= ?", start.Format("2006-01-02"), end.Format("2006-01-02")).
			Group("period").Order("period ASC").Scan(r.Context(), &rows)
		
		if err == nil {
			for _, r := range rows { dateMap[r.Period.UTC().Format("2006-01-01")] = r.Count }
			for d := start; !d.After(end); {
				firstOfMonth := time.Date(d.Year(), d.Month(), 1, 0, 0, 0, 0, time.UTC)
				ks := firstOfMonth.Format("2006-01-01")
				labels = append(labels, ks)
				data = append(data, dateMap[ks])
				d = firstOfMonth.AddDate(0, 1, 0)
			}
		}
	} else if timeRange == "1y" {
		alignToMonday := func(t time.Time) time.Time {
			wd := int(t.Weekday())
			if wd == 0 { wd = 7 }
			return t.AddDate(0, 0, -(wd - 1))
		}
		queryStart := alignToMonday(start)
		err := s.db.NewSelect().Table("daily_cve_counts").
			ColumnExpr("DATE_TRUNC('week', day) as period, sum(count) as count").
			Where("day >= ? AND day <= ?", queryStart.Format("2006-01-02"), end.Format("2006-01-02")).
			Group("period").Order("period ASC").Scan(r.Context(), &rows)
		
		if err == nil {
			for _, r := range rows { dateMap[r.Period.Format("2006-01-02")] = r.Count }
			for d := queryStart; !d.After(end); d = d.AddDate(0, 0, 7) {
				ks := d.Format("2006-01-02")
				labels = append(labels, ks)
				data = append(data, dateMap[ks])
			}
		}
	} else {
		err := s.db.NewSelect().Table("daily_cve_counts").
			ColumnExpr("day as period, count").
			Where("day >= ? AND day <= ?", start.Format("2006-01-02"), end.Format("2006-01-02")).
			Order("period ASC").Scan(r.Context(), &rows)
		
		if err == nil {
			for _, r := range rows { dateMap[r.Period.Format("2006-01-02")] = r.Count }
			for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
				ks := d.Format("2006-01-02")
				labels = append(labels, ks)
				data = append(data, dateMap[ks])
			}
		}
	}

	type cveWithAssets struct {
		model.CVE
		SourceName string `bun:"source_name"`
		Assets     string `bun:"assets"`
	}
	var recentCVEs []cveWithAssets
	s.db.NewSelect().TableExpr("cves AS c").
		ColumnExpr("c.*, cs.name AS source_name").
		ColumnExpr("STRING_AGG(DISTINCT v.name || ':' || p.name, ', ') AS assets").
		Join("LEFT JOIN cve_sources cs ON cs.id = c.source_id").
		Join("LEFT JOIN cve_products cp ON cp.cve_id = c.id").
		Join("LEFT JOIN products p ON p.id = cp.product_id").
		Join("LEFT JOIN vendors v ON v.id = p.vendor_id").
		Group("c.id", "cs.name").Order("c.published_at DESC", "c.source_uid ASC").Limit(10).Scan(r.Context(), &recentCVEs)

	totalCVEs, _ := s.db.NewSelect().Model((*model.CVE)(nil)).Count(r.Context())
	totalVendors, _ := s.db.NewSelect().Model((*model.Vendor)(nil)).Count(r.Context())
	totalProducts, _ := s.db.NewSelect().Model((*model.Product)(nil)).Count(r.Context())

	s.render(w, r, "index.html", map[string]interface{}{
		"TotalCVEs":     totalCVEs,
		"TotalVendors":  totalVendors,
		"TotalProducts": totalProducts,
		"RecentCVEs":    recentCVEs,
		"ChartLabels":   labels,
		"ChartData":     data,
		"ActiveRange":   timeRange,
		"StartStr":      start.Format("2006-01-02"),
		"EndStr":        end.Format("2006-01-02"),
	})
}

func (s *Server) handleProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		s.render(w, r, "user/profile.html", nil)
		return
	}
	cookie, _ := r.Cookie("session_token")
	email := r.FormValue("email")
	password := r.FormValue("password")
	frequency := r.FormValue("notification_frequency")
	timezone := r.FormValue("timezone")

	_, err := s.userSvc.UpdateProfile(r.Context(), cookie.Value, email, password, password, frequency, timezone)
	if err != nil {
		s.render(w, r, "user/profile.html", map[string]interface{}{"Error": err.Error()})
		return
	}
	http.Redirect(w, r, "/profile", http.StatusSeeOther)
}

func (s *Server) handleCVEList(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	search := q.Get("q")
	vendorName := q.Get("vendor")
	productName := q.Get("product")
	startDate := q.Get("start_date")
	endDate := q.Get("end_date")
	page, _ := strconv.Atoi(q.Get("page"))
	if page < 1 { page = 1 }
	limit := 50
	offset := (page - 1) * limit

	type cveWithAssets struct {
		model.CVE
		SourceName string `bun:"source_name"`
		Assets     string `bun:"assets"`
	}
	var results []cveWithAssets

	query := s.db.NewSelect().
		TableExpr("cves AS c").
		ColumnExpr("c.*, cs.name AS source_name").
		ColumnExpr("STRING_AGG(DISTINCT v.name || ':' || p.name, ', ') AS assets").
		Join("LEFT JOIN cve_sources cs ON cs.id = c.source_id").
		Join("LEFT JOIN cve_products cp ON cp.cve_id = c.id").
		Join("LEFT JOIN products p ON p.id = cp.product_id").
		Join("LEFT JOIN vendors v ON v.id = p.vendor_id")

	if search != "" {
		query.Where("c.source_uid ILIKE ? OR c.title ILIKE ? OR c.description ILIKE ?", "%"+search+"%", "%"+search+"%", "%"+search+"%")
	}
	if vendorName != "" {
		query.Where("v.name ILIKE ?", "%"+vendorName+"%")
	}
	if productName != "" {
		query.Where("p.name ILIKE ?", "%"+productName+"%")
	}
	if startDate != "" {
		if t, err := time.Parse("2006-01-02", startDate); err == nil {
			query.Where("c.published_at >= ?", t)
		}
	}
	if endDate != "" {
		if t, err := time.Parse("2006-01-02", endDate); err == nil {
			query.Where("c.published_at <= ?", t.Add(24*time.Hour))
		}
	}

	count, _ := query.Group("c.id", "cs.name").Count(r.Context())
	_ = query.Order("c.published_at DESC", "c.source_uid ASC").Limit(limit).Offset(offset).Scan(r.Context(), &results)
	
	totalPages := int(math.Ceil(float64(count) / float64(limit)))
	if totalPages == 0 { totalPages = 1 }

	var pages []int
	startPage := page - 2
	if startPage < 1 { startPage = 1 }
	endPage := page + 2
	if endPage > totalPages { endPage = totalPages }
	for i := startPage; i <= endPage; i++ {
		pages = append(pages, i)
	}

	s.render(w, r, "cve/list.html", map[string]interface{}{
		"CVEs":       results,
		"Search":     search,
		"Vendor":     vendorName,
		"Product":    productName,
		"StartDate":  startDate,
		"EndDate":    endDate,
		"Page":       page,
		"TotalCount": count,
		"TotalPages": totalPages,
		"Pages":      pages,
		"HasNext":    page < totalPages,
		"HasPrev":    page > 1,
	})
}

func (s *Server) handleCVEDetail(w http.ResponseWriter, r *http.Request) {
	id := filepath.Base(r.URL.Path)
	cve, err := s.cveSvc.Get(r.Context(), id)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	source := new(model.CVESource)
	s.db.NewSelect().Model(source).Where("id = ?", cve.SourceID).Scan(r.Context())
	type productWithVendor struct {
		model.Product
		VendorName string
	}
	var products []productWithVendor
	s.db.NewSelect().Model((*model.Product)(nil)).
		ColumnExpr("product.*, v.name AS vendor_name").
		Join("JOIN vendors v ON v.id = product.vendor_id").
		Where("cp.cve_id = ?", id).
		Join("JOIN cve_products cp ON cp.product_id = product.id").
		Order("vendor_name ASC", "name ASC").Scan(r.Context(), &products)
	s.render(w, r, "cve/detail.html", map[string]interface{}{
		"CVE":        cve,
		"SourceName": source.Name,
		"Products":   products,
	})
}

func (s *Server) handleCVENew(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		s.render(w, r, "cve/new.html", nil)
		return
	}
	sourceUID := r.FormValue("source_uid")
	title := r.FormValue("title")
	description := r.FormValue("description")
	severity := r.FormValue("severity")
	cvssScore, _ := strconv.ParseFloat(r.FormValue("cvss_score"), 64)
	source := new(model.CVESource)
	err := s.db.NewSelect().Model(source).Where("name = ?", "Manual").Scan(r.Context())
	if err != nil {
		source = &model.CVESource{ID: uuid.New().String(), Name: "Manual", Type: "manual", URL: "local", Enabled: true}
		s.db.NewInsert().Model(source).Exec(r.Context())
	}
	cve := &model.CVE{
		ID: util.CVEID(sourceUID), SourceID: source.ID, SourceUID: sourceUID, Title: title,
		Description: description, Severity: &severity, CVSSScore: &cvssScore,
		Status: "active", PublishedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC(),
	}
	s.db.NewInsert().Model(cve).Exec(r.Context())
	http.Redirect(w, r, "/cves/"+cve.ID, http.StatusSeeOther)
}

func (s *Server) handleVendorList(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	search := q.Get("q")
	page, _ := strconv.Atoi(q.Get("page"))
	if page < 1 { page = 1 }
	limit := 50
	offset := (page - 1) * limit

	user, isLogged := s.getUserFromSession(r)
	userID := uuid.Nil
	if isLogged { userID = user.ID }

	type vendorWithStats struct {
		model.Vendor
		ProductCount   int    `bun:"product_count"`
		SubscriptionID string `bun:"subscription_id"`
	}
	var results []vendorWithStats

	query := s.db.NewSelect().
		TableExpr("vendors AS v").
		ColumnExpr("v.*, count(p.id) as product_count").
		ColumnExpr("s.id AS subscription_id").
		Join("LEFT JOIN products p ON p.vendor_id = v.id").
		Join("LEFT JOIN subscription_targets st ON st.target_id = v.id AND st.target_type_id = 2").
		Join("LEFT JOIN subscriptions s ON s.id = st.subscription_id AND s.user_id = ?", userID)

	if search != "" {
		query.Where("v.name ILIKE ?", "%"+search+"%")
	}

	count, _ := query.Group("v.id", "s.id").Count(r.Context())
	_ = query.Order("v.name ASC", "v.id ASC").Limit(limit).Offset(offset).Scan(r.Context(), &results)

	totalPages := int(math.Ceil(float64(count) / float64(limit)))
	if totalPages == 0 { totalPages = 1 }
	var pages []int
	startPage := page - 2
	if startPage < 1 { startPage = 1 }
	endPage := page + 2
	if endPage > totalPages { endPage = totalPages }
	for i := startPage; i <= endPage; i++ {
		pages = append(pages, i)
	}

	s.render(w, r, "vendor/list.html", map[string]interface{}{
		"Vendors":    results,
		"Search":     search,
		"Page":       page,
		"TotalCount": count,
		"TotalPages": totalPages,
		"Pages":      pages,
		"HasNext":    page < totalPages,
		"HasPrev":    page > 1,
	})
}

func (s *Server) handleProductList(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	search := q.Get("q")
	vendorName := q.Get("vendor")
	page, _ := strconv.Atoi(q.Get("page"))
	if page < 1 { page = 1 }
	limit := 50
	offset := (page - 1) * limit

	user, isLogged := s.getUserFromSession(r)
	userID := uuid.Nil
	if isLogged { userID = user.ID }

	type productWithVendor struct {
		model.Product
		VendorName     string `bun:"vendor_name"`
		SubscriptionID string `bun:"subscription_id"`
	}
	var results []productWithVendor

	query := s.db.NewSelect().
		TableExpr("products AS p").
		ColumnExpr("p.*, v.name AS vendor_name").
		ColumnExpr("s.id AS subscription_id").
		Join("JOIN vendors v ON v.id = p.vendor_id").
		Join("LEFT JOIN subscription_targets st ON st.target_id = p.id AND st.target_type_id = 3").
		Join("LEFT JOIN subscriptions s ON s.id = st.subscription_id AND s.user_id = ?", userID)

	if search != "" {
		query.Where("p.name ILIKE ?", "%"+search+"%")
	}
	if vendorName != "" {
		query.Where("v.name ILIKE ?", "%"+vendorName+"%")
	}

	count, _ := query.Count(r.Context())
	_ = query.Order("vendor_name ASC", "p.name ASC", "p.id ASC").Limit(limit).Offset(offset).Scan(r.Context(), &results)

	totalPages := int(math.Ceil(float64(count) / float64(limit)))
	if totalPages == 0 { totalPages = 1 }
	var pages []int
	startPage := page - 2
	if startPage < 1 { startPage = 1 }
	endPage := page + 2
	if endPage > totalPages { endPage = totalPages }
	for i := startPage; i <= endPage; i++ {
		pages = append(pages, i)
	}

	s.render(w, r, "product/list.html", map[string]interface{}{
		"Products":   results,
		"Search":     search,
		"Vendor":     vendorName,
		"Page":       page,
		"TotalCount": count,
		"TotalPages": totalPages,
		"Pages":      pages,
		"HasNext":    page < totalPages,
		"HasPrev":    page > 1,
	})
}

func (s *Server) handleMyDashboard(w http.ResponseWriter, r *http.Request) {
	user, _ := s.getUserFromSession(r)
	type enrichedSub struct {
		ID                string
		TargetType        string
		TargetName        string
		SeverityThreshold string
	}
	vendorSubs := make([]enrichedSub, 0)
	s.db.NewSelect().TableExpr("subscriptions AS s").
		ColumnExpr("s.id, 'vendor' as target_type, v.name as target_name, s.severity_threshold").
		Join("JOIN subscription_targets st ON st.subscription_id = s.id").
		Join("JOIN vendors v ON v.id = st.target_id").
		Where("s.user_id = ? AND st.target_type_id = 2", user.ID).
		Order("target_name ASC").
		Scan(r.Context(), &vendorSubs)

	productSubs := make([]enrichedSub, 0)
	s.db.NewSelect().TableExpr("subscriptions AS s").
		ColumnExpr("s.id, 'product' as target_type, p.name as target_name, s.severity_threshold").
		Join("JOIN subscription_targets st ON st.subscription_id = s.id").
		Join("JOIN products p ON p.id = st.target_id").
		Where("s.user_id = ? AND st.target_type_id = 3", user.ID).
		Order("target_name ASC").
		Scan(r.Context(), &productSubs)

	for i := range vendorSubs {
		t, _ := strconv.Atoi(vendorSubs[i].SeverityThreshold)
		vendorSubs[i].SeverityThreshold = s.subscriptionSvc.SeverityToString(int16(t))
	}
	for i := range productSubs {
		t, _ := strconv.Atoi(productSubs[i].SeverityThreshold)
		productSubs[i].SeverityThreshold = s.subscriptionSvc.SeverityToString(int16(t))
	}

	s.render(w, r, "dashboard/my.html", map[string]interface{}{
		"VendorSubs":  vendorSubs,
		"ProductSubs": productSubs,
	})
}

func (s *Server) handleSubscribe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	user, _ := s.getUserFromSession(r)
	targetType := r.FormValue("target_type")
	targetID := r.FormValue("target_id")
	targets := []service.SubscriptionTarget{{TargetType: targetType, TargetID: targetID}}
	s.subscriptionSvc.CreateSubscription(r.Context(), user.ID.String(), targets, "MEDIUM")
	http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther)
}

func (s *Server) handleUnsubscribe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	user, _ := s.getUserFromSession(r)
	id := r.FormValue("id")
	s.subscriptionSvc.DeleteSubscription(r.Context(), user.ID.String(), id)
	http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther)
}

func (s *Server) handleUpdateSubscriptionThreshold(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	user, _ := s.getUserFromSession(r)
	subID := r.FormValue("id")
	newThreshold := r.FormValue("threshold")
	err := s.subscriptionSvc.UpdateThreshold(r.Context(), user.ID.String(), subID, newThreshold)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/my/dashboard", http.StatusSeeOther)
}

func (s *Server) handleAdminUsers(w http.ResponseWriter, r *http.Request) {
	var users []model.User
	s.db.NewSelect().Model(&users).Order("username ASC", "id ASC").Scan(r.Context())
	s.render(w, r, "user/list.html", map[string]interface{}{"Users": users})
}

func (s *Server) handleUpdateUserRole(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	userID := r.FormValue("user_id")
	newRole := r.FormValue("role")
	if newRole != "admin" && newRole != "user" {
		http.Error(w, "Invalid role", http.StatusBadRequest)
		return
	}
	_, err := s.db.NewUpdate().Table("users").Set("role = ?", newRole).Where("id = ?", userID).Exec(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/admin/users", http.StatusSeeOther)
}
