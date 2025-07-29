package notes

import (
	"context"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"log/slog"
	"net/http"
	"strconv"
	resp "todo/internal/api/response"
	"todo/internal/models"
	"todo/internal/storage"
	"todo/pkg/logger/sl"
)

type NoteUpdater interface {
	UpdateNote(ctx context.Context, noteID, userID int, title, content string) (int64, error)
}

func NewUpdateNoteHandler(log *slog.Logger, noteUpdater NoteUpdater) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.notes.NewUpdateNoteHandler"
		var req models.Request
		var validationErrs validator.ValidationErrors

		log := log.With(
			slog.String("op", op),
			slog.String("request-id", middleware.GetReqID(r.Context())),
		)

		userID, err := strconv.Atoi(chi.URLParam(r, "id"))
		if err != nil {
			log.Info("url parameter conversion error", sl.Err(err))

			w.WriteHeader(400)
			render.JSON(w, r, resp.Err("user id must be a number"))

			return
		}

		noteID, err := strconv.Atoi(chi.URLParam(r, "note_id"))
		if err != nil {
			log.Info("url parameter conversion error", sl.Err(err))

			w.WriteHeader(400)
			render.JSON(w, r, resp.Err("note id must be a number"))

			return
		}

		if err := render.DecodeJSON(r.Body, &req); err != nil {
			log.Info("request decoding failed", sl.Err(err))

			w.WriteHeader(400)
			render.JSON(w, r, resp.Err("invalid request"))

			return
		}

		log.Info("request decoded", slog.Any("request", req))

		err = validator.New().Struct(req)
		if errors.As(err, &validationErrs) {
			log.Info("request validation failed", sl.Err(err))

			w.WriteHeader(400)
			render.JSON(w, r, resp.ValidationErrorsResponse(validationErrs))

			return
		}
		if err != nil {
			log.Error("request validation failed", sl.Err(err))

			w.WriteHeader(500)
			render.JSON(w, r, resp.Err("internal error"))

			return
		}

		id, err := noteUpdater.UpdateNote(r.Context(), noteID, userID, req.Title, req.Content)
		if errors.Is(err, context.Canceled) {
			log.Info("connection closed from client side, request cancelled", sl.Err(err))

			return
		}
		if errors.Is(err, context.DeadlineExceeded) {
			log.Warn("failed to get notes", sl.Err(err))

			w.WriteHeader(504)
			render.JSON(w, r, resp.Err("request took too long to process, try again later"))

			return
		}
		if errors.Is(err, storage.ErrNoNotes) {
			log.Info("failed to update notes", sl.Err(err))

			w.WriteHeader(404)
			render.JSON(w, r, resp.Err("no notes with this id"))

			return
		}
		if err != nil {
			log.Error("failed to save note", sl.Err(err))

			w.WriteHeader(500)
			render.JSON(w, r, resp.Err("internal error"))

			return
		}

		log.Info("note updated", slog.Int64("id", id))

		w.WriteHeader(201)
		render.JSON(w, r, resp.OK())
	}
}
