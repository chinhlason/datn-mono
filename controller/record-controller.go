package controller

import (
	"HospitalManager/db/scylla/scylladb/execute"
	"HospitalManager/dto/req/record_req"
	"github.com/labstack/echo/v4"
	"net/http"
)

type RecordController struct {
	Queries *execute.Queries
}

func (rc *RecordController) CreateRecord(c echo.Context) error {
	var req record_req.InsertRecordReq
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	err := rc.Queries.InsertRecord(req, c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, "Create success")
}

func (rc *RecordController) GetRecordById(c echo.Context) error {
	id := c.QueryParam("id")
	record, err := rc.Queries.GetRecordWithGoRoutine(id)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, record)
}

func (rc *RecordController) GetPendingRecord(c echo.Context) error {
	record, err := rc.Queries.GetAllPendingRecord(c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, record)
}

func (rc *RecordController) LeaveHospital(c echo.Context) error {
	id := c.QueryParam("id")
	usageBeds, err := rc.Queries.GetUsageBedByOption(id, "id_record")
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	if len(usageBeds) > 0 {
		for _, usageBed := range usageBeds {
			if usageBed.Status == "IN_USE" {
				return c.JSON(http.StatusBadRequest, "Need to remove bed first")
			}
		}
	}

	usageDevices, err := rc.Queries.GetUsageDeviceByOption(id, "id_record")
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	if len(usageDevices) > 0 {
		for _, usageDevice := range usageDevices {
			if usageDevice.Status == "IN_USE" {
				return c.JSON(http.StatusBadRequest, "Need to remove device first")
			}
		}
	}

	err = rc.Queries.ChangeRecordStatus(id, "LEAVED")
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	err = rc.Queries.CreateRecordHistoryStr(id, "patient leave hospital", c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusOK, "Discharge from hospital success")
}

func (rc *RecordController) GetAllTotalRecord(c echo.Context) error {
	records, err := rc.Queries.GetAllPatientByDoctor(c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, records)
}
func (rc *RecordController) SearchTotalRecord(c echo.Context) error {
	search := c.QueryParam("q")
	records, err := rc.Queries.SearchTotalRecord(search, c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, records)
}
