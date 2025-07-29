package main

import (
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"log/slog"
	"net/http"
	"os"
	"todo/internal/config"
	"todo/internal/handlers/notes"
	"todo/internal/handlers/users"
	appmiddleware "todo/internal/middleware"
	"todo/internal/storage/postgres"
	"todo/pkg/auth"
	"todo/pkg/logger"
	"todo/pkg/logger/sl"
)

func main() {
	cfg := config.MustLoad()

	log := logger.New(cfg.Env)
	log.Debug("debug messages are enabled")

	manager, err := auth.NewManager()
	if err != nil {
		log.Error("manager initialization failed", sl.Err(err))
		os.Exit(1)
	}
	log.Info("manager initialized")

	storage, err := postgres.New(cfg.ConnectionString, cfg.StandardQueryTimeout)
	if err != nil {
		log.Error("storage initialization failed", sl.Err(err))
		os.Exit(1)
	}
	log.Info("storage initialized")

	router := chi.NewRouter()

	router.Use(
		middleware.RequestID,
		middleware.Logger,
		middleware.Recoverer,
	)

	router.Post("/users", users.NewSaveUserHandler(log, storage, manager, cfg.AccessTokenTTl))

	router.Route(
		"/users/{id}/notes",
		func(r chi.Router) {
			r.Use(appmiddleware.Authorize(log, manager))

			r.Post("/", notes.NewSaveNoteHandler(log, storage))
			r.Get("/", notes.NewGetNotesHandler(log, storage))
		},
	)

	router.Route(
		"/users/{id}/notes/{note_id}",
		func(r chi.Router) {
			r.Use(appmiddleware.Authorize(log, manager))

			r.Get("/", notes.NewGetNoteHandler(log, storage))
			r.Put("/", notes.NewUpdateNoteHandler(log, storage))
			r.Delete("/", notes.NewDeleteNoteHandler(log, storage))
		},
	)

	srv := http.Server{
		Addr:         cfg.Address,
		Handler:      http.TimeoutHandler(router, cfg.RequestTimeout, "service unavailable"),
		ReadTimeout:  cfg.Timeout,
		WriteTimeout: cfg.Timeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	log.Info("server started", slog.String("address", cfg.Address))

	err = srv.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Error("server error", sl.Err(err))
		os.Exit(1)
	}

	log.Info("server closed")
}
