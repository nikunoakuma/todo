package notes

import (
	"context"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
	"strconv"
	resp "todo/internal/api/response"
	"todo/internal/models"
	"todo/internal/storage"
	"todo/pkg/logger/sl"
)

type NoteGetter interface {
	GetNote(ctx context.Context, noteID, userID int) (models.Note, int64, error)
}

func NewGetNoteHandler(log *slog.Logger, noteGetter NoteGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.notes.NewGetNoteHandler"

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

		note, id, err := noteGetter.GetNote(r.Context(), noteID, userID)
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
			log.Info("failed to get notes", sl.Err(err))

			w.WriteHeader(404)
			render.JSON(w, r, resp.Err("no notes with this id"))

			return
		}
		if err != nil {
			log.Error("failed to get notes", sl.Err(err))

			w.WriteHeader(500)
			render.JSON(w, r, resp.Err("internal error"))

			return
		}

		log.Info("got note", slog.Int64("id", id))

		render.JSON(w, r, models.GetNoteResponse{
			Response: resp.OK(),
			Note:     note,
		})
	}
}
