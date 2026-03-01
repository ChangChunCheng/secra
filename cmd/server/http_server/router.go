package http_server

import (
	"encoding/json"
	"html/template"
	"log"
	"math"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

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

// respondJSON helper for REST API
func (s *Server) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

// enableCORS middleware
func (s *Server) enableCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next(w, r)
	}
}

func derefString(s *string) string {
	if s == nil { return "UNKNOWN" }
	return *s
}

func (s *Server) render(w http.ResponseWriter, r *http.Request, tmpl string, data map[string]interface{}) {
	if data == nil { data = make(map[string]interface{}) }
	if user, ok := s.getUserFromSession(r); ok { data["User"] = user } else { data["User"] = nil }

	t, err := template.New("layout.html").Funcs(template.FuncMap{
		"derefString": derefString,
		"formatDate": func(t time.Time) string { return t.UTC().Format("2006-01-02T15:04:05Z") },
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
	t.Execute(w, data)
}

func (s *Server) setupRoutes() {
	fs := http.FileServer(http.Dir("web/static"))
	s.mux.Handle("/static/", http.StripPrefix("/static/", fs))

	// Health check
	s.mux.HandleFunc("/health", s.handleHealth)

	// API Routes (v1)
	s.mux.HandleFunc("/api/v1/cves", s.enableCORS(s.apiCVEList))
	s.mux.HandleFunc("/api/v1/vendors", s.enableCORS(s.apiVendorList))
	s.mux.HandleFunc("/api/v1/products", s.enableCORS(s.apiProductList))
	s.mux.HandleFunc("/api/v1/me", s.enableCORS(s.apiMe))

	// Legacy Template Routes
	s.mux.HandleFunc("/", s.handleIndex)
	s.mux.HandleFunc("/login", s.handleLogin)
	s.mux.HandleFunc("/register", s.handleRegister)
	s.mux.HandleFunc("/logout", s.handleLogout)
	s.mux.HandleFunc("/profile", s.requireAuth(s.handleProfile))
	s.mux.HandleFunc("/cves", s.handleCVEList)
	s.mux.HandleFunc("/cves/", s.handleCVEDetail)
	s.mux.HandleFunc("/vendors", s.handleVendorList)
	s.mux.HandleFunc("/products", s.handleProductList)
	s.mux.HandleFunc("/my/dashboard", s.requireAuth(s.handleMyDashboard))
	s.mux.HandleFunc("/subscribe", s.requireAuth(s.handleSubscribe))
	s.mux.HandleFunc("/unsubscribe", s.requireAuth(s.handleUnsubscribe))
	s.mux.HandleFunc("/my/subscriptions/threshold", s.requireAuth(s.handleUpdateSubscriptionThreshold))
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

// --- API Handlers ---

func (s *Server) apiCVEList(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	search := q.Get("q")
	page, _ := strconv.Atoi(q.Get("page"))
	if page < 1 { page = 1 }
	limit := 50
	offset := (page - 1) * limit

	type cveResponse struct {
		model.CVE
		SourceName string `json:"source_name"`
		Assets     string `json:"assets"`
	}
	var results []cveResponse

	query := s.db.NewSelect().TableExpr("cves AS c").
		ColumnExpr("c.*, cs.name AS source_name").
		ColumnExpr("STRING_AGG(DISTINCT v.name || ':' || p.name, ', ') AS assets").
		Join("LEFT JOIN cve_sources cs ON cs.id = c.source_id").
		Join("LEFT JOIN cve_products cp ON cp.cve_id = c.id").
		Join("LEFT JOIN products p ON p.id = cp.product_id").
		Join("LEFT JOIN vendors v ON v.id = p.vendor_id")

	if search != "" {
		query.Where("c.source_uid ILIKE ? OR c.title ILIKE ?", "%"+search+"%", "%"+search+"%")
	}

	count, _ := query.Group("c.id", "cs.name").Count(r.Context())
	_ = query.Order("c.published_at DESC", "c.source_uid ASC").Limit(limit).Offset(offset).Scan(r.Context(), &results)

	s.respondJSON(w, http.StatusOK, map[string]interface{}{
		"data":        results,
		"total":       count,
		"page":        page,
		"total_pages": int(math.Ceil(float64(count) / float64(limit))),
	})
}

func (s *Server) apiVendorList(w http.ResponseWriter, r *http.Request) {
	var results []model.Vendor
	s.db.NewSelect().Model(&results).Order("name ASC").Limit(100).Scan(r.Context())
	s.respondJSON(w, http.StatusOK, results)
}

func (s *Server) apiProductList(w http.ResponseWriter, r *http.Request) {
	var results []model.Product
	s.db.NewSelect().Model(&results).Order("name ASC").Limit(100).Scan(r.Context())
	s.respondJSON(w, http.StatusOK, results)
}

func (s *Server) apiMe(w http.ResponseWriter, r *http.Request) {
	user, ok := s.getUserFromSession(r)
	if !ok {
		s.respondJSON(w, http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
		return
	}
	s.respondJSON(w, http.StatusOK, user)
}

// --- Legacy Helpers ---

func (s *Server) getUserFromSession(r *http.Request) (*model.User, bool) {
	cookie, err := r.Cookie("session_token")
	if err != nil { return nil, false }
	user, err := s.userSvc.GetProfile(r.Context(), cookie.Value)
	if err != nil { return nil, false }
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
	if r.Method == http.MethodGet { s.render(w, r, "auth/login.html", nil); return }
	username, password := r.FormValue("username"), r.FormValue("password")
	token, _, err := s.userSvc.Login(r.Context(), username, password)
	if err != nil {
		s.render(w, r, "auth/login.html", map[string]interface{}{"Error": "Invalid credentials"})
		return
	}
	http.SetCookie(w, &http.Cookie{Name: "session_token", Value: token, Expires: time.Now().Add(24 * time.Hour), HttpOnly: true, Path: "/"})
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet { s.render(w, r, "auth/register.html", nil); return }
	username, email, password := r.FormValue("username"), r.FormValue("email"), r.FormValue("password")
	_, err := s.userSvc.Register(r.Context(), username, email, password, password)
	if err != nil { s.render(w, r, "auth/register.html", map[string]interface{}{"Error": err.Error()}); return }
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{Name: "session_token", Value: "", Expires: time.Now().Add(-1 * time.Hour), HttpOnly: true, Path: "/"})
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" { http.NotFound(w, r); return }
	totalCVEs, _ := s.db.NewSelect().Model((*model.CVE)(nil)).Count(r.Context())
	totalVendors, _ := s.db.NewSelect().Model((*model.Vendor)(nil)).Count(r.Context())
	totalProducts, _ := s.db.NewSelect().Model((*model.Product)(nil)).Count(r.Context())
	s.render(w, r, "index.html", map[string]interface{}{"TotalCVEs": totalCVEs, "TotalVendors": totalVendors, "TotalProducts": totalProducts})
}

func (s *Server) handleProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet { s.render(w, r, "user/profile.html", nil); return }
	cookie, _ := r.Cookie("session_token")
	email, password := r.FormValue("email"), r.FormValue("password")
	frequency, timezone := r.FormValue("notification_frequency"), r.FormValue("timezone")
	s.userSvc.UpdateProfile(r.Context(), cookie.Value, email, password, password, frequency, timezone)
	http.Redirect(w, r, "/profile", http.StatusSeeOther)
}

func (s *Server) handleCVEList(w http.ResponseWriter, r *http.Request) { s.render(w, r, "cve/list.html", nil) }
func (s *Server) handleCVEDetail(w http.ResponseWriter, r *http.Request) {
	id := filepath.Base(r.URL.Path)
	cve, _ := s.cveSvc.Get(r.Context(), id)
	s.render(w, r, "cve/detail.html", map[string]interface{}{"CVE": cve})
}
func (s *Server) handleVendorList(w http.ResponseWriter, r *http.Request) { s.render(w, r, "vendor/list.html", nil) }
func (s *Server) handleProductList(w http.ResponseWriter, r *http.Request) { s.render(w, r, "product/list.html", nil) }
func (s *Server) handleMyDashboard(w http.ResponseWriter, r *http.Request) {
	user, _ := s.getUserFromSession(r)
	var vendorSubs []map[string]interface{}
	s.db.NewSelect().TableExpr("subscriptions AS s").Join("JOIN subscription_targets st ON st.subscription_id = s.id").Join("JOIN vendors v ON v.id = st.target_id").Where("s.user_id = ? AND st.target_type_id = 2", user.ID).Scan(r.Context(), &vendorSubs)
	s.render(w, r, "dashboard/my.html", map[string]interface{}{"VendorSubs": vendorSubs})
}
func (s *Server) handleSubscribe(w http.ResponseWriter, r *http.Request) {
	user, _ := s.getUserFromSession(r)
	s.subscriptionSvc.CreateSubscription(r.Context(), user.ID.String(), []service.SubscriptionTarget{{TargetType: r.FormValue("target_type"), TargetID: r.FormValue("target_id")}}, "MEDIUM")
	http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther)
}
func (s *Server) handleUnsubscribe(w http.ResponseWriter, r *http.Request) {
	user, _ := s.getUserFromSession(r)
	s.subscriptionSvc.DeleteSubscription(r.Context(), user.ID.String(), r.FormValue("id"))
	http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther)
}
func (s *Server) handleUpdateSubscriptionThreshold(w http.ResponseWriter, r *http.Request) {
	user, _ := s.getUserFromSession(r)
	s.subscriptionSvc.UpdateThreshold(r.Context(), user.ID.String(), r.FormValue("id"), r.FormValue("threshold"))
	http.Redirect(w, r, "/my/dashboard", http.StatusSeeOther)
}
func (s *Server) handleAdminUsers(w http.ResponseWriter, r *http.Request) {
	var users []model.User
	s.db.NewSelect().Model(&users).Order("username ASC").Scan(r.Context())
	s.render(w, r, "user/list.html", map[string]interface{}{"Users": users})
}
func (s *Server) handleUpdateUserRole(w http.ResponseWriter, r *http.Request) {
	s.db.NewUpdate().Table("users").Set("role = ?", r.FormValue("role")).Where("id = ?", r.FormValue("user_id")).Exec(r.Context())
	http.Redirect(w, r, "/admin/users", http.StatusSeeOther)
}
