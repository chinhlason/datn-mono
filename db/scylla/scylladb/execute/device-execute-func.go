package execute

import (
	"HospitalManager/db/scylla/scylladb"
	"HospitalManager/dto/req/device_req"
	"HospitalManager/dto/res"
	"HospitalManager/model"
	"context"
	"errors"
	"fmt"
	"github.com/gocql/gocql"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/scylladb/gocqlx/v2/qb"
	"sync"
	"time"
)

func (q *Queries) AddDevice(req device_req.AddDeviceReq) error {
	_, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()
	tableName := fmt.Sprintf("%s.devices", q.keyspace)
	device, err := q.GetDeviceByOption(req.Serial, "serial")
	if err != nil {
		return err
	}
	if len(device) > 0 {
		return errors.New("duplicate device's serial")
	}
	id, err := gocql.ParseUUID(uuid.New().String())
	if err != nil {
		panic(err)
	}
	insert := &model.Devices{
		Id:       id,
		Serial:   req.Serial,
		Warraty:  req.Warraty,
		Status:   "IN_STORAGE",
		CreateAt: time.Now(),
		UpdateAt: time.Now(),
	}
	stmt := qb.Insert(tableName).
		Columns("id", "serial", "warraty", "status", "create_at", "update_at").
		Query(q.session)
	stmt.BindStruct(insert)
	if err := stmt.ExecRelease(); err != nil {
		return err
	}
	return nil
}

func (q *Queries) GetDeviceByOption(value string, option string) ([]model.Devices, error) {
	_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	tableName := fmt.Sprintf("%s.devices", q.keyspace)
	var devices []model.Devices
	stmt, names := qb.Select(tableName).
		Where(qb.Eq(option)).
		ToCql()
	stmt += " ALLOW FILTERING"
	query := q.session.Query(stmt, names).BindMap(qb.M{
		option: value,
	})
	if err := query.SelectRelease(&devices); err != nil {
		return nil, err
	}
	return devices, nil
}

func (q *Queries) GetDeviceInStorage(value string, option string) ([]model.Devices, error) {
	_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	tableName := fmt.Sprintf("%s.devices", q.keyspace)
	var devices []model.Devices
	stmt, names := qb.Select(tableName).
		Where(qb.Eq(option)).
		ToCql()
	stmt += " ALLOW FILTERING"
	query := q.session.Query(stmt, names).BindMap(qb.M{
		option: value,
	})
	if err := query.SelectRelease(&devices); err != nil {
		return nil, err
	}
	return devices, nil
}

func (q *Queries) GetDeviceById(id string) (model.Devices, error) {
	_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var bed model.Devices
	tableName := fmt.Sprintf("%s.devices", q.keyspace)
	stmt, names := qb.Select(tableName).
		Where(qb.Eq("id")).
		ToCql()
	query := q.session.Query(stmt, names).BindMap(qb.M{
		"id": id,
	})
	if err := query.GetRelease(&bed); err != nil {
		return model.Devices{}, err
	}
	return bed, nil
}

func (q *Queries) GetUsageDeviceByOption(value string, option string) ([]model.UsageDevice, error) {
	_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	tableName := fmt.Sprintf("%s.usage_device", q.keyspace)
	var usage_device []model.UsageDevice
	stmt, names := qb.Select(tableName).
		Where(qb.Eq(option)).
		ToCql()
	stmt += " ALLOW FILTERING"
	query := q.session.Query(stmt, names).BindMap(qb.M{
		option: value,
	})
	if err := query.SelectRelease(&usage_device); err != nil {
		return nil, err
	}
	return usage_device, nil
}

func (q *Queries) GetAllDevice() ([]model.Devices, error) {
	_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	tableName := fmt.Sprintf("%s.devices", q.keyspace)
	var devices []model.Devices
	stmt, names := qb.Select(tableName).
		ToCql()
	stmt += " ALLOW FILTERING"
	query := q.session.Query(stmt, names)
	if err := query.SelectRelease(&devices); err != nil {
		return nil, err
	}
	return devices, nil
}

func (q *Queries) UpdateDevice(oldSerial string, newSerial string) error {
	_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	tableName := fmt.Sprintf("%s.devices", q.keyspace)
	deviceNew, err := q.GetDeviceByOption(newSerial, "serial")
	if err != nil {
		return err
	}
	deviceOld, err := q.GetDeviceByOption(oldSerial, "serial")
	if err != nil {
		return err
	}
	if len(deviceNew) > 0 {
		return errors.New("duplicate serial")
	}
	if len(deviceOld) == 0 {
		return errors.New("No device data found")
	}
	update := &scylladb.UpdateDevice{
		Serial:   newSerial,
		UpdateAt: time.Now(),
		Id:       deviceOld[0].Id.String(),
	}
	stmt, names := qb.Update(tableName).
		Set("serial").
		Set("update_at").
		Where(qb.Eq("id")).
		ToCql()

	query := q.session.Query(stmt, names).BindStruct(update)
	if err := query.ExecRelease(); err != nil {
		return err
	}
	return nil
}

func (q *Queries) UpdateDeviceStt(serial string, status string) error {
	_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	tableName := fmt.Sprintf("%s.devices", q.keyspace)
	device, err := q.GetDeviceByOption(serial, "serial")
	if err != nil {
		return err
	}
	if len(device) == 0 {
		return errors.New("No device data found")
	}
	update := &scylladb.UpdateDeviceStt{
		Status:   status,
		UpdateAt: time.Now(),
		Id:       device[0].Id.String(),
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

func (q *Queries) UseDevice(req device_req.UseDeviceReq, c echo.Context) error {
	_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	tableName := fmt.Sprintf("%s.usage_device", q.keyspace)
	patient, err := q.GetPatient(req.PatientCode, "patient_code")
	if err != nil {
		return err
	}
	if len(patient) == 0 {
		return errors.New("No patient data found")
	}
	device, err := q.GetDeviceByOption(req.Serial, "serial")
	if err != nil {
		return err
	}
	if len(device) == 0 {
		return errors.New("No device data found")
	}
	if device[0].Status == "IN_USE" || device[0].Status == "DISABLED" {
		return errors.New("device is being used, can handover ")
	}
	record, err := q.GetRecordByOption(patient[0].Id.String(), "id_patient")
	if err != nil {
		return err
	}
	if len(record) == 0 {
		return errors.New("No record data found")
	}

	usageDevice, err := q.GetUsageDeviceByOption(record[0].Id.String(), "id_record")
	if err != nil {
		return err
	}
	if len(usageDevice) > 0 {
		for _, usagedevice := range usageDevice {
			if usagedevice.Status == "IN_USE" {
				return errors.New("Need remove device first")
			}
		}
	}

	id, err := gocql.ParseUUID(uuid.New().String())
	if err != nil {
		panic(err)
	}
	insert := &model.UsageDevice{
		Id:       id,
		IdDevice: device[0].Id,
		IdRecord: record[0].Id,
		Status:   "IN_USE",
		CreateAt: time.Now(),
		EndAt:    time.Time{},
	}
	stmt := qb.Insert(tableName).
		Columns("id", "id_record", "id_device", "create_at", "end_at", "status").
		Query(q.session)
	stmt.BindStruct(insert)
	if err := stmt.ExecRelease(); err != nil {
		return err
	}

	err = q.UpdateUpdater(c, record[0].Id.String())
	if err != nil {
		return err
	}

	content := "Doctor handover device"

	err = q.CreateRecordHistory(record[0].Id, content, c)
	if err != nil {
		return err
	}

	err = q.UpdateDeviceStt(req.Serial, "IN_USE")
	if err != nil {
		return err
	}
	return nil
}

func (q *Queries) UnuseDevice(serial string, c echo.Context) error {
	_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	tableName := fmt.Sprintf("%s.usage_device", q.keyspace)
	device, err := q.GetDeviceByOption(serial, "serial")
	if err != nil {
		return err
	}
	if len(device) == 0 {
		return errors.New("No device data found")
	}

	usage_device, err := q.GetUsageDeviceByOption(device[0].Id.String(), "id_device")
	if err != nil {
		return err
	}
	if len(usage_device) == 0 {
		return errors.New("No usage device data found")
	}

	record, err := q.GetRecordByOption(usage_device[0].IdRecord.String(), "id")
	if err != nil {
		return err
	}
	if len(record) == 0 {
		return errors.New("No record data found")
	}

	var updateUsageDevice model.UsageDevice
	for _, usagedevice := range usage_device {
		if usagedevice.Status == "IN_USE" {
			updateUsageDevice = usagedevice
		}
	}

	update := &scylladb.UpdateUsageTable{
		Status: "NOT_IN_USE",
		EndAt:  time.Now(),
		Id:     updateUsageDevice.Id.String(),
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

	err = q.UpdateUpdater(c, record[0].Id.String())
	if err != nil {
		return err
	}

	content := "Doctor remove device"

	err = q.CreateRecordHistory(record[0].Id, content, c)
	if err != nil {
		return err
	}

	err = q.UpdateDeviceStt(device[0].Serial, "IN_STORAGE")
	if err != nil {
		return err
	}
	return nil
}

func (q *Queries) DisableOrEnableDevice(id string, status string) error {
	_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	tableName := fmt.Sprintf("%s.devices", q.keyspace)
	device, err := q.GetDeviceById(id)
	if err != nil {
		return err
	}
	if status == "DISABLED" {
		if device.Status == "IN_USE" {
			return errors.New("Device is used, can not disable")
		}
		update := &scylladb.DisableOrEnable{
			Status:   status,
			UpdateAt: time.Now(),
			Id:       device.Id.String(),
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
	if device.Status != "DISABLED" {
		return errors.New("Device is available, can not enable")
	}
	update := &scylladb.DisableOrEnable{
		Status:   "IN_STORAGE",
		UpdateAt: time.Now(),
		Id:       device.Id.String(),
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

func (q *Queries) GetInUseDevice(c echo.Context) ([]res.DeviceInUse, error) {
	_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var result []res.DeviceInUse
	var records []model.MedicalRecords
	//doctor, err := q.GetProfileCurrent(c)
	//if err != nil {
	//	return nil, err
	//}
	//records, err := q.GetRecordByOption(doctor.Id.String(), "id_doctor")
	//if err != nil {
	//	return nil, err
	//}

	rooms, _ := q.SelectAllRoomByCurrDoctor(c)
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

	type DeviceInUseResult struct {
		deviceInUse res.DeviceInUse
		err         error
	}

	resultChan := make(chan DeviceInUseResult, len(records))
	var wg sync.WaitGroup

	for _, record := range records {
		if record.Status == "TREATING" {
			wg.Add(1)
			go func(record model.MedicalRecords) {
				defer wg.Done()
				usageDevices, err := q.GetUsageDeviceByOption(record.Id.String(), "id_record")
				if err != nil {
					resultChan <- DeviceInUseResult{err: err}
					return
				}
				for _, usageDevice := range usageDevices {
					if usageDevice.Status == "IN_USE" {
						device, err := q.GetDeviceByOption(usageDevice.IdDevice.String(), "id")
						if err != nil {
							resultChan <- DeviceInUseResult{err: err}
							return
						}
						usageBed, err := q.GetUsageBedByOption(record.Id.String(), "id_record")
						if err != nil {
							resultChan <- DeviceInUseResult{err: err}
							return
						}
						if len(usageBed) == 0 {
							resultChan <- DeviceInUseResult{err: errors.New("No usage beds data found")}
							return
						}
						for _, temp := range usageBed {
							if temp.Status == "IN_USE" {
								bed, _ := q.GetBedById(temp.IdBed.String())
								room, _ := q.GetRoomByOption(bed.IdRoom.String(), "id")
								res := res.DeviceInUse{
									Device:   device[0],
									IdRecord: record.Id.String(),
									Room:     room[0].Name,
									Bed:      bed.Name,
									InUseAt:  usageDevice.CreateAt,
								}
								resultChan <- DeviceInUseResult{deviceInUse: res}
							}
						}
					}
				}
			}(record)
		}
	}

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	for res := range resultChan {
		if res.err != nil {
			return nil, res.err
		}
		result = append(result, res.deviceInUse)
	}

	if len(result) == 0 {
		return nil, errors.New("No data found")
	}
	return result, nil
}

func (q *Queries) GetInUseDeviceAdmin(c echo.Context) ([]res.DeviceInUse, error) {
	_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var result []res.DeviceInUse
	var records []model.MedicalRecords
	//doctor, err := q.GetProfileCurrent(c)
	//if err != nil {
	//	return nil, err
	//}
	//records, err := q.GetRecordByOption(doctor.Id.String(), "id_doctor")
	//if err != nil {
	//	return nil, err
	//}

	records, _ = q.getAllRecord()

	type DeviceInUseResult struct {
		deviceInUse res.DeviceInUse
		err         error
	}

	resultChan := make(chan DeviceInUseResult, len(records))
	var wg sync.WaitGroup

	for _, record := range records {
		if record.Status == "TREATING" {
			wg.Add(1)
			go func(record model.MedicalRecords) {
				defer wg.Done()
				usageDevices, err := q.GetUsageDeviceByOption(record.Id.String(), "id_record")
				if err != nil {
					resultChan <- DeviceInUseResult{err: err}
					return
				}
				for _, usageDevice := range usageDevices {
					if usageDevice.Status == "IN_USE" {
						device, err := q.GetDeviceByOption(usageDevice.IdDevice.String(), "id")
						if err != nil {
							resultChan <- DeviceInUseResult{err: err}
							return
						}
						usageBed, err := q.GetUsageBedByOption(record.Id.String(), "id_record")
						if err != nil {
							resultChan <- DeviceInUseResult{err: err}
							return
						}
						if len(usageBed) == 0 {
							resultChan <- DeviceInUseResult{err: errors.New("No usage beds data found")}
							return
						}
						for _, temp := range usageBed {
							if temp.Status == "IN_USE" {
								bed, _ := q.GetBedById(temp.IdBed.String())
								room, _ := q.GetRoomByOption(bed.IdRoom.String(), "id")
								res := res.DeviceInUse{
									Device:   device[0],
									IdRecord: record.Id.String(),
									Room:     room[0].Name,
									Bed:      bed.Name,
									InUseAt:  usageDevice.CreateAt,
								}
								resultChan <- DeviceInUseResult{deviceInUse: res}
							}
						}
					}
				}
			}(record)
		}
	}

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	for res := range resultChan {
		if res.err != nil {
			return nil, res.err
		}
		result = append(result, res.deviceInUse)
	}

	if len(result) == 0 {
		return nil, errors.New("No data found")
	}
	return result, nil
}
