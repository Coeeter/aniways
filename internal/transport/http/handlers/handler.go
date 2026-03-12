package handlers

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/coeeter/aniways/docs"
	"github.com/coeeter/aniways/internal/app"
	"github.com/coeeter/aniways/internal/models"
	"github.com/coeeter/aniways/internal/service"
	"github.com/coeeter/aniways/internal/utils"
	"github.com/flowchartsman/swaggerui"
	"github.com/ggicci/httpin"
	"github.com/ggicci/httpin/integration"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
)

type Handler struct {
	r         *chi.Mux
	deps      *app.Deps
	validator *validator.Validate
	services  *service.Services
}

func New(deps *app.Deps, r *chi.Mux) *Handler {
	services := service.NewServices(deps)

	return &Handler{
		r:         r,
		deps:      deps,
		validator: validator.New(),
		services:  services,
	}
}

func (h *Handler) RegisterRoutes() {
	h.r.Get("/", h.home)
	h.r.Get("/home", h.getHome)
	h.r.Get("/admin", h.serveAdminPage)

	h.HealthRoutes()
	h.AnimeDetailsRoutes()
	h.AnimeListingRoutes()
	h.AnimeEpisodeRoutes()
	h.CharacterRoutes()
	h.AuthRoutes()
	h.OauthRoutes()
	h.UserRoutes()
	h.LibraryRoutes()
	h.SettingsRoutes()
	h.AdminRoutes()

	h.RegisterOpenAPIRoutes()

	h.r.Options("/*", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	h.r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	})
}

func (h *Handler) RegisterOpenAPIRoutes() {
	if h.deps.Env.AppEnv == "development" {
		h.r.Handle("/swagger/*", http.StripPrefix("/swagger", swaggerui.Handler(docs.OpenAPISpec)))
	}
}

func (h *Handler) home(w http.ResponseWriter, _ *http.Request) {
	w.Write([]byte("AniWays API\n"))
}

func (h *Handler) Router() *chi.Mux {
	return h.r
}

func (h *Handler) parsePagination(r *http.Request, defaultPage, defaultSize int) (page, size int, err error) {
	q := r.URL.Query()

	page = defaultPage
	if v := q.Get("page"); v != "" {
		pageVal, err2 := strconv.Atoi(v)
		if err2 != nil || pageVal < 1 {
			return 0, 0, fmt.Errorf("invalid page")
		}
		page = pageVal
	}

	size = defaultSize
	if v := q.Get("itemsPerPage"); v != "" {
		sizeVal, err2 := strconv.Atoi(v)
		if err2 != nil || sizeVal < 1 {
			return 0, 0, fmt.Errorf("invalid itemsPerPage")
		}
		size = sizeVal
	}

	return page, size, nil
}

func (h *Handler) pathParam(r *http.Request, key string) (string, error) {
	v := chi.URLParam(r, key)
	if v == "" {
		return "", fmt.Errorf("%s is required", key)
	}
	return v, nil
}

func (h *Handler) jsonError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(models.ErrorResponse{Error: msg})
}

func (h *Handler) jsonOK(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

func (h *Handler) logger(r *http.Request) *slog.Logger {
	logger, ok := utils.CtxValue[*slog.Logger](r.Context())
	if !ok {
		return slog.Default()
	}
	return logger.With("layer", "controller")
}

func (h *Handler) parseAndValidate(w http.ResponseWriter, r *http.Request, req any) bool {
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		h.jsonError(w, http.StatusBadRequest, "Invalid JSON")
		return false
	}

	if err := h.validator.Struct(req); err != nil {
		h.jsonValidationError(w, err)
		return false
	}

	return true
}

func (h *Handler) jsonValidationError(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)

	validationErrors, ok := err.(validator.ValidationErrors)
	if !ok {
		_ = json.NewEncoder(w).Encode(models.ErrorResponse{Error: "Validation failed"})
		return
	}

	details := make(map[string]string)
	for _, fieldErr := range validationErrors {
		field := fieldErr.Field()
		switch fieldErr.Tag() {
		case "required":
			details[field] = field + " is required"
		case "email":
			details[field] = "Invalid email format"
		case "min":
			details[field] = field + " must be at least " + fieldErr.Param() + " characters"
		case "max":
			details[field] = field + " must be at most " + fieldErr.Param() + " characters"
		case "oneof":
			details[field] = field + " must be one of: " + fieldErr.Param()
		default:
			details[field] = field + " is invalid"
		}
	}

	_ = json.NewEncoder(w).Encode(models.ValidationErrorResponse{
		Error:   "Validation failed",
		Details: details,
	})
}

func init() {
	integration.UseGochiURLParam("path", chi.URLParam)
}

func (h *Handler) getHttpInput(r *http.Request) any {
	return r.Context().Value(httpin.Input)
}
