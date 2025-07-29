package postgres

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"time"
	"todo/internal/models"
	"todo/internal/storage"
)

const migrationDir = "internal/storage/migrations"

type Storage struct {
	pool            *pgxpool.Pool
	standardTimeout time.Duration
}

func New(connectionString string, standardQueryTimeout time.Duration) (storage *Storage, err error) {
	const op = "storage.postgres.New"
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	pool, err := pgxpool.New(ctx, connectionString)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	defer func() {
		if err != nil {
			pool.Close()
		}
	}()

	if err := goose.SetDialect("pgx"); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if err := goose.UpContext(ctx, stdlib.OpenDBFromPool(pool), migrationDir); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{pool, standardQueryTimeout}, nil
}

func (s *Storage) SaveUser(ctx context.Context, username string) (int64, error) {
	const op = "storage.postgres.SaveUser"
	ctx, cancel := context.WithTimeout(ctx, s.standardTimeout)
	defer cancel()
	var id int64
	var pgErr *pgconn.PgError

	err := s.pool.QueryRow(
		ctx,
		`INSERT INTO users (username)
		VALUES ($1)
		RETURNING id`,
		username,
	).Scan(&id)
	if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
		return 0, fmt.Errorf("%s: %w", op, storage.ErrUserExist)
	}
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

func (s *Storage) SaveNote(ctx context.Context, userID int, title, content string) (int64, error) {
	const op = "storage.postgres.SaveNote"
	ctx, cancel := context.WithTimeout(ctx, s.standardTimeout)
	defer cancel()
	var id int64

	if err := s.pool.QueryRow(
		ctx,
		`INSERT INTO notes(user_id, title, content)
		VALUES ($1, $2, $3)
		RETURNING id`,
		userID,
		title,
		content,
	).Scan(&id); err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

func (s *Storage) GetNotes(ctx context.Context, userID, limit, offset int, sort string) ([]models.Note, []int64, error) {
	const op = "storage.postgres.GetNotes"
	ctx, cancel := context.WithTimeout(ctx, s.standardTimeout)
	defer cancel()
	resNotes := make([]models.Note, 0, limit)
	resIDs := make([]int64, 0, limit)

	safeQuery := fmt.Sprintf(
		`SELECT id, title, content, created_at, updated_at
		FROM notes
		WHERE user_id = $1
		ORDER BY created_at %s
		LIMIT $2
		OFFSET $3`,
		sort,
	)

	rows, err := s.pool.Query(ctx, safeQuery, userID, limit, offset)
	if err != nil {
		return nil, nil, fmt.Errorf("%s: %w", op, err)
	}

	for rows.Next() {
		var note models.Note

		if err := rows.Scan(&note.ID, &note.Title, &note.Content, &note.CreatedAt, &note.UpdatedAt); err != nil {
			return nil, nil, fmt.Errorf("%s: %w", op, err)
		}

		resNotes = append(resNotes, note)
		resIDs = append(resIDs, note.ID)
	}

	if len(resNotes) == 0 {
		return nil, nil, fmt.Errorf("%s: %w", op, storage.ErrUserNoNotes)
	}

	return resNotes, resIDs, nil
}

func (s *Storage) GetNote(ctx context.Context, noteID, userID int) (models.Note, int64, error) {
	const op = "storage.postgres.GetNote"
	ctx, cancel := context.WithTimeout(ctx, s.standardTimeout)
	defer cancel()
	var note models.Note

	err := s.pool.QueryRow(
		ctx,
		`SELECT id, title, content, created_at, updated_at
		FROM notes
		WHERE id = $1 AND user_id = $2`,
		noteID,
		userID,
	).Scan(&note.ID, &note.Title, &note.Content, &note.CreatedAt, &note.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.Note{}, 0, fmt.Errorf("%s: %w", op, storage.ErrNoNotes)
	}
	if err != nil {
		return models.Note{}, 0, fmt.Errorf("%s: %w", op, err)
	}

	return note, note.ID, nil
}

func (s *Storage) UpdateNote(ctx context.Context, noteID, userID int, title, content string) (int64, error) {
	const op = "storage.postgres.UpdateNote"
	ctx, cancel := context.WithTimeout(ctx, s.standardTimeout)
	defer cancel()
	var id int64

	err := s.pool.QueryRow(
		ctx,
		`UPDATE notes
		SET title = $1, content = $2
		WHERE id = $3 AND user_id = $4
		RETURNING id`,
		title,
		content,
		noteID,
		userID,
	).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, fmt.Errorf("%s: %w", op, storage.ErrNoNotes)
	}
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

func (s *Storage) DeleteNote(ctx context.Context, noteID, userID int) (int64, error) {
	const op = "storage.postgres.DeleteNote"
	ctx, cancel := context.WithTimeout(ctx, s.standardTimeout)
	defer cancel()
	var id int64

	err := s.pool.QueryRow(
		ctx,
		`DELETE FROM notes
		WHERE id = $1 AND user_id = $2
		RETURNING id`,
		noteID,
		userID,
	).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, fmt.Errorf("%s: %w", op, storage.ErrNoNotes)
	}
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}
