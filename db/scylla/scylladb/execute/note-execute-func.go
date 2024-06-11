package execute

import (
	"HospitalManager/dto/req/note_req"
	"HospitalManager/model"
	"context"
	"errors"
	"fmt"
	"github.com/gocql/gocql"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/scylladb/gocqlx/v2/qb"
	"time"
)

func (q *Queries) CreateNote(req note_req.NoteReq, c echo.Context) error {
	_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	tableName := fmt.Sprintf("%s.notes", q.keyspace)
	idRecord, err := gocql.ParseUUID(uuid.New().String())
	if err != nil {
		panic(err)
	}

	var record model.MedicalRecords
	records, err := q.GetRecordByOption(req.IdRecord, "id")
	if err != nil {
		return err
	}
	if len(records) == 0 {
		return errors.New("No record data found")
	}
	for _, rc := range records {
		if rc.Status == "PENDING" || rc.Status == "TREATING" {
			record = rc
		}
	}
	if record == (model.MedicalRecords{}) {
		return errors.New("No record data found")
	}
	doctorNote, err := q.GetProfileCurrent(c)
	if err != nil {
		return err
	}

	insert := &model.Notes{
		Id:       idRecord,
		IdDoctor: doctorNote.Id,
		IdRecord: record.Id,
		Content:  req.Content,
		ImgUrl:   req.ImgUrl,
		CreateAt: time.Now(),
		UpdateAt: time.Now(),
	}
	stmt := qb.Insert(tableName).
		Columns("id", "id_doctor", "id_record", "content", "img_url", "create_at", "update_at").
		Query(q.session)
	stmt.BindStruct(insert)
	if err := stmt.ExecRelease(); err != nil {
		return err
	}

	err = q.UpdateUpdater(c, record.Id.String())
	if err != nil {
		return err
	}

	content := "Doctor create note"

	err = q.CreateRecordHistory(record.Id, content, c)
	if err != nil {
		return err
	}

	return nil
}

func (q *Queries) DeleteNote(id string) error {
	_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	tableName := fmt.Sprintf("%s.notes", q.keyspace)
	stmt, names := qb.Delete(tableName).
		Where(qb.Eq("id")).
		ToCql()
	query := q.session.Query(stmt, names).BindMap(qb.M{
		"id": id,
	})
	if err := query.ExecRelease(); err != nil {
		return err
	}
	return nil
}

func (q *Queries) UpdateNote(req note_req.UpdateNoteReq, c echo.Context) error {
	_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	tableName := fmt.Sprintf("%s.notes", q.keyspace)
	update := &model.Notes{
		Content:  req.Content,
		ImgUrl:   req.ImgUrl,
		UpdateAt: time.Now(),
		Id:       req.Id,
	}
	stmt, names := qb.Update(tableName).
		Set("content").
		Set("img_url").
		Set("update_at").
		Where(qb.Eq("id")).
		ToCql()

	query := q.session.Query(stmt, names).BindStruct(update)
	if err := query.ExecRelease(); err != nil {
		return err
	}

	err := q.UpdateUpdater(c, req.IdRecord.String())
	if err != nil {
		return err
	}

	content := "Doctor update note"

	err = q.CreateRecordHistory(req.IdRecord, content, c)
	if err != nil {
		return err
	}

	return nil
}

func (q *Queries) GetAllNote(idRecord string) ([]model.Notes, error) {
	_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	tableName := fmt.Sprintf("%s.notes", q.keyspace)
	var notes []model.Notes
	stmt, names := qb.Select(tableName).
		Where(qb.Eq("id_record")).
		ToCql()
	stmt += " ALLOW FILTERING"
	query := q.session.Query(stmt, names).BindMap(qb.M{
		"id_record": idRecord,
	})
	if err := query.SelectRelease(&notes); err != nil {
		return []model.Notes{}, err
	}
	return notes, nil
}
