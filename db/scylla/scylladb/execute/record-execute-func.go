package execute

import (
	"HospitalManager/db/scylla/scylladb"
	"HospitalManager/dto/req/record_req"
	"HospitalManager/dto/res"
	"HospitalManager/helper"
	"HospitalManager/model"
	"context"
	"errors"
	"fmt"
	"github.com/gocql/gocql"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/scylladb/gocqlx/v2/qb"
	"regexp"
	"sort"
	"sync"
	"time"
)

func (q *Queries) InsertRecord(req record_req.InsertRecordReq, c echo.Context) error {
	_, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()
	tableName := fmt.Sprintf("%s.medical_records", q.keyspace)
	patientCode := helper.GenPatientCode(req.Fullname, req.Phone, req.Ccid)
	idRecord, err := gocql.ParseUUID(uuid.New().String())
	if err != nil {
		panic(err)
	}
	idPatient, err := gocql.ParseUUID(uuid.New().String())
	if err != nil {
		panic(err)
	}
	doctor, err := q.GetUserByOption(req.DoctorCode, "doctor_code")
	if err != nil {
		return err
	}

	patientCheck, err := q.GetPatient(patientCode, "patient_code")
	if err != nil {
		return err
	}
	if len(patientCheck) > 0 {
		for _, patient := range patientCheck {
			record, err := q.GetRecordByOption(patient.Id.String(), "id_patient")
			if err != nil {
				return err
			}
			if len(record) > 0 {
				if record[0].Status == "PENDING" || record[0].Status == "TREATING" {
					return errors.New("Duplicate Medical record")
				}
			}
		}
	}
	updater, err := q.GetProfileCurrent(c)
	if err != nil {
		return err
	}
	insert := &model.MedicalRecords{
		Id:        idRecord,
		IdPatient: idPatient,
		IdDoctor:  doctor[0].Id,
		IdUpdater: updater.Id,
		Status:    "PENDING",
		CreateAt:  time.Now(),
		UpdateAt:  time.Now(),
	}
	stmt := qb.Insert(tableName).
		Columns("id", "id_patient", "id_doctor", "id_updater", "status", "create_at", "update_at").
		Query(q.session)
	stmt.BindStruct(insert)
	if err := stmt.ExecRelease(); err != nil {
		return err
	}
	err = q.InsertPatient(req, idPatient, patientCode)
	if err != nil {
		return err
	}
	return nil
}

func (q *Queries) GetRecordByOption(value string, option string) ([]model.MedicalRecords, error) {
	_, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()
	tableName := fmt.Sprintf("%s.medical_records", q.keyspace)
	var records []model.MedicalRecords
	stmt, names := qb.Select(tableName).
		Where(qb.Eq(option)).
		ToCql()
	stmt += " ALLOW FILTERING"
	query := q.session.Query(stmt, names).BindMap(qb.M{
		option: value,
	})
	if err := query.SelectRelease(&records); err != nil {
		return nil, err
	}
	return records, nil
}

func (q *Queries) UpdateUpdater(c echo.Context, id string) error {
	_, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()
	user, err := q.GetProfileCurrent(c)
	if err != nil {
		return err
	}
	tableName := fmt.Sprintf("%s.medical_records", q.keyspace)
	update := &scylladb.UpdateUpdaterRecord{
		IdUpdater: user.Id.String(),
		Id:        id,
		UpdateAt:  time.Now(),
	}
	stmt, names := qb.Update(tableName).
		Set("id_updater").
		Set("update_at").
		Where(qb.Eq("id")).
		ToCql()

	query := q.session.Query(stmt, names).BindStruct(update)
	if err := query.ExecRelease(); err != nil {
		return err
	}
	return nil
}

func (q *Queries) ChangeRecordStatus(id string, status string) error {
	_, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()
	tableName := fmt.Sprintf("%s.medical_records", q.keyspace)
	update := &scylladb.UpdateRecordStt{
		Status:   status,
		Id:       id,
		UpdateAt: time.Now(),
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

func (q *Queries) CreateRecordHistory(idRecord gocql.UUID, content string, c echo.Context) error {
	_, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()
	id, err := gocql.ParseUUID(uuid.New().String())
	if err != nil {
		return err
	}
	user, err := q.GetProfileCurrent(c)
	if err != nil {
		return err
	}
	tableName := fmt.Sprintf("%s.record_history", q.keyspace)
	insert := &model.RecordHistory{
		Id:       id,
		IdRecord: idRecord,
		IdDoctor: user.Id,
		Content:  content,
		CreateAt: time.Now(),
	}
	stmt := qb.Insert(tableName).
		Columns("id", "id_record", "id_doctor", "content", "create_at").
		Query(q.session)
	stmt.BindStruct(insert)
	if err := stmt.ExecRelease(); err != nil {
		return err
	}
	return nil
}

func (q *Queries) CreateRecordHistoryStr(idRecord string, content string, c echo.Context) error {
	_, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()
	id, err := gocql.ParseUUID(uuid.New().String())
	if err != nil {
		return err
	}
	user, err := q.GetProfileCurrent(c)
	if err != nil {
		return err
	}
	tableName := fmt.Sprintf("%s.record_history", q.keyspace)
	insert := &model.RecordHistoryStr{
		Id:       id,
		IdRecord: idRecord,
		IdDoctor: user.Id,
		Content:  content,
		CreateAt: time.Now(),
	}
	stmt := qb.Insert(tableName).
		Columns("id", "id_record", "id_doctor", "content", "create_at").
		Query(q.session)
	stmt.BindStruct(insert)
	if err := stmt.ExecRelease(); err != nil {
		return err
	}
	return nil
}

func (q *Queries) GetHistoryById(idRecord string) ([]model.RecordHistory, error) {
	_, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()
	tableName := fmt.Sprintf("%s.record_history", q.keyspace)
	var records []model.RecordHistory
	stmt, names := qb.Select(tableName).
		Where(qb.Eq("id_record")).
		ToCql()
	stmt += " ALLOW FILTERING"
	query := q.session.Query(stmt, names).BindMap(qb.M{
		"id_record": idRecord,
	})
	if err := query.SelectRelease(&records); err != nil {
		return nil, err
	}
	if len(records) == 0 {
		return nil, errors.New("No Record History Data Found")
	}
	return records, nil
}

//
//func (q *Queries) GetRecord(id string, c echo.Context) (res.RecordRes, error) {
//	_, cancel := context.WithTimeout(context.Background(), 100*time.Second)
//	defer cancel()
//	record, err := q.GetRecordByOption(id, "id")
//	if err != nil {
//		return res.RecordRes{}, err
//	}
//	if len(record) == 0 {
//		return res.RecordRes{}, errors.New("No record data found")
//	}
//	patient, err := q.GetPatient(record[0].IdPatient.String(), "id")
//	if err != nil {
//		return res.RecordRes{}, err
//	}
//	if len(patient) == 0 {
//		return res.RecordRes{}, errors.New("No patient data found")
//	}
//	doctor, err := q.GetUserByOption(record[0].IdDoctor.String(), "id")
//	if err != nil {
//		return res.RecordRes{}, err
//	}
//	if len(doctor) == 0 {
//		return res.RecordRes{}, errors.New("No doctor data found")
//	}
//	updater, err := q.GetUserByOption(record[0].IdUpdater.String(), "id")
//	if err != nil {
//		return res.RecordRes{}, err
//	}
//	if len(updater) == 0 {
//		return res.RecordRes{}, errors.New("No updater data found")
//	}
//	history, err := q.GetHistoryById(id)
//	if err != nil {
//		return res.RecordRes{}, err
//	}
//	if len(history) == 0 {
//		return res.RecordRes{}, errors.New("No history data found")
//	}
//	notes, err := q.GetAllNote(id)
//	if err != nil {
//		return res.RecordRes{}, err
//	}
//	if len(notes) == 0 {
//		return res.RecordRes{}, errors.New("No notes data found")
//	}
//	usageBeds, err := q.GetUsageBedByOption(id, "id_record")
//	if err != nil {
//		return res.RecordRes{}, err
//	}
//	if len(usageBeds) == 0 {
//		return res.RecordRes{}, errors.New("No usageBed data found")
//	}
//	var beds []model.Beds
//	for _, usageBed := range usageBeds {
//		bed, err := q.GetBedById(usageBed.IdBed.String())
//		if err != nil {
//			return res.RecordRes{}, err
//		}
//		beds = append(beds, bed)
//	}
//
//	usageDevices, err := q.GetUsageDeviceByOption(id, "id_record")
//	if err != nil {
//		return res.RecordRes{}, err
//	}
//	if len(usageDevices) == 0 {
//		return res.RecordRes{}, errors.New("No usageDevices data found")
//	}
//	var devices []model.Devices
//	for _, usageDevice := range usageDevices {
//		device, err := q.GetDeviceById(usageDevice.IdDevice.String())
//		if err != nil {
//			return res.RecordRes{}, err
//		}
//		devices = append(devices, device)
//	}
//
//	result := res.RecordRes{
//		Id:       id,
//		Patient:  patient[0],
//		Doctor:   doctor[0],
//		Updater:  updater[0],
//		Notes:    notes,
//		Beds:     beds,
//		Devices:  devices,
//		History:  history,
//		CreateAt: record[0].CreateAt,
//		UpdateAt: record[0].UpdateAt,
//	}
//
//	return result, nil
//}

func (q *Queries) mapHistoryToHistoryRes(histories []model.RecordHistory) []res.HistoryRes {
	var result []res.HistoryRes
	for _, history := range histories {
		user, _ := q.GetUserByOption(history.IdDoctor.String(), "id")
		res := res.HistoryRes{
			Id:       history.Id.String(),
			Doctor:   user[0],
			Content:  history.Content,
			CreateAt: history.CreateAt,
		}
		result = append(result, res)
	}
	return result
}

func (q *Queries) mapNoteToNoteRes(notes []model.Notes) []res.Note {
	var responses []res.Note
	for _, note := range notes {
		doctor, _ := q.GetUserByOption(note.IdDoctor.String(), "id")
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

func (q *Queries) GetRecordWithGoRoutine(id string) (res.RecordRes, error) {
	// Tạo context với hạn chế thời gian
	_, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	// Tạo các kênh để nhận kết quả từ các goroutine
	recordCh := make(chan []model.MedicalRecords, 1)
	patientCh := make(chan []model.Patients, 1)
	doctorCh := make(chan []model.Users, 1)
	updaterCh := make(chan []model.Users, 1)
	historyCh := make(chan []model.RecordHistory, 1)
	notesCh := make(chan []model.Notes, 1)
	//bedsCh := make(chan []model.Beds, 1)
	//devicesCh := make(chan []model.Devices, 1)
	usageBedsCh := make(chan []model.UsageBed, 1)
	usageDevicesCh := make(chan []model.UsageDevice, 1)

	// Thực hiện các truy vấn trong các goroutine riêng biệt
	go func() {
		record, err := q.GetRecordByOption(id, "id")
		if err != nil {
			recordCh <- nil
			return
		}
		recordCh <- record
	}()

	// Lấy kết quả từ các goroutine và xử lý
	record := <-recordCh
	if len(record) == 0 {
		return res.RecordRes{}, errors.New("No record data found")
	}

	go func() {
		patient, err := q.GetPatient(record[0].IdPatient.String(), "id")
		if err != nil {
			patientCh <- nil
			return
		}
		patientCh <- patient
	}()

	go func() {
		doctor, err := q.GetUserByOption(record[0].IdDoctor.String(), "id")
		if err != nil {
			doctorCh <- nil
			return
		}
		doctorCh <- doctor
	}()

	go func() {
		updater, err := q.GetUserByOption(record[0].IdUpdater.String(), "id")
		if err != nil {
			updaterCh <- nil
			return
		}
		updaterCh <- updater
	}()

	go func() {
		history, err := q.GetHistoryById(id)
		if err != nil {
			historyCh <- nil
			return
		}
		historyCh <- history
	}()

	go func() {
		notes, err := q.GetAllNote(id)
		if err != nil {
			notesCh <- nil
			return
		}
		notesCh <- notes
	}()

	go func() {
		usageBeds, err := q.GetUsageBedByOption(id, "id_record")
		if err != nil {
			usageBedsCh <- nil
			return
		}
		usageBedsCh <- usageBeds
	}()

	go func() {
		usageDevices, err := q.GetUsageDeviceByOption(id, "id_record")
		if err != nil {
			usageDevicesCh <- nil
			return
		}
		usageDevicesCh <- usageDevices
	}()

	patient := <-patientCh
	if len(patient) == 0 {
		return res.RecordRes{}, errors.New("No patient data found")
	}

	doctor := <-doctorCh
	if len(doctor) == 0 {
		return res.RecordRes{}, errors.New("No doctor data found")
	}

	updater := <-updaterCh
	if len(updater) == 0 {
		return res.RecordRes{}, errors.New("No updater data found")
	}

	var historyRes []res.HistoryRes
	history := <-historyCh
	if len(history) == 0 {
		historyRes = nil
	}

	notes := <-notesCh
	if len(notes) == 0 {
		notes = nil
	}

	usageBeds := <-usageBedsCh
	if len(usageBeds) == 0 {
	}

	usageDevices := <-usageDevicesCh
	if len(usageDevices) == 0 {
	}

	// Lặp qua các usageBeds để lấy thông tin về Beds
	var beds []res.BedRes
	for _, usageBed := range usageBeds {
		bed, err := q.GetBedById(usageBed.IdBed.String())
		if err != nil {
			return res.RecordRes{}, err
		}
		room, err := q.GetRoomByOption(bed.IdRoom.String(), "id")
		newBedRes := res.BedRes{
			Id:         bed.Id,
			Name:       bed.Name,
			Status:     bed.Status,
			CreateAt:   bed.CreateAt,
			UpdateAt:   bed.UpdateAt,
			InuseAt:    usageBed.CreateAt,
			NotInuseAt: usageBed.EndAt,
			Room:       room[0],
		}
		beds = append(beds, newBedRes)
	}
	var latestDevice model.Devices

	var DeviceHistory []res.DeviceHistory
	// Lặp qua các usageDevices để lấy thông tin về Devices
	if len(usageDevices) > 0 {
		for _, usageDevice := range usageDevices {
			device, err := q.GetDeviceById(usageDevice.IdDevice.String())
			if err != nil {
				return res.RecordRes{}, err
			}
			deviceRes := res.DeviceHistory{
				Device:     device,
				InuseAt:    usageDevice.CreateAt,
				NotInuseAt: usageDevice.EndAt,
			}
			DeviceHistory = append(DeviceHistory, deviceRes)
		}

		sort.Slice(usageDevices, func(i, j int) bool {
			return usageDevices[i].CreateAt.After(usageDevices[j].CreateAt)
		})
		// Lấy device mới nhất từ danh sách đã sắp xếp
		latestDeviceDB, err := q.GetDeviceById(usageDevices[0].IdDevice.String())
		if err != nil {
			return res.RecordRes{}, err
		}
		if latestDeviceDB.Status == "IN_STORAGE" {
			latestDeviceDB = model.Devices{}
		}
		latestDevice = latestDeviceDB
	}

	var lastestBedRes res.BedRes

	if record[0].Status == "LEAVED" {
		lastestBedRes = res.BedRes{}
	} else {
		if len(usageBeds) > 0 {
			var latestBed model.Beds
			sort.Slice(usageBeds, func(i, j int) bool {
				return usageBeds[i].CreateAt.After(usageBeds[j].CreateAt)
			})
			// Lấy bed mới nhất từ danh sách đã sắp xếp
			latestBed, err := q.GetBedById(usageBeds[0].IdBed.String())
			if err != nil {
				return res.RecordRes{}, err
			}
			room, _ := q.GetRoomByOption(latestBed.IdRoom.String(), "id")

			lastestBedRes.Id = latestBed.Id
			lastestBedRes.Name = latestBed.Name
			lastestBedRes.Status = latestBed.Status
			lastestBedRes.CreateAt = latestBed.CreateAt
			lastestBedRes.UpdateAt = latestBed.UpdateAt
			lastestBedRes.Room = room[0]

			if lastestBedRes.Status == "AVAILABLE" {
				lastestBedRes = res.BedRes{}
			}
		}
	}

	historyRes = q.mapHistoryToHistoryRes(history)
	noteRes := q.mapNoteToNoteRes(notes)

	// Tạo và trả về kết quả cuối cùng
	result := res.RecordRes{
		Id:         id,
		Status:     record[0].Status,
		Patient:    patient[0],
		Doctor:     doctor[0],
		Updater:    updater[0],
		Notes:      noteRes,
		Beds:       beds,
		Devices:    DeviceHistory,
		CurrBed:    lastestBedRes,
		CurrDevice: latestDevice,
		History:    historyRes,
		CreateAt:   record[0].CreateAt,
		UpdateAt:   record[0].UpdateAt,
	}

	return result, nil
}

func (q *Queries) GetAllPendingRecord(c echo.Context) ([]res.PendingRecordRes, error) {
	_, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()
	doctor, err := q.GetProfileCurrent(c)
	if err != nil {
		return nil, err
	}
	records, err := q.GetRecordByOption("PENDING", "status")
	if err != nil {
		return nil, err
	}
	if len(records) == 0 {
		return nil, errors.New("no record data found")
	}
	var recordRes []res.PendingRecordRes
	for _, record := range records {
		patient, err := q.GetPatient(record.IdPatient.String(), "id")
		if err != nil {
			return nil, err
		}
		if len(patient) == 0 {
			return nil, errors.New("No patient data found")
		}
		res := res.PendingRecordRes{
			Id:          record.Id.String(),
			DoctorCode:  doctor.DoctorCode,
			PatientCode: patient[0].PatientCode,
			Fullname:    patient[0].Fullname,
			Phone:       patient[0].Phone,
			Detail:      patient[0].Reason,
			CreateAt:    record.CreateAt,
		}
		recordRes = append(recordRes, res)
	}
	return recordRes, nil
}

func (q *Queries) GetAllPatientByDoctor(c echo.Context) ([]res.TotalRecordRes, error) {
	_, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	var resp []res.TotalRecordRes
	var mu sync.Mutex     // To protect concurrent writes to resp
	var wg sync.WaitGroup // To wait for all go routines to complete
	var records []model.MedicalRecords
	rooms, err := q.SelectAllRoomByCurrDoctor(c)
	if err != nil {
		return nil, err
	}

	for _, room := range rooms {
		record, err := q.GetRecordByOption(room.Id.String(), "id_room")
		if err != nil {
			return nil, err
		}
		if len(record) > 0 {
			for _, temp := range record {
				records = append(records, temp)
			}
		}
	}

	// Channel to collect errors
	errChan := make(chan error, len(records))
	defer close(errChan)

	for _, record := range records {
		wg.Add(1)
		go func(record model.MedicalRecords) {
			defer wg.Done()
			patient, err := q.GetPatient(record.IdPatient.String(), "id")
			if err != nil {
				errChan <- err
				return
			}
			temp := res.TotalRecordRes{
				Id:          record.Id.String(),
				PatientCode: patient[0].PatientCode,
				Fullname:    patient[0].Fullname,
				Address:     patient[0].Address,
				Phone:       patient[0].Phone,
				Status:      record.Status,
				More:        patient[0].Reason,
			}

			mu.Lock()
			resp = append(resp, temp)
			mu.Unlock()
		}(record)
	}

	// Wait for all goroutines to finish
	wg.Wait()

	// Check if there were any errors
	select {
	case err := <-errChan:
		return nil, err
	default:
	}

	return resp, nil
}

func (q *Queries) getAllRecord() ([]model.MedicalRecords, error) {
	_, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()
	tableName := fmt.Sprintf("%s.medical_records", q.keyspace)
	var records []model.MedicalRecords
	stmt, names := qb.Select(tableName).
		ToCql()
	query := q.session.Query(stmt, names)
	if err := query.SelectRelease(&records); err != nil {
		return nil, err
	}
	return records, nil
}

func (q *Queries) GetAllPatientByAdmin(c echo.Context) ([]res.TotalRecordRes, error) {
	_, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	var resp []res.TotalRecordRes
	var mu sync.Mutex     // To protect concurrent writes to resp
	var wg sync.WaitGroup // To wait for all go routines to complete
	var records []model.MedicalRecords
	records, err := q.getAllRecord()
	if err != nil {
		return nil, err
	}
	if len(records) == 0 {
		return nil, errors.New("no data record found")
	}

	// Channel to collect errors
	errChan := make(chan error, len(records))
	defer close(errChan)

	for _, record := range records {
		wg.Add(1)
		go func(record model.MedicalRecords) {
			defer wg.Done()
			patient, err := q.GetPatient(record.IdPatient.String(), "id")
			if err != nil {
				errChan <- err
				return
			}
			temp := res.TotalRecordRes{
				Id:          record.Id.String(),
				PatientCode: patient[0].PatientCode,
				Fullname:    patient[0].Fullname,
				Address:     patient[0].Address,
				Phone:       patient[0].Phone,
				Status:      record.Status,
				More:        patient[0].Reason,
			}

			mu.Lock()
			resp = append(resp, temp)
			mu.Unlock()
		}(record)
	}

	// Wait for all goroutines to finish
	wg.Wait()

	// Check if there were any errors
	select {
	case err := <-errChan:
		return nil, err
	default:
	}

	return resp, nil
}

func (q *Queries) searchTotalPatient(search string) ([]model.Patients, error) {
	_, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()
	tableName := fmt.Sprintf("%s.patients", q.keyspace)
	searchStr := "%" + search + "%"
	var records []model.Patients
	fmt.Println(searchStr)
	re := regexp.MustCompile(`\d|^BN`)
	if re.MatchString(search) {
		stmt, names := qb.Select(tableName).
			Where(qb.Like("patient_code")).
			AllowFiltering().
			ToCql()
		query := q.session.Query(stmt, names).BindMap(qb.M{
			"patient_code": searchStr,
		})
		fmt.Println(stmt)
		if err := query.SelectRelease(&records); err != nil {
			return nil, err
		}
		if len(records) == 0 {
			return nil, errors.New("No Record Data Found")
		}
		return records, nil
	}

	stmt, names := qb.Select(tableName).
		Where(qb.Like("fullname")).
		AllowFiltering().
		ToCql()
	query := q.session.Query(stmt, names).BindMap(qb.M{
		"fullname": searchStr,
	})
	if err := query.SelectRelease(&records); err != nil {
		return nil, err
	}
	if len(records) == 0 {
		return nil, errors.New("No Record Data Found")
	}
	return records, nil
}

func (q *Queries) SearchTotalRecord(search string, c echo.Context) ([]res.TotalRecordRes, error) {
	_, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()
	patients, err := q.searchTotalPatient(search)
	if err != nil {
		return nil, err
	}
	rooms, err := q.SelectAllRoomByCurrDoctor(c)
	if err != nil {
		return nil, err
	}
	var result []res.TotalRecordRes
	for _, patient := range patients {
		records, err := q.GetRecordByOption(patient.Id.String(), "id_patient")
		if err != nil {
			return nil, err
		}
		if len(records) == 0 {
			return nil, errors.New("No record data found")
		}
		if len(records) > 1 {
			for _, record := range records {
				for _, room := range rooms {
					if record.IdRoom == room.Id.String() {
						resp := res.TotalRecordRes{
							Id:          record.Id.String(),
							PatientCode: patient.PatientCode,
							Fullname:    patient.Fullname,
							Address:     patient.Address,
							Phone:       patient.Phone,
							Status:      record.Status,
							More:        patient.Reason,
						}
						result = append(result, resp)
					}
				}
			}
		}
		for _, room := range rooms {
			if records[0].IdRoom == room.Id.String() {
				resp := res.TotalRecordRes{
					Id:          records[0].Id.String(),
					PatientCode: patient.PatientCode,
					Fullname:    patient.Fullname,
					Address:     patient.Address,
					Phone:       patient.Phone,
					Status:      records[0].Status,
					More:        patient.Reason,
				}
				result = append(result, resp)
			}
		}
	}
	if len(result) == 0 {
		return nil, errors.New("No record data found")
	}
	return result, nil
}

func (q *Queries) UpdateRoomForRecord(idRoom, id string) error {
	_, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()
	tableName := fmt.Sprintf("%s.medical_records", q.keyspace)
	update := &model.UpdateRoomForRecord{
		IdRoom: idRoom,
		Id:     id,
	}
	stmt, names := qb.Update(tableName).
		Set("id_room").
		Where(qb.Eq("id")).
		ToCql()

	query := q.session.Query(stmt, names).BindStruct(update)
	if err := query.ExecRelease(); err != nil {
		return err
	}
	return nil
}

func (q *Queries) SearchPendingRecord(search string, c echo.Context) ([]res.PendingRecordRes, error) {
	_, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()
	patients, err := q.searchTotalPatient(search)
	if err != nil {
		return nil, err
	}
	var result []res.PendingRecordRes
	for _, patient := range patients {
		records, err := q.GetRecordByOption(patient.Id.String(), "id_patient")
		if err != nil {
			return nil, err
		}
		if len(records) == 0 {
			return nil, errors.New("No record data found")
		}
		if len(records) > 1 {
			for _, record := range records {
				if record.Status == "PENDING" {
					resp := res.PendingRecordRes{
						Id:          record.Id.String(),
						PatientCode: patient.PatientCode,
						Fullname:    patient.Fullname,
						Phone:       patient.Phone,
						Detail:      patient.Reason,
						CreateAt:    record.CreateAt,
					}
					result = append(result, resp)
				}
			}
		}
		if records[0].Status == "PENDING" {
			resp := res.PendingRecordRes{
				Id:          records[0].Id.String(),
				PatientCode: patient.PatientCode,
				Fullname:    patient.Fullname,
				Phone:       patient.Phone,
				Detail:      patient.Reason,
				CreateAt:    records[0].CreateAt,
			}
			result = append(result, resp)
		}
	}
	if len(result) == 0 {
		return nil, errors.New("No record data found")
	}
	return result, nil
}

func (q *Queries) SearchRecordByAdmin(search string, c echo.Context) ([]res.TotalRecordRes, error) {
	_, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()
	patients, err := q.searchTotalPatient(search)
	if err != nil {
		return nil, err
	}
	var result []res.TotalRecordRes
	for _, patient := range patients {
		records, err := q.GetRecordByOption(patient.Id.String(), "id_patient")
		if err != nil {
			return nil, err
		}
		if len(records) == 0 {
			return nil, errors.New("No record data found")
		}
		if len(records) > 1 {
			for _, record := range records {

				resp := res.TotalRecordRes{
					Id:          record.Id.String(),
					PatientCode: patient.PatientCode,
					Fullname:    patient.Fullname,
					Address:     patient.Address,
					Phone:       patient.Phone,
					Status:      record.Status,
					More:        patient.Reason,
				}
				result = append(result, resp)
			}
		}
		resp := res.TotalRecordRes{
			Id:          records[0].Id.String(),
			PatientCode: patient.PatientCode,
			Fullname:    patient.Fullname,
			Address:     patient.Address,
			Phone:       patient.Phone,
			Status:      records[0].Status,
			More:        patient.Reason,
		}
		result = append(result, resp)
	}
	if len(result) == 0 {
		return nil, errors.New("No record data found")
	}
	return result, nil
}

func (q *Queries) StatisticalNumber(c echo.Context) (res.StatisticalRes, error) {
	type result struct {
		pendingRecords   []res.PendingRecordRes
		availableBeds    []model.Beds
		availableDevices []model.Devices
		err              error
	}

	resultsCh := make(chan result)

	go func() {
		pendingRecords, _ := q.GetAllPendingRecord(c)
		resultsCh <- result{pendingRecords: pendingRecords, err: nil}
	}()

	go func() {
		availableBeds, _ := q.GetAvailableBed(c)
		resultsCh <- result{availableBeds: availableBeds, err: nil}
	}()

	go func() {
		devices, err := q.GetAllDevice()
		if err != nil {
			resultsCh <- result{err: err}
			return
		}
		var availableDevices []model.Devices
		for _, device := range devices {
			if device.Status == "IN_STORAGE" || device.Status == "DISABLED" {
				availableDevices = append(availableDevices, device)
			}
		}
		resultsCh <- result{availableDevices: availableDevices}
	}()

	var resStatistical res.StatisticalRes
	for i := 0; i < 3; i++ {
		r := <-resultsCh
		if r.err != nil {
			return res.StatisticalRes{}, r.err
		}
		if r.pendingRecords != nil {
			resStatistical.PendingRc = len(r.pendingRecords)
		}
		if r.availableBeds != nil {
			resStatistical.AvailableBed = len(r.availableBeds)
		}
		if r.availableDevices != nil {
			resStatistical.AvailableDevice = len(r.availableDevices)
		}
	}

	return resStatistical, nil
}
