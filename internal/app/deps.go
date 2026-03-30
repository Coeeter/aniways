package app

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/coeeter/aniways/internal/config"
	"github.com/coeeter/aniways/internal/infra/cache"
	"github.com/coeeter/aniways/internal/infra/client/anilist"
	"github.com/coeeter/aniways/internal/infra/client/hianime"
	"github.com/coeeter/aniways/internal/infra/client/jikan"
	"github.com/coeeter/aniways/internal/infra/client/myanimelist"
	"github.com/coeeter/aniways/internal/infra/client/shikimori"
	"github.com/coeeter/aniways/internal/infra/database"
	"github.com/coeeter/aniways/internal/infra/email"
	"github.com/coeeter/aniways/internal/repository"
	"github.com/coeeter/aniways/internal/service/auth/oauth"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Deps struct {
	Env         *config.Env
	Log         *slog.Logger
	Db          *pgxpool.Pool
	Repo        *repository.Queries
	Cache       *cache.RedisClient
	Scraper     *hianime.HianimeScraper
	MAL         *myanimelist.Client
	Jikan       *jikan.Client
	Anilist     *anilist.Client
	Shiki       *shikimori.Client
	Cld         *cloudinary.Cloudinary
	EmailClient email.EmailClient
	Providers   map[string]oauth.Provider
}

func InitDeps(ctx context.Context, svcName string) (*Deps, error) {
	env, err := config.LoadEnv()
	rootLogger := NewLogger(svcName) // if env is loaded wrongly just use json handler (prod way)
	if err != nil {
		return &Deps{Log: rootLogger}, fmt.Errorf("failed to load environment variables: %w", err)
	}

	deps := &Deps{
		Env: env,
		Log: rootLogger,
	}

	dbLog := rootLogger.With("component", "database")
	db, err := database.New(ctx, env, dbLog)
	if err != nil {
		return deps, fmt.Errorf("failed to initialize database: %w", err)
	}
	deps.Db = db

	cacheLog := rootLogger.With("component", "cache")
	useCache := env.AppEnv != "development" || env.UseCache
	redisCache, err := cache.NewRedisClient(ctx, env.AppEnv, env.RedisAddr, env.RedisPassword, cacheLog, useCache)
	if err != nil {
		deps.Close()
		return deps, fmt.Errorf("failed to initialize redis cache: %w", err)
	}
	deps.Cache = redisCache

	deps.Repo = repository.New(db)
	deps.Scraper = hianime.NewHianimeScraper()
	deps.MAL = myanimelist.NewClient(env.MyAnimeListClientID)
	deps.Jikan = jikan.NewClient()
	deps.Anilist = anilist.New()
	deps.Shiki = shikimori.NewClient(redisCache)
	deps.EmailClient = email.NewClient(env.ResendAPIKey, env.ResendFromEmail)

	cld, err := cloudinary.NewFromParams(env.CloudinaryName, env.CloudinaryAPIKey, env.CloudinaryAPISecret)
	if err != nil {
		deps.Close()
		return deps, fmt.Errorf("failed to initialize cloudinary client: %w", err)
	}
	deps.Cld = cld

	malOauthProvider := oauth.NewMALProvider(
		env.MyAnimeListClientID,
		env.MyAnimeListClientSecret,
		fmt.Sprintf("%s/auth/oauth/myanimelist/callback", env.ApiURL),
		deps.Repo,
		redisCache,
	)

	anilistOauthProvider := oauth.NewAnilistProvider(
		env.AnilistClientID,
		env.AnilistClientSecret,
		fmt.Sprintf("%s/auth/oauth/anilist/callback", env.ApiURL),
		deps.Repo,
	)

	deps.Providers = map[string]oauth.Provider{
		malOauthProvider.Name():     malOauthProvider,
		anilistOauthProvider.Name(): anilistOauthProvider,
	}

	return deps, nil
}

func (d *Deps) Close() error {
	if d.Db != nil {
		d.Db.Close()
	}

	if d.Cache != nil {
		if err := d.Cache.Close(); err != nil {
			d.Log.Error("Failed to close cache", "error", err)
			return err
		}
	}

	return nil
}
