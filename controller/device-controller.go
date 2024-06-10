package controller

import (
	"HospitalManager/db/scylla/scylladb/execute"
	"HospitalManager/dto/req/device_req"
	"HospitalManager/dto/res"
	"HospitalManager/model"
	mqtt2 "HospitalManager/mqtt"
	"errors"
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/labstack/echo/v4"
	"net/http"
	"strconv"
)

type DeviceController struct {
	Queries *execute.Queries
	Mqtt    mqtt.Client
}

func (d *DeviceController) AddDevices(c echo.Context) error {
	var reqs []device_req.AddDeviceReq
	if err := c.Bind(&reqs); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	var bugs []device_req.AddDeviceReq
	for _, req := range reqs {
		device, err := d.Queries.GetDeviceByOption(req.Serial, "serial")
		if err != nil {
			return c.JSON(http.StatusBadRequest, err.Error())
		}
		if len(device) > 0 {
			return c.JSON(http.StatusBadRequest, res.Response{
				Message: "Cannot add devices",
				Data:    device,
			})
		}
	}
	for _, req := range reqs {
		err := d.Queries.AddDevice(req)
		if err != nil {
			bugs = append(bugs, req)
		}
	}
	if len(bugs) > 0 {
		return c.JSON(http.StatusBadRequest, res.Response{
			Message: "Cannot add devices",
			Data:    bugs,
		})
	}
	return c.JSON(http.StatusOK, "Add device success")
}

func (d *DeviceController) UpdateDevice(c echo.Context) error {
	oldSerial := c.QueryParam("oldserial")
	newSerial := c.QueryParam("newserial")
	err := d.Queries.UpdateDevice(oldSerial, newSerial)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, "update success")
}

func (d *DeviceController) GetDeviceByOption(c echo.Context) error {
	value := c.QueryParam("value")
	option := c.QueryParam("option")
	devices, err := d.Queries.GetDeviceByOption(value, option)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, devices)
}
func (d *DeviceController) GetDeviceInStorage(c echo.Context) error {
	devices, err := d.Queries.GetAllDevice()
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	var result []model.Devices
	for _, device := range devices {
		if device.Status == "IN_STORAGE" || device.Status == "DISABLED" {
			result = append(result, device)
		}
	}
	return c.JSON(http.StatusOK, result)
}
func (d *DeviceController) GetAllDevices(c echo.Context) error {
	devices, err := d.Queries.GetAllDevice()
	var res res.DeviceRes
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	res.DeviceNumber = len(devices)
	for _, device := range devices {
		res.Devices = append(res.Devices, device)
	}
	return c.JSON(http.StatusOK, res)
}

func (d *DeviceController) UseDevice(c echo.Context) error {
	var req device_req.UseDeviceReq
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	err := d.Queries.UseDevice(req, c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, "handover device success")
}

func (d *DeviceController) UnuseDevice(c echo.Context) error {
	serial := c.QueryParam("serial")
	err := d.Queries.UnuseDevice(serial, c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, "unuse device success")
}

func (d *DeviceController) DisableDevice(c echo.Context) error {
	bed := c.QueryParam("device")
	err := d.Queries.DisableOrEnableDevice(bed, "DISABLED")
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, "disable device success")
}
func (d *DeviceController) EnableDevice(c echo.Context) error {
	bed := c.QueryParam("device")
	err := d.Queries.DisableOrEnableDevice(bed, "ENABLE")
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, "enable device success")
}

func (d *DeviceController) GetInUse(c echo.Context) error {
	data, err := d.Queries.GetInUseDevice(c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, data)
}

func (d *DeviceController) GetInUseByAdmin(c echo.Context) error {
	data, err := d.Queries.GetInUseDeviceAdmin(c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, data)
}

type Shutdown struct {
	NewStatus int8 `json:"new_status"`
}

type Message struct {
	Id       string   `json:"id"`
	Shutdown Shutdown `json:"shutdown"`
}

func (d *DeviceController) OnOffDevice(c echo.Context) error {
	control := c.QueryParam("control")
	var controlInt int8
	if control == "on" {
		controlInt = 0
	} else {
		controlInt = 1
	}
	fmt.Println(control == "on")
	device := c.QueryParam("device")
	topic := fmt.Sprintf("ibme/device/shutdown/update/%s", device)
	msg := Message{
		Id: device,
		Shutdown: Shutdown{
			NewStatus: controlInt,
		},
	}
	mqtt2.Publish(d.Mqtt, msg, topic)
	return c.JSON(http.StatusOK, msg)
}
func (d *DeviceController) CheckOnline(c echo.Context) error {
	topicSub := "ibme/server/online/list"
	topicPub := "ibme/web/online/list"

	type Result struct {
		devices []model.Devices
		err     error
	}

	rooms, _ := d.Queries.SelectAllRoomByCurrDoctor(c)

	// Kênh để nhận kết quả từ các goroutine
	resultChan := make(chan Result)

	for _, room := range rooms {
		go func(room model.Rooms) {
			var result []model.Devices
			record, _ := d.Queries.GetRecordByOption(room.Id.String(), "id_room")
			if len(record) > 0 {
				var records []model.MedicalRecords
				for _, temp := range record {
					records = append(records, temp)
				}

				listOnline, err := mqtt2.CheckOnline(d.Mqtt, topicSub, topicPub)
				if err != nil {
					resultChan <- Result{nil, err}
					return
				}

				if len(listOnline.Device) == 0 {
					resultChan <- Result{nil, nil}
					return
				}

				for _, temp := range listOnline.Device {
					device, err := d.Queries.GetDeviceByOption(temp, "serial")
					if err != nil {
						resultChan <- Result{nil, err}
						return
					}
					if len(device) > 0 {
						for _, record := range records {
							usageDevice, _ := d.Queries.GetUsageDeviceByOption(record.Id.String(), "id_record")
							for _, usaDevice := range usageDevice {
								if usaDevice.IdDevice == device[0].Id && usaDevice.Status == "IN_USE" {
									result = append(result, device[0])
								}
							}
						}
					}
				}
				resultChan <- Result{result, nil}
			} else {
				resultChan <- Result{nil, nil}
			}
		}(room)
	}

	var finalResult []model.Devices
	for range rooms {
		res := <-resultChan
		if res.err != nil {
			return c.JSON(http.StatusBadRequest, res.err.Error())
		}
		if res.devices != nil {
			finalResult = append(finalResult, res.devices...)
		}
	}

	if len(finalResult) == 0 {
		return c.JSON(http.StatusBadRequest, errors.New("no device data found"))
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": strconv.Itoa(len(finalResult)),
		"data":    finalResult,
	})
}

func (d *DeviceController) CheckOnlineAdmin(c echo.Context) error {
	topicSub := "ibme/server/online/list"
	topicPub := "ibme/web/online/list"
	var result []model.Devices
	listOnline, err := mqtt2.CheckOnline(d.Mqtt, topicSub, topicPub)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	if len(listOnline.Device) == 0 {
		return c.JSON(http.StatusOK, res.Response{Message: "no online data", Data: nil})
	}
	for _, temp := range listOnline.Device {
		device, err := d.Queries.GetDeviceByOption(temp, "serial")
		if err != nil {
			return c.JSON(http.StatusBadRequest, err.Error())
		}
		if len(device) > 0 {
			result = append(result, device[0])
		}
	}
	if len(result) == 0 {
		return c.JSON(http.StatusBadRequest, errors.New("no device data found"))
	}
	return c.JSON(http.StatusOK, res.Response{
		Message: strconv.Itoa(listOnline.Number),
		Data:    result,
	})
}
