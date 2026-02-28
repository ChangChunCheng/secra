package http_server

import (
	"html/template"
	"net/http"
	"path/filepath"
	"time"

	"github.com/golang-jwt/jwt/v4"
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

	// Services
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

// derefString and other helpers...
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
	// Inject current user into all templates
	if user, ok := s.getUserFromSession(r); ok {
		data["User"] = user
	}

	t, err := template.New("layout.html").Funcs(template.FuncMap{
		"derefString": derefString,
		"formatDate": func(t time.Time) string {
			return t.Format("2006-01-02")
		},
	}).ParseFiles(
		filepath.Join("web", "templates", "layout.html"),
		filepath.Join("web", "templates", tmpl),
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = t.Execute(w, data)
	if err != nil {
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
	
	// CVE Routes
	s.mux.HandleFunc("/cves", s.handleCVEList)
	s.mux.HandleFunc("/cves/", s.handleCVEDetail) // Handles /cves/:id
	s.mux.HandleFunc("/cves/new", s.requireAuth(s.handleCVENew))

	// Subscription & User Dashboard
	s.mux.HandleFunc("/my/dashboard", s.requireAuth(s.handleMyDashboard))
	s.mux.HandleFunc("/subscribe", s.requireAuth(s.handleSubscribe))

	// Admin
	s.mux.HandleFunc("/admin/users", s.requireAdmin(s.handleAdminUsers))
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

// Auth Helpers
func (s *Server) getUserFromSession(r *http.Request) (*model.User, bool) {
	cookie, err := r.Cookie("session_token")
	if err != nil {
		return nil, false
	}

	// Validate JWT
	token, err := jwt.Parse(cookie.Value, func(token *jwt.Token) (interface{}, error) {
		return s.cfg.JWTConfig.Secret, nil
	})
	if err != nil || !token.Valid {
		return nil, false
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, false
	}

	// For simplicity, we just return a partial user object from claims
	// Or we could fetch from DB if needed
	sub, ok := claims["sub"].(string)
	if !ok {
		return nil, false
	}
	
	// Try to get user from DB
	user := new(model.User)
	err = s.db.NewSelect().Model(user).Where("username = ?", sub).Scan(r.Context())
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

// Handlers (Implementation of basic auth for now)
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
		
		// Real count query for daily CVEs
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

// Placeholders for other handlers
func (s *Server) handleProfile(w http.ResponseWriter, r *http.Request) { s.render(w, r, "user/profile.html", nil) }
func (s *Server) handleCVEList(w http.ResponseWriter, r *http.Request) {
	cves, _ := s.cveSvc.List(r.Context(), 50, 0)
	s.render(w, r, "cve/list.html", map[string]interface{}{"CVEs": cves})
}
func (s *Server) handleCVEDetail(w http.ResponseWriter, r *http.Request) {
	id := filepath.Base(r.URL.Path)
	cve, _ := s.cveSvc.Get(r.Context(), id)
	
	// Get products for this CVE
	var products []model.Product
	s.db.NewSelect().Model(&products).
		Join("JOIN cve_products ON cve_products.product_id = product.id").
		Where("cve_products.cve_id = ?", id).
		Scan(r.Context())

	s.render(w, r, "cve/detail.html", map[string]interface{}{
		"CVE": cve,
		"Products": products,
	})
}
func (s *Server) handleCVENew(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		s.render(w, r, "cve/new.html", nil)
		return
	}
	// TODO: implement creation
	http.Redirect(w, r, "/cves", http.StatusSeeOther)
}
func (s *Server) handleMyDashboard(w http.ResponseWriter, r *http.Request) {
	user, _ := s.getUserFromSession(r)
	subs, _ := s.subscriptionSvc.ListSubscriptions(r.Context(), user.ID.String())
	s.render(w, r, "dashboard/my.html", map[string]interface{}{"Subscriptions": subs})
}
func (s *Server) handleSubscribe(w http.ResponseWriter, r *http.Request) {
	// TODO: implement subscription logic
	http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther)
}
func (s *Server) handleAdminUsers(w http.ResponseWriter, r *http.Request) {
	var users []model.User
	s.db.NewSelect().Model(&users).Scan(r.Context())
	s.render(w, r, "user/list.html", map[string]interface{}{"Users": users})
}
