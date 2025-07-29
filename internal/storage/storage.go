package storage

import (
	"errors"
)

var (
	ErrUserExist   = errors.New("user with this id already exists")
	ErrUserNoNotes = errors.New("no notes with this user_id and query parameters")
	ErrNoNotes     = errors.New("no notes with this id")
)
