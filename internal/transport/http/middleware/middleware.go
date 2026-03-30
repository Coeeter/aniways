package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/coeeter/aniways/internal/config"
	"github.com/coeeter/aniways/internal/models"
	"github.com/coeeter/aniways/internal/repository"
	"github.com/coeeter/aniways/internal/service/users"
	"github.com/coeeter/aniways/internal/utils"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httplog/v3"
	"github.com/go-chi/httprate"
	httprateredis "github.com/go-chi/httprate-redis"
)

type MiddlewareConfig struct {
	Router *chi.Mux
	Logger *slog.Logger
	Env    *config.Env
	Repo   *repository.Queries
	Cld    *cloudinary.Cloudinary
}

func UseMiddlewares(c MiddlewareConfig) {
	userService := users.NewUserService(c.Repo, c.Cld)
	c.Router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{c.Env.AllowedOrigins},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type"},
		MaxAge:           300,
		AllowCredentials: true,
	}))

	c.Router.Use(
		middleware.RealIP,
		middleware.RequestID,
		rateLimiter(c.Env),
		injectUser(userService),
		injectLogger(c.Logger),
		requestLogger,
		middleware.Recoverer,
		middleware.Timeout(60*time.Second),
	)
}

func rateLimiter(env *config.Env) func(http.Handler) http.Handler {
	redisHost := strings.Split(env.RedisAddr, ":")[0]
	redisPort, err := strconv.Atoi(strings.Split(env.RedisAddr, ":")[1])
	if err != nil {
		redisPort = 6379
	}

	// General rate limiters for non-auth endpoints
	nosessionLimiter := httprate.Limit(
		60, 5*time.Minute,
		httprate.WithKeyFuncs(func(r *http.Request) (string, error) {
			ip := strings.TrimSpace(r.Header.Get("CF-Connecting-IP"))
			if ip == "" {
				ip, _ = httprate.KeyByIP(r)
			}
			return fmt.Sprintf("ip:%s", ip), nil
		}),
		httprateredis.WithRedisLimitCounter(
			&httprateredis.Config{
				Host:      redisHost,
				Port:      uint16(redisPort),
				Password:  env.RedisPassword,
				PrefixKey: "aniways_rl_ip_nosession:",
			},
		),
	)

	sessionLimiter := httprate.Limit(
		120, 5*time.Minute,
		httprate.WithKeyFuncs(func(r *http.Request) (string, error) {
			ip := strings.TrimSpace(r.Header.Get("CF-Connecting-IP"))
			if ip == "" {
				ip, _ = httprate.KeyByIP(r)
			}
			session, _ := r.Cookie("aniways_session")
			return fmt.Sprintf("ip:%s:session:%s", ip, session), nil
		}),
		httprateredis.WithRedisLimitCounter(
			&httprateredis.Config{
				Host:      redisHost,
				Port:      uint16(redisPort),
				Password:  env.RedisPassword,
				PrefixKey: "aniways_rl_ip_session:",
			},
		),
	)

	// Stricter rate limiter for login endpoint (security-critical)
	loginLimiter := httprate.Limit(
		5, 3*time.Minute, // e.g., 5 login attempts every 3 minutes
		httprate.WithKeyFuncs(func(r *http.Request) (string, error) {
			ip := strings.TrimSpace(r.Header.Get("CF-Connecting-IP"))
			if ip == "" {
				ip, _ = httprate.KeyByIP(r)
			}
			return fmt.Sprintf("ip:%s", ip), nil
		}),
		httprateredis.WithRedisLimitCounter(
			&httprateredis.Config{
				Host:      redisHost,
				Port:      uint16(redisPort),
				Password:  env.RedisPassword,
				PrefixKey: "aniways_rl_login:",
			},
		),
	)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			// Separate rate limiting for login endpoint (security-critical)
			if r.URL.Path == "/auth/login" && r.Method == http.MethodPost {
				loginLimiter(next).ServeHTTP(w, r)
				return
			}

			if r.URL.Path == "/anime/listings" || r.URL.Path == "/anime/listings/search" || r.URL.Path == "/home" {
				next.ServeHTTP(w, r)
				return
			}

			// Apply general rate limiting to other endpoints
			session, err := r.Cookie("aniways_session")
			if err != nil || session == nil || session.Value == "" {
				nosessionLimiter(next).ServeHTTP(w, r)
				return
			}

			sessionLimiter(next).ServeHTTP(w, r)
		})
	}
}

func injectLogger(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			reqLogger := logger.With("request_id", middleware.GetReqID(r.Context()))

			if user, ok := utils.CtxValue[models.UserResponse](r.Context()); ok {
				reqLogger = reqLogger.With("user_id", user.ID)
			}

			ctx := utils.CtxWithValue(r.Context(), reqLogger)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func injectUser(userService *users.UserService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("aniways_session")
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			user, err := userService.GetUserBySessionID(r.Context(), cookie.Value)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			ctx := utils.CtxWithValue(r.Context(), user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func requestLogger(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger, ok := utils.CtxValue[*slog.Logger](r.Context())
		if !ok {
			logger = slog.Default()
		}

		mw := httplog.RequestLogger(logger, &httplog.Options{
			Level: slog.LevelInfo,
		})

		mw(h).ServeHTTP(w, r)
	})
}
