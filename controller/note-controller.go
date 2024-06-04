package controller

import (
	"HospitalManager/db/scylla/scylladb/execute"
	"HospitalManager/dto/req/note_req"
	"HospitalManager/dto/res"
	"HospitalManager/model"
	"github.com/labstack/echo/v4"
	"net/http"
)

type NoteController struct {
	Queries *execute.Queries
}

func (n *NoteController) CreateNote(c echo.Context) error {
	req := note_req.NoteReq{}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	err := n.Queries.CreateNote(req, c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, "create note success")
}

func (n *NoteController) DeleteNote(c echo.Context) error {
	id := c.QueryParam("id")
	err := n.Queries.DeleteNote(id)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, "delete note success")
}

func (n *NoteController) UpdateNote(c echo.Context) error {
	req := note_req.UpdateNoteReq{}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	err := n.Queries.UpdateNote(req, c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, "update note success")
}

func (n *NoteController) mapNoteToNoteRes(notes []model.Notes) []res.Note {
	var responses []res.Note
	for _, note := range notes {
		doctor, _ := n.Queries.GetUserByOption(note.IdDoctor.String(), "id")
		response := res.Note{
			Id:       note.Id,
			IdRecord: note.IdRecord,
			Content:  note.Content,
			ImgUrl:   note.ImgUrl,
			Doctor:   doctor[0],
			CreateAt: note.CreateAt,
			UpdateAt: note.UpdateAt,
		}
		responses = append(responses, response)
	}
	return responses
}

func (n *NoteController) GetAll(c echo.Context) error {
	id := c.QueryParam("id")
	notes, err := n.Queries.GetAllNote(id)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	res := n.mapNoteToNoteRes(notes)
	return c.JSON(http.StatusOK, res)
}
