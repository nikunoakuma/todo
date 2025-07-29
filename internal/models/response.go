package models

type Response struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}
type GetNoteResponse struct {
	Response
	Note `json:"note,omitempty"`
}

type GetNotesResponse struct {
	Response
	Notes []Note `json:"notes,omitempty"`
}

type SaveNoteResponse struct {
	Response
	ID int64 `json:"id"`
}

type SaveUserResponse struct {
	Response
	ID  int64  `json:"id"`
	JWT string `json:"jwt"`
}
