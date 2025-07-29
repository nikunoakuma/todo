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
	"strings"
	resp "todo/internal/api/response"
	"todo/internal/models"
	"todo/internal/storage"
	"todo/pkg/logger/sl"
)

type NotesGetter interface {
	GetNotes(ctx context.Context, userID, limit, offset int, sort string) ([]models.Note, []int64, error)
}

func NewGetNotesHandler(log *slog.Logger, notesGetter NotesGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.notes.NewGetNotesHandler"
		resSort := "ASC"
		resLimit := 10
		resOffset := 0

		log := log.With(
			slog.String("op", op),
			slog.String("request-id", middleware.GetReqID(r.Context())),
		)

		userID, err := strconv.Atoi(chi.URLParam(r, "id"))
		if err != nil {
			log.Info("query parameter conversion error")

			w.WriteHeader(400)
			render.JSON(w, r, resp.Err("user id must be a number"))

			return
		}

		if limit := r.URL.Query().Get("limit"); limit != "" {
			resLimit, err = strconv.Atoi(limit)
			if err != nil {
				log.Info("query parameter conversion error")

				w.WriteHeader(400)
				render.JSON(w, r, resp.Err("limit must be a number"))

				return
			}
		}

		if offset := r.URL.Query().Get("offset"); offset != "" {
			resOffset, err = strconv.Atoi(offset)
			if err != nil {
				log.Info("query parameter conversion error")

				w.WriteHeader(400)
				render.JSON(w, r, resp.Err("offset must be a number"))

				return
			}
		}

		if sort := r.URL.Query().Get("sort"); sort != "" {
			resSort = sort
		}

		switch strings.ToLower(resSort) {
		case "asc":
			resSort = "ASC"
		case "desc":
			resSort = "DESC"
		default:
			log.Info("value of sort is not safe", slog.String("sort", resSort))

			w.WriteHeader(400)
			render.JSON(w, r, resp.Err(`sort must be either "asc" or "desc"`))

			return
		}

		notes, ids, err := notesGetter.GetNotes(r.Context(), userID, resLimit, resOffset, resSort)
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
		if errors.Is(err, storage.ErrUserNoNotes) {
			log.Info("failed to get notes", sl.Err(err))

			w.WriteHeader(404)
			render.JSON(w, r, resp.Err("user have no notes or offset is too big"))

			return
		}
		if err != nil {
			log.Error("failed to get notes", sl.Err(err))

			w.WriteHeader(500)
			render.JSON(w, r, resp.Err("internal error"))

			return
		}

		log.Info("got notes", slog.Any("ids", ids))

		render.JSON(w, r, models.GetNotesResponse{
			Response: resp.OK(),
			Notes:    notes,
		})
	}
}
