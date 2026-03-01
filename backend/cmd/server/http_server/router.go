package http_server

import (
	"net/http"

	"github.com/uptrace/bun"

	"gitlab.com/jacky850509/secra/internal/api/web"
	"gitlab.com/jacky850509/secra/internal/config"
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

func (s *Server) setupRoutes() {
	// Health check
	s.mux.HandleFunc("/health", s.handleHealth)

	// Register REST API routes (handles all /api/v1/*)
	web.RegisterRoutes(s.mux, s.db, s.userSvc, s.cveSvc, s.subscriptionSvc)
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
