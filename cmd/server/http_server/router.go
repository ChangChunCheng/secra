package http_server

import (
	"html/template"
	"log"
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

	s.mux.HandleFunc("/admin/users", s.requireAdmin(s.handleAdminUsers))
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
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
			http.Error(w, "Forbidden", http.StatusForbidden)
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
	token, _, err := s.userSvc.Login(r.Context(), username, password)
	if err != nil {
		s.render(w, r, "auth/login.html", map[string]interface{}{"Error": "Invalid credentials"})
		return
	}
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
	cves, err := s.cveSvc.List(r.Context(), 10, 0)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	var labels []string
	var data []int
	now := time.Now()
	for i := 6; i >= 0; i-- {
		day := now.AddDate(0, 0, -i)
		labels = append(labels, day.Format("01-02"))
		count, _ := s.db.NewSelect().Model((*model.CVE)(nil)).
			Where("published_at >= ? AND published_at < ?", 
				time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, time.UTC),
				time.Date(day.Year(), day.Month(), day.Day(), 23, 59, 59, 0, time.UTC)).
			Count(r.Context())
		data = append(data, int(count))
	}
	totalCVEs, _ := s.db.NewSelect().Model((*model.CVE)(nil)).Count(r.Context())
	totalVendors, _ := s.db.NewSelect().Model((*model.Vendor)(nil)).Count(r.Context())
	totalProducts, _ := s.db.NewSelect().Model((*model.Product)(nil)).Count(r.Context())

	s.render(w, r, "index.html", map[string]interface{}{
		"TotalCVEs":     totalCVEs,
		"TotalVendors":  totalVendors,
		"TotalProducts": totalProducts,
		"RecentCVEs":    cves,
		"ChartLabels":   labels,
		"ChartData":     data,
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
	_, err := s.userSvc.UpdateProfile(r.Context(), cookie.Value, email, password, password)
	if err != nil {
		s.render(w, r, "user/profile.html", map[string]interface{}{"Error": err.Error()})
		return
	}
	http.Redirect(w, r, "/profile", http.StatusSeeOther)
}

func (s *Server) handleCVEList(w http.ResponseWriter, r *http.Request) {
	cves, _ := s.cveSvc.List(r.Context(), 50, 0)
	var sources []model.CVESource
	s.db.NewSelect().Model(&sources).Scan(r.Context())
	sourceMap := make(map[string]string)
	for _, src := range sources {
		sourceMap[src.ID] = src.Name
	}
	s.render(w, r, "cve/list.html", map[string]interface{}{
		"CVEs":      cves,
		"SourceMap": sourceMap,
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
		Scan(r.Context(), &products)
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
		ID: uuid.New().String(), SourceID: source.ID, SourceUID: sourceUID, Title: title,
		Description: description, Severity: &severity, CVSSScore: &cvssScore,
		Status: "active", PublishedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC(),
	}
	s.db.NewInsert().Model(cve).Exec(r.Context())
	http.Redirect(w, r, "/cves/"+cve.ID, http.StatusSeeOther)
}

func (s *Server) handleVendorList(w http.ResponseWriter, r *http.Request) {
	var vendors []model.Vendor
	s.db.NewSelect().Model(&vendors).Order("name ASC").Limit(100).Scan(r.Context())
	s.render(w, r, "vendor/list.html", map[string]interface{}{"Vendors": vendors})
}

func (s *Server) handleProductList(w http.ResponseWriter, r *http.Request) {
	type productWithVendor struct {
		model.Product
		VendorName string
	}
	var products []productWithVendor
	s.db.NewSelect().Model((*model.Product)(nil)).
		ColumnExpr("product.*, v.name AS vendor_name").
		Join("JOIN vendors v ON v.id = product.vendor_id").
		Order("product.name ASC").Limit(100).
		Scan(r.Context(), &products)
	s.render(w, r, "product/list.html", map[string]interface{}{"Products": products})
}

func (s *Server) handleMyDashboard(w http.ResponseWriter, r *http.Request) {
	user, _ := s.getUserFromSession(r)
	type enrichedSub struct {
		ID                string
		TargetType        string
		TargetName        string
		SeverityThreshold string
	}
	var subs []enrichedSub
	var vendorSubs []enrichedSub
	s.db.NewSelect().Table("subscriptions").
		ColumnExpr("subscriptions.id, 'vendor' as target_type, v.name as target_name, subscriptions.severity_threshold").
		Join("JOIN vendors v ON v.id::text = (subscriptions.targets->0->>'target_id')").
		Where("subscriptions.user_id = ?", user.ID).
		Scan(r.Context(), &vendorSubs)
	var productSubs []enrichedSub
	s.db.NewSelect().Table("subscriptions").
		ColumnExpr("subscriptions.id, 'product' as target_type, p.name as target_name, subscriptions.severity_threshold").
		Join("JOIN products p ON p.id::text = (subscriptions.targets->0->>'target_id')").
		Where("subscriptions.user_id = ?", user.ID).
		Scan(r.Context(), &productSubs)
	subs = append(vendorSubs, productSubs...)
	s.render(w, r, "dashboard/my.html", map[string]interface{}{
		"EnrichedSubs": subs,
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

func (s *Server) handleAdminUsers(w http.ResponseWriter, r *http.Request) {
	var users []model.User
	s.db.NewSelect().Model(&users).Order("username ASC").Scan(r.Context())
	s.render(w, r, "user/list.html", map[string]interface{}{"Users": users})
}
