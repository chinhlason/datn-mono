package execute

import (
	"HospitalManager/db/scylla/scylladb"
	"HospitalManager/dto/req/room_req"
	"HospitalManager/dto/res"
	"HospitalManager/model"
	"context"
	"errors"
	"fmt"
	"github.com/gocql/gocql"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/scylladb/gocqlx/v2/qb"
	"sort"
	"time"
)

func (q *Queries) InsertRoom(req room_req.CreateRoomReq) error {
	_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	tableName := fmt.Sprintf("%s.rooms", q.keyspace)
	room, err := q.GetRoomByOption(req.Name, "name")
	if err != nil {
		return err
	}
	if len(room) > 0 {
		return errors.New("Duplicate bed's name")
	}
	id, err := gocql.ParseUUID(uuid.New().String())
	if err != nil {
		panic(err)
	}
	doctor, err := q.GetUserByOption(req.DoctorCode, "doctor_code")
	if err != nil {
		return err
	}
	insert := &model.Rooms{
		Id:            id,
		Name:          req.Name,
		IdDoctor:      doctor[0].Id,
		BedNumber:     0,
		PatientNumber: 0,
		CreateAt:      time.Now(),
		UpdateAt:      time.Now(),
	}
	stmt := qb.Insert(tableName).
		Columns("id", "name", "id_doctor", "bed_number", "patient_number", "create_at", "update_at").
		Query(q.session)
	stmt.BindStruct(insert)
	if err := stmt.ExecRelease(); err != nil {
		return err
	}
	err = q.HandoverRoomForNormalDoctor(id.String(), doctor[0].Id.String())
	if err != nil {
		return err
	}
	return nil
}

func (q *Queries) GetRoomByOption(value string, option string) ([]model.Rooms, error) {
	_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	tableName := fmt.Sprintf("%s.rooms", q.keyspace)
	if option == "doctor_code" {
		user, err := q.GetUserByOption(value, option)
		if err != nil {
			return nil, err
		}
		value = user[0].Id.String()
		option = "id_doctor"
	}
	var rooms []model.Rooms
	stmt, names := qb.Select(tableName).
		Where(qb.Eq(option)).
		ToCql()
	stmt += " ALLOW FILTERING"
	query := q.session.Query(stmt, names).BindMap(qb.M{
		option: value,
	})
	if err := query.SelectRelease(&rooms); err != nil {
		return []model.Rooms{}, err
	}
	return rooms, nil
}

func (q *Queries) GetAllRooms() ([]model.Rooms, error) {
	_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	tableName := fmt.Sprintf("%s.rooms", q.keyspace)
	var res []model.Rooms
	stmt, names := qb.Select(tableName).
		ToCql()
	query := q.session.Query(stmt, names)
	if err := query.SelectRelease(&res); err != nil {
		return []model.Rooms{}, err
	}
	return res, nil
}

func (q *Queries) UpdateRoom(req room_req.UpdateRoomReq) error {
	_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	tableName := fmt.Sprintf("%s.rooms", q.keyspace)
	room, err := q.GetRoomByOption(req.OldName, "name")
	if err != nil {
		return err
	}
	checkRoom, err := q.GetRoomByOption(req.Name, "name")
	if err != nil {
		return err
	}
	if len(checkRoom) > 0 && checkRoom[0].Id != room[0].Id {
		return errors.New("duplicate room's name")
	}
	update := &scylladb.UpdateRoomReq{
		Name:          req.Name,
		BedNumber:     req.BedNumber,
		PatientNumber: req.PatientNumber,
		Id:            room[0].Id.String(),
		UpdateAt:      time.Now(),
	}
	stmt, names := qb.Update(tableName).
		Set("name").
		Set("bed_number").
		Set("patient_number").
		Set("update_at").
		Where(qb.Eq("id")).
		ToCql()

	query := q.session.Query(stmt, names).BindStruct(update)
	if err := query.ExecRelease(); err != nil {
		return err
	}
	return nil
}

func (q *Queries) UpdateNumber(value int, option string, roomName string) error {
	_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	tableName := fmt.Sprintf("%s.rooms", q.keyspace)
	checkRoom, err := q.GetRoomByOption(roomName, "name")
	if err != nil {
		return err
	}
	if len(checkRoom) == 0 {
		return errors.New("No room data found")
	}
	if option == "bed_number" {
		update := &scylladb.UpdateBedNumber{
			BedNumber: checkRoom[0].BedNumber + value,
			Id:        checkRoom[0].Id.String(),
			UpdateAt:  time.Now(),
		}
		stmt, names := qb.Update(tableName).
			Set("bed_number").
			Set("update_at").
			Where(qb.Eq("id")).
			ToCql()

		query := q.session.Query(stmt, names).BindStruct(update)
		if err := query.ExecRelease(); err != nil {
			return err
		}
		return nil
	}
	update := &scylladb.UpdatePatientNumber{
		PatientNumber: checkRoom[0].PatientNumber + value,
		Id:            checkRoom[0].Id.String(),
		UpdateAt:      time.Now(),
	}
	stmt, names := qb.Update(tableName).
		Set("patient_number").
		Set("update_at").
		Where(qb.Eq("id")).
		ToCql()

	query := q.session.Query(stmt, names).BindStruct(update)
	if err := query.ExecRelease(); err != nil {
		return err
	}
	return nil
}

func (q *Queries) HandoverRoom(name string, doctorCode string) error {
	_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	tableName := fmt.Sprintf("%s.rooms", q.keyspace)
	checkRoom, err := q.GetRoomByOption(name, "name")
	if err != nil {
		return err
	}
	user, err := q.GetUserByOption(doctorCode, "doctor_code")
	if err != nil {
		return err
	}
	update := &scylladb.HandOver{
		IdDoctor: user[0].Id.String(),
		UpdateAt: time.Now(),
		Id:       checkRoom[0].Id.String(),
	}
	stmt, names := qb.Update(tableName).
		Set("id_doctor").
		Set("update_at").
		Where(qb.Eq("id")).
		ToCql()

	query := q.session.Query(stmt, names).BindStruct(update)
	if err := query.ExecRelease(); err != nil {
		return err
	}
	return nil
}
func (q *Queries) GetAllRecordInRoom(roomName string) ([]res.ShortRecord, error) {
	_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var result []res.ShortRecord

	beds, err := q.SelectAllBedsFromRoom(roomName)
	if err != nil {
		return nil, err
	}
	if len(beds) == 0 {
		return nil, errors.New("No beds data found")
	}

	// Channel để thu thập kết quả
	resultChan := make(chan res.ShortRecord, len(beds))
	errChan := make(chan error, len(beds))
	defer close(resultChan)
	defer close(errChan)

	for _, bed := range beds {
		go func(bed model.Beds) {
			if bed.Status == "UNAVAILABLE" {
				usagebed, err := q.GetUsageBedByOptionAndUnavalible(bed.Id.String(), "id_bed")
				if err != nil {
					errChan <- err
					return
				}
				if len(usagebed) == 0 {
					errChan <- errors.New("No usagebed data found")
					return
				}
				record, err := q.GetRecordWithGoRoutine(usagebed[0].IdRecord.String())
				if err != nil {
					errChan <- err
					return
				}
				shortRecord := res.ShortRecord{
					IdRecord:     record.Id,
					PatientCode:  record.Patient.PatientCode,
					PatientName:  record.Patient.Fullname,
					RoomName:     roomName,
					BedName:      bed.Name,
					DeviceSerial: record.CurrDevice.Serial,
					BedStt:       bed.Status,
					Contact:      record.Patient.Phone,
					More:         record.Patient.Reason,
				}
				resultChan <- shortRecord
			} else {
				nullShortRecord := res.ShortRecord{
					IdRecord:     "",
					PatientCode:  "",
					PatientName:  "",
					RoomName:     roomName,
					BedName:      bed.Name,
					BedStt:       bed.Status,
					DeviceSerial: "",
					Contact:      "",
					More:         "",
				}
				resultChan <- nullShortRecord
			}
		}(bed)
	}

	// Thu thập kết quả
	for i := 0; i < len(beds); i++ {
		select {
		case res := <-resultChan:
			result = append(result, res)
		case err := <-errChan:
			return nil, err
		}
	}

	return result, nil
}
func (q *Queries) GetAllRecordInRoomPagination(roomName string, page, pageSize int) ([]res.ShortRecord, error, int) {
	_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var result []res.ShortRecord

	beds, err := q.SelectAllBedsFromRoom(roomName)
	if err != nil {
		return nil, err, 0
	}
	if len(beds) == 0 {
		return nil, errors.New("No beds data found"), 0
	}
	maxPage := len(beds) / pageSize
	if maxPage == 0 {
		maxPage = 1
	}
	// Tạo slice riêng biệt cho giường UNAVAILABLE và giường khác
	var unavailableBeds []model.Beds
	var otherBeds []model.Beds

	for _, bed := range beds {
		if bed.Status == "UNAVAILABLE" {
			unavailableBeds = append(unavailableBeds, bed)
		} else {
			otherBeds = append(otherBeds, bed)
		}
	}

	sort.Slice(unavailableBeds, func(i, j int) bool {
		return unavailableBeds[i].Name < unavailableBeds[j].Name
	})

	// Kết hợp hai slice lại với nhau, giường UNAVAILABLE đặt trước
	sortedBeds := append(unavailableBeds, otherBeds...)

	// Tính toán phạm vi dữ liệu cần trả về
	start := (page - 1) * pageSize
	end := start + pageSize
	if start > len(sortedBeds) {
		return nil, errors.New("Page number out of range"), 0
	}
	if end > len(sortedBeds) {
		end = len(sortedBeds)
	}
	pagedBeds := sortedBeds[start:end]

	// Channel để thu thập kết quả
	resultChan := make(chan res.ShortRecord, len(pagedBeds))
	errChan := make(chan error, len(pagedBeds))
	defer close(resultChan)
	defer close(errChan)

	for _, bed := range pagedBeds {
		go func(bed model.Beds) {
			if bed.Status == "UNAVAILABLE" {
				usagebed, err := q.GetUsageBedByOptionAndUnavalible(bed.Id.String(), "id_bed")
				if err != nil {
					errChan <- err
					return
				}
				if len(usagebed) == 0 {
					errChan <- errors.New("No usagebed data found")
					return
				}
				var record res.RecordRes
				for _, element := range usagebed {
					if element.Status == "IN_USE" {
						record, err = q.GetRecordWithGoRoutine(usagebed[0].IdRecord.String())
						if err != nil {
							errChan <- err
							return
						}
						shortRecord := res.ShortRecord{
							IdRecord:     record.Id,
							PatientCode:  record.Patient.PatientCode,
							PatientName:  record.Patient.Fullname,
							RoomName:     roomName,
							BedName:      bed.Name,
							DeviceSerial: record.CurrDevice.Serial,
							Status:       record.Status,
							BedStt:       bed.Status,
							Contact:      record.Patient.Phone,
							More:         record.Patient.Reason,
						}
						resultChan <- shortRecord
					}
				}

			} else {
				nullShortRecord := res.ShortRecord{
					IdRecord:     "",
					PatientCode:  "",
					PatientName:  "",
					RoomName:     roomName,
					BedName:      bed.Name,
					BedStt:       bed.Status,
					DeviceSerial: "",
					Status:       "",
					Contact:      "",
					More:         "",
				}
				resultChan <- nullShortRecord
			}
		}(bed)
	}

	// Thu thập kết quả
	for i := 0; i < len(pagedBeds); i++ {
		select {
		case res := <-resultChan:
			result = append(result, res)
		case err := <-errChan:
			return nil, err, 0
		}
	}

	return result, nil, maxPage
}

func (q *Queries) HandoverRoomForNormalDoctor(idRoom string, idDoctor string) error {
	_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	tableName := fmt.Sprintf("%s.usage_room", q.keyspace)
	id, err := gocql.ParseUUID(uuid.New().String())
	if err != nil {
		panic(err)
	}
	insert := &model.UsageRoomDoctors{
		Id:       id,
		IdRoom:   idRoom,
		IdDoctor: idDoctor,
		CreateAt: time.Now(),
	}
	stmt := qb.Insert(tableName).
		Columns("id", "id_room", "id_doctor", "create_at").
		Query(q.session)
	stmt.BindStruct(insert)
	if err := stmt.ExecRelease(); err != nil {
		return err
	}
	return nil
}

func (q *Queries) GetUsageBedByIdDoctorAndIdRoom(idDoctor, idRoom string) (model.UsageRoomDoctors, error) {
	_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var usageRoom model.UsageRoomDoctors
	tableName := fmt.Sprintf("%s.usage_room", q.keyspace)

	stmt, names := qb.Select(tableName).
		Where(qb.Eq("id_doctor"), qb.Eq("id_room")).
		AllowFiltering().
		ToCql()

	query := q.session.Query(stmt, names).BindMap(qb.M{
		"id_doctor": idDoctor,
		"id_room":   idRoom,
	})

	if err := query.GetRelease(&usageRoom); err != nil {
		return model.UsageRoomDoctors{}, err
	}

	return usageRoom, nil
}

func (q *Queries) DeleteDoctorFromRoom(idRoom string, idDoctor string) error {
	_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	tableName := fmt.Sprintf("%s.usage_room", q.keyspace)
	usageRoom, err := q.GetUsageBedByIdDoctorAndIdRoom(idDoctor, idRoom)
	if err != nil {
		fmt.Println(err)
		return err
	}
	update := &scylladb.HandOver{
		IdDoctor: "00000000-0000-0000-0000-000000000000",
		Id:       usageRoom.Id.String(),
	}
	stmt, names := qb.Update(tableName).
		Set("id_doctor").
		Where(qb.Eq("id")).
		ToCql()

	query := q.session.Query(stmt, names).BindStruct(update)
	if err := query.ExecRelease(); err != nil {
		return err
	}
	return nil
}

func (q *Queries) SelectUsageRoomByOption(option string, value string) ([]model.UsageRoomDoctors, error) {
	_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	tableName := fmt.Sprintf("%s.usage_room", q.keyspace)
	var usage_room []model.UsageRoomDoctors
	stmt, names := qb.Select(tableName).
		Where(qb.Eq(option)).
		ToCql()
	stmt += " ALLOW FILTERING"
	query := q.session.Query(stmt, names).BindMap(qb.M{
		option: value,
	})
	if err := query.SelectRelease(&usage_room); err != nil {
		return nil, err
	}
	return usage_room, nil
}

func (q *Queries) SelectAllRoomByCurrDoctor(c echo.Context) ([]model.Rooms, error) {
	doctor, err := q.GetProfileCurrent(c)
	var result []model.Rooms
	if err != nil {
		return nil, err
	}
	records, err := q.SelectUsageRoomByOption("id_doctor", doctor.Id.String())
	if err != nil {
		return nil, err
	}
	for _, record := range records {
		room, err := q.GetRoomByOption(record.IdRoom, "id")
		if err != nil {
			return nil, err
		}
		if len(room) == 0 {
			return nil, errors.New("no room data found")
		}
		result = append(result, room[0])
	}
	return result, nil
}

func (q *Queries) CheckPermissionInRoomById(idRoom string, c echo.Context) bool {
	rooms, err := q.SelectAllRoomByCurrDoctor(c)
	var result = false
	if err != nil {
		return false
	}
	for _, room := range rooms {
		if room.Id.String() == idRoom {
			result = true
		}
	}
	return result
}

func (q *Queries) CheckPermissionInRoomByName(nameRoom string, c echo.Context) bool {
	rooms, err := q.SelectAllRoomByCurrDoctor(c)
	var result = false
	if err != nil {
		return false
	}
	for _, room := range rooms {
		if room.Name == nameRoom {
			result = true
			break
		}
	}
	return result
}

func (q *Queries) GetRoomDetail(roomName string) (res.RoomDetailRes, error) {
	room, err := q.GetRoomByOption(roomName, "name")
	var member []model.Users
	if err != nil {
		return res.RoomDetailRes{}, err
	}
	if len(room) == 0 {
		return res.RoomDetailRes{}, errors.New("no room data found")
	}
	fmt.Println(room[0].Id.String())
	usagesRoom, err := q.SelectUsageRoomByOption("id_room", room[0].Id.String())
	if err != nil {
		fmt.Println(err)
		return res.RoomDetailRes{}, err
	}
	if len(usagesRoom) == 0 {
		return res.RoomDetailRes{}, errors.New("no usageRoom data found")
	}
	for _, usageRoom := range usagesRoom {
		mem, _ := q.GetUserByOption(usageRoom.IdDoctor, "id")
		if len(mem) > 0 {
			member = append(member, mem[0])
		}
	}
	leader, err := q.GetUserByOption(room[0].IdDoctor.String(), "id")
	if err != nil {
		return res.RoomDetailRes{}, err
	}
	if len(leader) == 0 {
		return res.RoomDetailRes{}, errors.New("no member data found")
	}
	return res.RoomDetailRes{
		Id:            room[0].Id.String(),
		PatientNumber: room[0].PatientNumber,
		BedNumber:     room[0].BedNumber,
		Leader:        leader[0],
		Members:       member,
	}, nil
}
