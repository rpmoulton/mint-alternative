package server

import (
	"fmt"
	"log/slog"
	"net/http"

	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/rpmoulton/mint-alternative/config"
	"github.com/rpmoulton/mint-alternative/internal/api/accounts"
	"github.com/rpmoulton/mint-alternative/internal/api/auth"
	"github.com/rpmoulton/mint-alternative/internal/api/budgets"
	"github.com/rpmoulton/mint-alternative/internal/api/categories"
	"github.com/rpmoulton/mint-alternative/internal/api/dashboard"
	"github.com/rpmoulton/mint-alternative/internal/api/institutions"
	"github.com/rpmoulton/mint-alternative/internal/api/transactions"
	"github.com/rpmoulton/mint-alternative/internal/db"
	"github.com/rpmoulton/mint-alternative/internal/httpx"
	"github.com/rpmoulton/mint-alternative/internal/middleware"
	"github.com/rpmoulton/mint-alternative/pkg/clientip"
)

// Routes declares the complete API tree. Build compiles request binding,
// validation, response writing, errors, guards, routing, and OpenAPI from the
// same immutable operation plans.
// The tree is declared exactly once and compiled by both the serving path
// and the docs generator, so a route cannot exist in one and not the other.
func (s *Server) Routes() *httpx.Group {
	authHandler := auth.NewHandler(auth.NewService(auth.NewRepository(s.db)))
	institutionsHandler := institutions.NewHandler(institutions.NewService(institutions.NewRepository(s.db)))
	accountsHandler := accounts.NewHandler(accounts.NewService(accounts.NewRepository(s.db)))
	categoriesHandler := categories.NewHandler(categories.NewService(categories.NewRepository(s.db)))
	transactionsHandler := transactions.NewHandler(transactions.NewService(transactions.NewRepository(s.db)))
	budgetsHandler := budgets.NewHandler(budgets.NewService(budgets.NewRepository(s.db)))
	dashboardHandler := dashboard.NewHandler(dashboard.NewService(dashboard.NewRepository(s.db)))

	root := httpx.NewGroup(httpx.Defaults{})
	root.HandleInfra(http.MethodGet, "/metrics", promhttp.Handler())
	root.HandleInfra(http.MethodGet, "/docs/json", http.HandlerFunc(openAPIJSON))
	root.HandleInfra(http.MethodGet, "/docs/ui", http.HandlerFunc(openAPIUI))

	v1 := root.Sub("/api/v1")
	v1.Mount("/auth", authHandler.Routes())
	v1.Mount("/institutions", institutionsHandler.Routes())
	v1.Mount("/accounts", accountsHandler.Routes())
	v1.Mount("/categories", categoriesHandler.Routes())
	v1.Mount("/transactions", transactionsHandler.Routes())
	v1.Mount("/budgets", budgetsHandler.Routes())
	v1.Mount("/dashboard", dashboardHandler.Routes())
	return root
}

// MustBuildHandler is BuildHandler with the composition root's
// fail-loudly-at-startup behavior.
func (s *Server) MustBuildHandler() http.Handler {
	handler, err := s.BuildHandler()
	if err != nil {
		panic(err.Error())
	}
	return handler
}

func (s *Server) BuildHandler() (http.Handler, error) {
	app, err := s.Routes().Build()
	if err != nil {
		return nil, err
	}
	return s.wrapTransport(app), nil
}

func (s *Server) wrapTransport(app http.Handler) http.Handler {
	clientAddr := clientip.NewResolver(s.conf.Server.TrustedProxyPrefixes())
	middlewares := []func(http.Handler) http.Handler{
		clientAddr.Middleware,
		chimiddleware.RequestID,
		middleware.Logger(s.log),
		middleware.Recovery,
		cors.Handler(s.conf.Cors.CorsOptions()),
		s.conf.Secure.SecureOptions().Handler,
		chimiddleware.Heartbeat("/ping"),
	}
	handler := app
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}
	return handler
}

// RouterForDocs compiles the REAL Routes() tree over inert dependencies —
// cmd/openapi generates the document from its OpenAPI(), and any coverage
// gate walks its Operations(). It mirrors Bootstrap's construction order
// without a database: every service constructor runs for real, and the inert
// client panics loudly if that invariant is ever broken (see db.NewInert).
//
// jwt.InitJWT is deliberately skipped: the JWT service is a request-time
// singleton that no constructor dereferences, so docs generation needs no
// keys. Failures here are programmer errors in hard-coded configuration on a
// build-time tool path, so they panic rather than returning.
func RouterForDocs(conf *config.Config) *httpx.Application {
	if err := httpx.InitValidator(); err != nil {
		panic(fmt.Sprintf("docs router: registering validators: %v", err))
	}

	srv := &Server{
		conf: conf,
		db:   db.NewInert(),
		log:  slog.Default(),
	}
	return srv.Routes().MustBuild()
}
