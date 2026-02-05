package di

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/fx"

	"github.com/marcelofabianov/course/config"
	"github.com/marcelofabianov/course/pkg/cache"
	"github.com/marcelofabianov/course/pkg/database"
	"github.com/marcelofabianov/course/pkg/logger"
	"github.com/marcelofabianov/course/pkg/web"
	webchi "github.com/marcelofabianov/course/pkg/web/chi"
)

type DatabaseHealthChecker struct {
	db *database.DB
}

func NewDatabaseHealthChecker(db *database.DB) web.HealthChecker {
	return &DatabaseHealthChecker{db: db}
}

func (d *DatabaseHealthChecker) Name() string {
	return "database"
}

func (d *DatabaseHealthChecker) Check(ctx context.Context) error {
	return d.db.Ping(ctx)
}

type CacheHealthChecker struct {
	cache *cache.Cache
}

func NewCacheHealthChecker(cache *cache.Cache) web.HealthChecker {
	return &CacheHealthChecker{cache: cache}
}

func (c *CacheHealthChecker) Name() string {
	return "cache"
}

func (c *CacheHealthChecker) Check(ctx context.Context) error {
	return c.cache.Ping(ctx)
}

type RouterParams struct {
	fx.In

	Config         *config.Config
	Logger         *logger.Logger
	Cache          *cache.Cache
	Routers        []web.Router        `group:"routers"`
	HealthCheckers []web.HealthChecker `group:"health_checkers"`
}

func ProvideRouter(params RouterParams) *chi.Mux {
	router := webchi.NewRouter(webchi.RouterConfig{
		Config:  params.Config,
		Logger:  params.Logger,
		Cache:   params.Cache,
		Routers: params.Routers,
	})

	router.Get("/health/ready", web.ReadinessHandler(params.HealthCheckers...))

	return router
}

func ProvideServer(cfg *config.Config, log *logger.Logger, router *chi.Mux, lc fx.Lifecycle) *web.Server {
	server := web.NewServer(cfg, log, router)

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			log.Info("starting HTTP server",
				"addr", server.Addr(),
				"service", cfg.General.ServiceName,
				"env", cfg.General.Env,
			)
			go func() {
				if err := server.Start(); err != nil && err != http.ErrServerClosed {
					log.Error("server error", "error", err.Error())
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Info("shutting down HTTP server")
			return server.Shutdown(ctx)
		},
	})

	return server
}

func AsRouter(f any) any {
	return fx.Annotate(
		f,
		fx.As(new(web.Router)),
		fx.ResultTags(`group:"routers"`),
	)
}

func AsHealthChecker(f any) any {
	return fx.Annotate(
		f,
		fx.As(new(web.HealthChecker)),
		fx.ResultTags(`group:"health_checkers"`),
	)
}

var AppModule = fx.Module("app",
	fx.Provide(
		ProvideRouter,
		ProvideServer,
		AsHealthChecker(NewDatabaseHealthChecker),
		AsHealthChecker(NewCacheHealthChecker),
	),
)
