package middleware

import (
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/golang-jwt/jwt/v5"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	resp "todo/internal/api/response"
	"todo/pkg/auth"
	"todo/pkg/logger/sl"
)

type TokenParser interface {
	ParseToken(token string) (int, error)
}

func Authorize(log *slog.Logger, tokenParser TokenParser) func(handler http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		f := func(w http.ResponseWriter, r *http.Request) {
			const op = "middleware.Authorize"

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

			authHeader := r.Header.Get("Authorization")

			authHeaderSplit := strings.Split(authHeader, " ")
			if len(authHeaderSplit) != 2 || authHeaderSplit[0] != "Bearer" {
				log.Info("Invalid Authorization header value format")

				w.WriteHeader(401)
				render.JSON(w, r, resp.Err("invalid Authorization header value format"))

				return
			}

			token := authHeaderSplit[1]
			if token == "" {
				log.Info("bearer token is missing")

				w.WriteHeader(401)
				render.JSON(w, r, resp.Err("bearer token is missing"))

				return
			}

			sub, err := tokenParser.ParseToken(token)
			if errors.Is(err, auth.ErrSubEmpty) {
				log.Info("invalid token", sl.Err(err))

				w.WriteHeader(401)
				render.JSON(w, r, resp.Err("invalid token"))

				return
			}
			if errors.Is(err, jwt.ErrTokenExpired) {
				log.Info("invalid token", sl.Err(err))

				w.WriteHeader(401)
				render.JSON(w, r, resp.Err("token is expired"))

				return
			}
			if err != nil {
				log.Info("invalid token", sl.Err(err))

				w.WriteHeader(401)
				render.JSON(w, r, resp.Err("invalid token"))

				return
			}

			log.Info("token validated", slog.Int("sub", sub))

			if sub != userID {
				log.Info("forbidden access attempt")

				w.WriteHeader(403)
				render.JSON(w, r, resp.Err("you have no access to this user's notes"))

				return
			}

			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(f)
	}
}
