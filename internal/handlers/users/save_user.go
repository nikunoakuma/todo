package users

import (
	"context"
	"errors"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"log/slog"
	"net/http"
	"time"
	resp "todo/internal/api/response"
	"todo/internal/models"
	"todo/internal/storage"
	"todo/pkg/logger/sl"
)

type UserSaver interface {
	SaveUser(ctx context.Context, username string) (int64, error)
}

type AccessTokenGenerator interface {
	NewAccessToken(userID int, tokenTTL time.Duration) (string, error)
}

func NewSaveUserHandler(log *slog.Logger, userSaver UserSaver, accessTokenGenerator AccessTokenGenerator, jwtTTL time.Duration) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.users.NewSaveUserHandler"
		var req models.SaveUserRequest
		var validationErrs validator.ValidationErrors

		log := log.With(
			slog.String("op", op),
			slog.String("request-id", middleware.GetReqID(r.Context())),
		)

		if err := render.DecodeJSON(r.Body, &req); err != nil {
			log.Info("request decoding failed", sl.Err(err))

			w.WriteHeader(400)
			render.JSON(w, r, resp.Err("invalid request"))

			return
		}

		log.Info("request decoded", slog.Any("request", req))

		err := validator.New().Struct(req)
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

		id, err := userSaver.SaveUser(r.Context(), req.Username)
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
		if errors.Is(err, storage.ErrUserExist) {
			log.Info("failed to save user", sl.Err(err))

			w.WriteHeader(409)
			render.JSON(w, r, resp.Err("user with this username already exists"))

			return
		}
		if err != nil {
			log.Error("failed to save user", sl.Err(err))

			w.WriteHeader(500)
			render.JSON(w, r, resp.Err("internal error"))

			return
		}

		log.Info("user saved", slog.Int64("id", id))

		jwt, err := accessTokenGenerator.NewAccessToken(int(id), jwtTTL)
		if err != nil {
			log.Error("failed to generate jwt", sl.Err(err))

			w.WriteHeader(500)
			render.JSON(w, r, resp.Err("internal error"))

			return
		}

		w.WriteHeader(201)
		render.JSON(w, r, models.SaveUserResponse{
			Response: resp.OK(),
			ID:       id,
			JWT:      jwt,
		})
	}
}
