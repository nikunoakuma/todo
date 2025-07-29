package models

type Request struct {
	Title   string `json:"title" validate:"required"`
	Content string `json:"content,omitempty"`
}
type SaveUserRequest struct {
	Username string `json:"username" validate:"required"`
}
