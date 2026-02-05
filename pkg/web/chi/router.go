package chi

import (
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger/v2"

	"github.com/marcelofabianov/course/config"
	_ "github.com/marcelofabianov/course/docs" // Swagger docs
	"github.com/marcelofabianov/course/pkg/cache"
	"github.com/marcelofabianov/course/pkg/logger"
	"github.com/marcelofabianov/course/pkg/web"
	"github.com/marcelofabianov/course/pkg/web/middleware"
)

type RouterConfig struct {
	Config  *config.Config
	Logger  *logger.Logger
	Cache   *cache.Cache
	Routers []web.Router
}

func NewRouter(cfg RouterConfig) *chi.Mux {
	r := chi.NewRouter()

	securityLogger := middleware.NewSecurityLogger(cfg.Logger)

	r.Use(middleware.Recovery(cfg.Logger))
	r.Use(middleware.RequestID())
	r.Use(middleware.RealIP())
	r.Use(middleware.Logger(cfg.Logger))
	r.Use(middleware.SecurityHeaders(cfg.Config.HTTP.SecurityHeaders))

	// HTTPS enforcement
	if cfg.Config.HTTP.TLS.Enabled && cfg.Config.HTTP.TLS.HTTPSOnly {
		r.Use(middleware.HTTPSOnly(middleware.HTTPSOnlyConfig{
			Enabled:     cfg.Config.HTTP.TLS.HTTPSOnly,
			RedirectURL: cfg.Config.HTTP.TLS.RedirectURL,
		}))
	}

	if cfg.Config.HTTP.CORS.Enabled {
		r.Use(middleware.CORS(cfg.Config.HTTP.CORS))
	}

	r.Use(middleware.RequestSize(cfg.Config.HTTP.MaxBodySize))

	if cfg.Config.HTTP.Compression.Enabled {
		r.Use(chimiddleware.Compress(cfg.Config.HTTP.Compression.Level))
	}

	if cfg.Config.HTTP.RateLimit.Enabled && cfg.Cache != nil {
		rateLimiter := middleware.NewRateLimiter(
			cfg.Cache.Client(),
			true,
			cfg.Config.HTTP.RateLimit.TrustedProxies,
			securityLogger,
		)
		r.Use(rateLimiter.GlobalLimit(
			cfg.Config.HTTP.RateLimit.Global.Limit,
			cfg.Config.HTTP.RateLimit.Global.Window,
			cfg.Config.HTTP.RateLimit.Global.Burst,
		))
	}

	r.Use(chimiddleware.Heartbeat("/ping"))

	r.Get("/", web.RootHandler)
	r.Get("/health", web.LivenessHandler)
	r.Get("/health/ready", web.ReadinessHandler())

	r.Route("/api/v1", func(v1 chi.Router) {
		v1.Use(middleware.Timeout(cfg.Config.HTTP.RequestTimeout))
		v1.Use(middleware.AcceptJSON())
		v1.Use(chimiddleware.AllowContentType("application/json"))

		if cfg.Config.HTTP.CSRF.Enabled {
			csrf := middleware.NewCSRFProtection(
				cfg.Config.HTTP.CSRF.Secret,
				cfg.Config.HTTP.CSRF.CookieName,
				cfg.Config.HTTP.CSRF.HeaderName,
				cfg.Config.HTTP.CSRF.TTL,
				cfg.Config.HTTP.CSRF.Exempt,
				true,
				securityLogger,
			)
			v1.Use(csrf.Protect())
			v1.Get("/csrf-token", csrf.GetTokenHandler())
		}

		v1.Get("/swagger/*", httpSwagger.Handler(
			httpSwagger.URL("/api/v1/swagger/doc.json"),
		))

		for _, router := range cfg.Routers {
			router.RegisterRoutes(v1)
		}
	})

	return r
}
