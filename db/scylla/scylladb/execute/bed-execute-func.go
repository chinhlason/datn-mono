package execute

import (
	"HospitalManager/db/scylla/scylladb"
	req2 "HospitalManager/dto/req/bed_req"
	"HospitalManager/dto/res"
	"HospitalManager/model"
	"context"
	"errors"
	"fmt"
	"github.com/gocql/gocql"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/scylladb/gocqlx/v2/qb"
	"log"
	"sync"
	"time"
)

func (q *Queries) InsertBed(req req2.InsertBedReq) error {
	_, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()
	tableName := fmt.Sprintf("%s.beds", q.keyspace)
	id, err := gocql.ParseUUID(uuid.New().String())
	if err != nil {
		panic(err)
	}
	room, err := q.GetRoomByOption(req.RoomName, "name")
	if err != nil {
		return err
	}
	if len(room) == 0 {
		return errors.New("No Room Data Found")
	}
	beds, _ := q.SelectBedFromRoom(req.Name, req.RoomName)
	if len(beds) > 0 {
		return errors.New("duplicate bed's name")
	}
	insert := &model.Beds{
		Id:       id,
		Name:     req.Name,
		IdRoom:   room[0].Id,
		Status:   "AVAILABLE",
		CreateAt: time.Now(),
		UpdateAt: time.Now(),
	}
	stmt := qb.Insert(tableName).
		Columns("id", "name", "id_room", "status", "create_at", "update_at").
		Query(q.session)
	stmt.BindStruct(insert)
	if err := stmt.ExecRelease(); err != nil {
		return err
	}
	return nil
}

func (q *Queries) SelectBedFromRoom(bed string, roomName string) ([]model.Beds, error) {
	_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	tableName := fmt.Sprintf("%s.beds", q.keyspace)
	room, err := q.GetRoomByOption(roomName, "name")
	if err != nil {
		return nil, err
	}
	if len(room) == 0 {
		return nil, errors.New("No Room Data Found")
	}
	var beds []model.Beds
	stmt, names := qb.Select(tableName).
		Where(qb.Eq("name"), qb.Eq("id_room")).
		ToCql()
	stmt += " ALLOW FILTERING"
	query := q.session.Query(stmt, names).BindMap(qb.M{
		"name":    bed,
		"id_room": room[0].Id,
	})
	if err := query.SelectRelease(&beds); err != nil {
		return nil, err
	}
	return beds, nil
}

func (q *Queries) SelectAllBedsFromRoom(roomName string) ([]model.Beds, error) {
	_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	tableName := fmt.Sprintf("%s.beds", q.keyspace)
	room, err := q.GetRoomByOption(roomName, "name")
	if err != nil {
		return nil, err
	}
	if len(room) == 0 {
		return nil, errors.New("No Room Data Found")
	}
	var beds []model.Beds
	stmt, names := qb.Select(tableName).
		Where(qb.Eq("id_room")).
		ToCql()
	stmt += " ALLOW FILTERING"
	query := q.session.Query(stmt, names).BindMap(qb.M{
		"id_room": room[0].Id,
	})
	if err := query.SelectRelease(&beds); err != nil {
		return nil, err
	}
	return beds, nil
}

func (q *Queries) GetBedByOption(value string, option string, roomName string) ([]model.Beds, error) {
	_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	tableName := fmt.Sprintf("%s.beds", q.keyspace)
	room, err := q.GetRoomByOption(roomName, "name")
	if err != nil {
		return nil, err
	}
	if len(room) == 0 {
		return nil, errors.New("No Room Data Found")
	}
	var beds []model.Beds
	stmt, names := qb.Select(tableName).
		Where(qb.Eq(option), qb.Eq("id_room")).
		ToCql()
	stmt += " ALLOW FILTERING"
	query := q.session.Query(stmt, names).BindMap(qb.M{
		option:    value,
		"id_room": room[0].Id,
	})
	if err := query.SelectRelease(&beds); err != nil {
		return nil, err
	}
	return beds, nil
}

func (q *Queries) GetAvailableAndDisableBed(c echo.Context) ([]model.Beds, error) {
	_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	rooms, err := q.SelectAllRoomByCurrDoctor(c)
	if err != nil {
		return nil, err
	}
	if len(rooms) == 0 {
		return nil, errors.New("No room data found")
	}
	var bedsRes []model.Beds
	for _, room := range rooms {
		beds, err := q.SelectAllBedsFromRoom(room.Name)
		if err != nil {
			return nil, err
		}
		for _, bed := range beds {
			if bed.Status == "AVAILABLE" || bed.Status == "DISABLED" {
				bedsRes = append(bedsRes, bed)
			}
		}
	}
	return bedsRes, nil
}

func (q *Queries) GetAvailableBed(c echo.Context) ([]model.Beds, error) {
	_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	rooms, err := q.SelectAllRoomByCurrDoctor(c)
	if err != nil {
		return nil, err
	}
	if len(rooms) == 0 {
		return nil, errors.New("No room data found")
	}
	var bedsRes []model.Beds
	for _, room := range rooms {
		beds, err := q.SelectAllBedsFromRoom(room.Name)
		if err != nil {
			return nil, err
		}
		for _, bed := range beds {
			if bed.Status == "AVAILABLE" {
				bedsRes = append(bedsRes, bed)
			}
		}
	}
	return bedsRes, nil
}
func (q *Queries) GetAvailableBedPagination(pageState []byte, pageSize int) (result []model.Beds, nextPage []byte, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	tableName := fmt.Sprintf("%s.beds", q.keyspace)
	stmt, names := qb.Select(tableName).
		Where(qb.Eq("status")).
		AllowFiltering().
		ToCql()

	query := q.session.Query(stmt, names).BindMap(qb.M{
		"status": "AVAILABLE",
	}).WithContext(ctx)

	query.PageSize(pageSize)
	query.PageState(pageState)

	iter := query.Iter()
	err = iter.Select(&result)
	if err != nil {
		log.Println("Error during iteration:", err)
		return []model.Beds{}, nil, errors.New("no beds found")
	}

	log.Println("Page State:", iter.PageState())
	log.Println("Number of beds found:", len(result))

	return result, iter.PageState(), nil
}

func (q *Queries) GetBedById(id string) (model.Beds, error) {
	_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var bed model.Beds
	tableName := fmt.Sprintf("%s.beds", q.keyspace)
	stmt, names := qb.Select(tableName).
		Where(qb.Eq("id")).
		ToCql()
	query := q.session.Query(stmt, names).BindMap(qb.M{
		"id": id,
	})
	if err := query.GetRelease(&bed); err != nil {
		return model.Beds{}, err
	}
	return bed, nil
}

func (q *Queries) UpdateBed(req req2.UpdateBedReq) error {
	_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	tableName := fmt.Sprintf("%s.beds", q.keyspace)
	room, err := q.GetRoomByOption(req.RoomName, "name")
	if err != nil {
		return err
	}
	if len(room) == 0 {
		return errors.New("No Room Data Found")
	}
	beds, err := q.GetBedByOption(req.OldName, "name", req.OldRoomName)
	if err != nil {
		return err
	}
	if len(beds) == 0 {
		return errors.New("No bed data found")
	}
	newBed, err := q.GetBedByOption(req.Name, "name", req.OldRoomName)
	if err != nil {
		return err
	}
	if len(newBed) > 0 && newBed[0].Id != beds[0].Id {
		return errors.New("duplicate name")
	}
	update := &scylladb.UpdateBedReq{
		Name:     req.Name,
		IdRoom:   room[0].Id.String(),
		Status:   req.Status,
		UpdateAt: time.Now(),
		Id:       beds[0].Id.String(),
	}
	stmt, names := qb.Update(tableName).
		Set("name").
		Set("id_room").
		Set("status").
		Set("update_at").
		Where(qb.Eq("id")).
		ToCql()

	query := q.session.Query(stmt, names).BindStruct(update)
	if err := query.ExecRelease(); err != nil {
		return err
	}
	return nil
}

func (q *Queries) ChangeBedStatus(bedName string, roomName string) error {
	_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	tableName := fmt.Sprintf("%s.beds", q.keyspace)
	bed, err := q.SelectBedFromRoom(bedName, roomName)
	if err != nil {
		return err
	}
	if len(bed) == 0 {
		return errors.New("No bed data found")
	}
	status := bed[0].Status
	if status == "AVAILABLE" {
		update := &scylladb.UpdateBedStt{
			Status:   "UNAVAILABLE",
			UpdateAt: time.Now(),
			Id:       bed[0].Id.String(),
		}
		stmt, names := qb.Update(tableName).
			Set("status").
			Set("update_at").
			Where(qb.Eq("id")).
			ToCql()

		query := q.session.Query(stmt, names).BindStruct(update)
		if err := query.ExecRelease(); err != nil {
			return err
		}
		return nil
	}
	update := &scylladb.UpdateBedStt{
		Status:   "AVAILABLE",
		UpdateAt: time.Now(),
		Id:       bed[0].Id.String(),
	}
	stmt, names := qb.Update(tableName).
		Set("status").
		Set("update_at").
		Where(qb.Eq("id")).
		ToCql()

	query := q.session.Query(stmt, names).BindStruct(update)
	if err := query.ExecRelease(); err != nil {
		return err
	}
	return nil
}

func (q *Queries) CreateUsageBed(req req2.UsageBedReq, c echo.Context) error {
	_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	tableName := fmt.Sprintf("%s.usage_bed", q.keyspace)
	bed, err := q.SelectBedFromRoom(req.BedName, req.RoomName)
	if err != nil {
		return err
	}
	if len(bed) == 0 {
		return errors.New("No bed data found")
	}
	if bed[0].Status == "UNAVAILABLE" || bed[0].Status == "DISABLED" {
		return errors.New("Bed is not available, cant handover")
	}

	room, err := q.GetRoomByOption(req.RoomName, "name")
	if err != nil {
		return err
	}
	if len(room) == 0 {
		return errors.New("no room data found")
	}

	record, err := q.GetRecordByOption(req.IdRecord, "id")
	if err != nil {
		return err
	}
	if len(record) == 0 {
		return errors.New("No record data found")
	}
	if record[0].Status == "TREATING" {
		return errors.New("Record already been handovered before")
	}
	id, err := gocql.ParseUUID(uuid.New().String())
	if err != nil {
		panic(err)
	}
	insert := &model.UsageBed{
		Id:       id,
		IdBed:    bed[0].Id,
		IdRecord: record[0].Id,
		Status:   "IN_USE",
		CreateAt: time.Now(),
		EndAt:    time.Time{},
	}
	stmt := qb.Insert(tableName).
		Columns("id", "id_record", "id_bed", "create_at", "end_at", "status").
		Query(q.session)
	stmt.BindStruct(insert)
	if err := stmt.ExecRelease(); err != nil {
		return err
	}
	err = q.UpdateNumber(1, "patient_number", req.RoomName)
	if err != nil {
		return err
	}

	err = q.UpdateRoomForRecord(room[0].Id.String(), record[0].Id.String())
	if err != nil {
		return err
	}

	err = q.ChangeRecordStatus(record[0].Id.String(), "TREATING")
	if err != nil {
		return err
	}

	err = q.UpdateUpdater(c, record[0].Id.String())
	if err != nil {
		return err
	}

	content := "Doctor handover bed"

	err = q.CreateRecordHistory(record[0].Id, content, c)
	if err != nil {
		return err
	}

	err = q.ChangeBedStatus(req.BedName, req.RoomName)
	if err != nil {
		return err
	}
	return nil
}

func (q *Queries) GetUsageBedByOption(value string, option string) ([]model.UsageBed, error) {
	_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	tableName := fmt.Sprintf("%s.usage_bed", q.keyspace)
	var usage_bed []model.UsageBed
	stmt, names := qb.Select(tableName).
		Where(qb.Eq(option)).
		ToCql()
	stmt += " ALLOW FILTERING"
	query := q.session.Query(stmt, names).BindMap(qb.M{
		option: value,
	})
	if err := query.SelectRelease(&usage_bed); err != nil {
		return nil, err
	}
	return usage_bed, nil
}

func (q *Queries) GetUsageBedByOptionAndUnavalible(value string, option string) ([]model.UsageBed, error) {
	_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	tableName := fmt.Sprintf("%s.usage_bed", q.keyspace)
	var usage_bed []model.UsageBed
	stmt, names := qb.Select(tableName).
		Where(qb.Eq(option)).
		ToCql()
	stmt += " ALLOW FILTERING"
	query := q.session.Query(stmt, names).BindMap(qb.M{
		option:   value,
		"status": "IN_USE",
	})
	if err := query.SelectRelease(&usage_bed); err != nil {
		return nil, err
	}
	return usage_bed, nil
}

func (q *Queries) UnuseBed(bedName string, roomName string, c echo.Context) error {
	_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	tableName := fmt.Sprintf("%s.usage_bed", q.keyspace)
	bed, err := q.GetBedByOption(bedName, "name", roomName)
	if err != nil {
		return err
	}
	if len(bed) == 0 {
		return errors.New("No bed data found")
	}
	if bed[0].Status == "AVAILABLE" {
		return errors.New("Bed already available")
	}

	usage_bed, err := q.GetUsageBedByOption(bed[0].Id.String(), "id_bed")
	if err != nil {
		return err
	}
	if len(usage_bed) == 0 {
		return errors.New("No usage bed data found")
	}

	record, err := q.GetRecordByOption(usage_bed[0].IdRecord.String(), "id")
	if err != nil {
		return err
	}
	if len(record) == 0 {
		return errors.New("No record data found")
	}
	update := &scylladb.UpdateUsageTable{
		Status: "NOT_IN_USE",
		EndAt:  time.Now(),
		Id:     usage_bed[0].Id.String(),
	}
	stmt, names := qb.Update(tableName).
		Set("status").
		Set("end_at").
		Where(qb.Eq("id")).
		ToCql()

	query := q.session.Query(stmt, names).BindStruct(update)
	if err := query.ExecRelease(); err != nil {
		return err
	}
	err = q.UpdateNumber((-1), "patient_number", roomName)
	if err != nil {
		return err
	}

	err = q.ChangeRecordStatus(record[0].Id.String(), "PENDING")
	if err != nil {
		return err
	}

	err = q.UpdateUpdater(c, record[0].Id.String())
	if err != nil {
		return err
	}

	content := "Doctor remove handover bed"

	err = q.CreateRecordHistory(record[0].Id, content, c)
	if err != nil {
		return err
	}

	err = q.ChangeBedStatus(bedName, roomName)
	if err != nil {
		return err
	}
	return nil
}

func (q *Queries) DisableOrEnableBed(id string, status string) error {
	_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	tableName := fmt.Sprintf("%s.beds", q.keyspace)
	bed, err := q.GetBedById(id)
	if err != nil {
		return err
	}
	if status == "DISABLED" {
		if bed.Status == "UNAVAILABLE" {
			return errors.New("Bed is used, can not disable")
		}
		update := &scylladb.DisableOrEnable{
			Status:   status,
			UpdateAt: time.Now(),
			Id:       bed.Id.String(),
		}
		stmt, names := qb.Update(tableName).
			Set("status").
			Set("update_at").
			Where(qb.Eq("id")).
			ToCql()

		query := q.session.Query(stmt, names).BindStruct(update)
		if err := query.ExecRelease(); err != nil {
			return err
		}
		return nil
	}
	if bed.Status != "DISABLED" {
		return errors.New("Bed is used, can not disable")
	}
	update := &scylladb.DisableOrEnable{
		Status:   "AVAILABLE",
		UpdateAt: time.Now(),
		Id:       bed.Id.String(),
	}
	stmt, names := qb.Update(tableName).
		Set("status").
		Set("update_at").
		Where(qb.Eq("id")).
		ToCql()

	query := q.session.Query(stmt, names).BindStruct(update)
	if err := query.ExecRelease(); err != nil {
		return err
	}
	return nil
}
func (q *Queries) GetBedRecord(c echo.Context) ([]res.BedRecord, error) {
	_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var results []res.BedRecord

	rooms, err := q.SelectAllRoomByCurrDoctor(c)
	if err != nil {
		return nil, err
	}

	if len(rooms) == 0 {
		return nil, errors.New("No room data found")
	}

	type bedRecordResult struct {
		record res.BedRecord
		err    error
	}

	bedRecordCh := make(chan bedRecordResult)

	var wg sync.WaitGroup

	for _, room := range rooms {
		wg.Add(1)
		go func(room model.Rooms) {
			defer wg.Done()

			beds, err := q.SelectAllBedsFromRoom(room.Name)
			if err != nil {
				bedRecordCh <- bedRecordResult{err: err}
				return
			}

			for _, bed := range beds {
				if bed.Status == "UNAVAILABLE" {
					wg.Add(1)
					go func(bed model.Beds) {
						defer wg.Done()

						usageBeds, err := q.GetUsageBedByOption(bed.Id.String(), "id_bed")
						if err != nil {
							bedRecordCh <- bedRecordResult{err: err}
							return
						}

						for _, usageBed := range usageBeds {
							if usageBed.Status == "IN_USE" {
								record, err := q.GetRecordByOption(usageBed.IdRecord.String(), "id")
								if err != nil {
									bedRecordCh <- bedRecordResult{err: err}
									return
								}

								if len(record) == 0 {
									bedRecordCh <- bedRecordResult{err: errors.New("No record data found")}
									return
								}

								patient, err := q.GetPatient(record[0].IdPatient.String(), "id")
								if err != nil {
									bedRecordCh <- bedRecordResult{err: err}
									return
								}

								resp := res.BedRecord{
									BedName:     bed.Name,
									RoomName:    room.Name,
									IdRecord:    record[0].Id.String(),
									PatientName: patient[0].Fullname,
								}
								bedRecordCh <- bedRecordResult{record: resp}
							}
						}
					}(bed)
				}
			}
		}(room)
	}

	// Close the channel when all goroutines are done
	go func() {
		wg.Wait()
		close(bedRecordCh)
	}()

	// Collect results from the channel
	for result := range bedRecordCh {
		if result.err != nil {
			return nil, result.err
		}
		results = append(results, result.record)
	}

	return results, nil
}
