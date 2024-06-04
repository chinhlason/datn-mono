package note_req

import "github.com/gocql/gocql"

type UpdateNoteReq struct {
	Id       gocql.UUID
	IdRecord gocql.UUID
	Content  string
	ImgUrl   string
}
