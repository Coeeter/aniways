package service

import (
	"github.com/coeeter/aniways/internal/app"
	"github.com/coeeter/aniways/internal/service/admin"
	"github.com/coeeter/aniways/internal/service/anime"
	"github.com/coeeter/aniways/internal/service/auth"
	"github.com/coeeter/aniways/internal/service/library"
	"github.com/coeeter/aniways/internal/service/settings"
	"github.com/coeeter/aniways/internal/service/users"
)

type Services struct {
	Anime    *anime.AnimeService
	Library  *library.LibraryService
	Auth     *auth.AuthService
	Users    *users.UserService
	Settings *settings.SettingsService
	Admin    *admin.AdminService
}

func NewServices(deps *app.Deps) *Services {
	refresher := anime.NewRefresher(deps.Repo, deps.MAL)
	animeService := anime.NewAnimeService(deps.Repo, refresher, deps.MAL, deps.Jikan, deps.Anilist, deps.Shiki, deps.Cache)
	libraryService := library.NewLibraryService(deps.Repo, refresher)
	authService := auth.NewAuthService(deps.Repo, deps.EmailClient, deps.Env.FrontendURL)
	userService := users.NewUserService(deps.Repo, deps.Cld)
	settingsService := settings.NewSettingsService(deps.Repo)
	adminService := admin.NewAdminService(deps.Repo, deps.Scraper)

	return &Services{
		Anime:    animeService,
		Library:  libraryService,
		Auth:     authService,
		Users:    userService,
		Settings: settingsService,
		Admin:    adminService,
	}
}
