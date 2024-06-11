package controller

import (
	"HospitalManager/db/scylla/scylladb/execute"
	"HospitalManager/dto/req/bed_req"
	"HospitalManager/dto/res"
	"HospitalManager/model"
	"encoding/base64"
	"errors"
	"github.com/labstack/echo/v4"
	"net/http"
	"strconv"
)

type BedController struct {
	Queries *execute.Queries
}

func (b *BedController) mapBedRes(bed model.Beds, room model.Rooms) res.BedRes {
	return res.BedRes{
		Id:       bed.Id,
		Name:     bed.Name,
		Status:   bed.Status,
		CreateAt: bed.CreateAt,
		UpdateAt: bed.UpdateAt,
		Room:     room,
	}
}

func (b *BedController) InsertBeds(c echo.Context) error {
	var reqs []bed_req.InsertBedReq
	if err := c.Bind(&reqs); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	var bugs []bed_req.InsertBedReq
	for _, req := range reqs {
		bed, err := b.Queries.GetBedByOption(req.Name, "name", req.RoomName)
		if err != nil {
			return c.JSON(http.StatusBadRequest, err.Error())
		}
		if len(bed) > 0 {
			return c.JSON(http.StatusBadRequest, res.Response{
				Message: "Cannot insert beds",
				Data:    req.Name,
			})
		}

		if b.Queries.CheckPermissionInRoomByName(req.RoomName, c) == false {
			return c.JSON(http.StatusBadRequest, res.Response{
				Message: "Do not have permission in this room",
				Data:    req.Name,
			})
		}
	}
	for _, req := range reqs {
		err := b.Queries.InsertBed(req)
		if err != nil {
			bugs = append(bugs, req)
		}
		b.Queries.UpdateNumber(1, "bed_number", req.RoomName)
	}
	if len(bugs) > 0 {
		return c.JSON(http.StatusBadRequest, res.Response{
			Message: "Cannot insert beds",
			Data:    bugs,
		})
	}
	return c.JSON(http.StatusOK, "insert success")
}

func (b *BedController) GetAllBedFromRoom(c echo.Context) error {
	room := c.QueryParam("room")
	var res []res.BedRes
	beds, err := b.Queries.SelectAllBedsFromRoom(room)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	if len(beds) == 0 {
		return c.JSON(http.StatusBadRequest, nil)
	}
	rooms, err := b.Queries.GetRoomByOption(room, "name")
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	for _, bed := range beds {
		bedResp := b.mapBedRes(bed, rooms[0])
		res = append(res, bedResp)
	}
	return c.JSON(http.StatusOK, res)
}

func (b *BedController) GetBedByStatus(c echo.Context) error {
	status := c.QueryParam("status")
	room := c.QueryParam("room")
	var res []res.BedRes
	beds, err := b.Queries.GetBedByOption(status, "status", room)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	if len(beds) == 0 {
		return c.JSON(http.StatusBadRequest, nil)
	}
	rooms, err := b.Queries.GetRoomByOption(room, "name")
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	for _, bed := range beds {
		bedResp := b.mapBedRes(bed, rooms[0])
		res = append(res, bedResp)
	}
	return c.JSON(http.StatusOK, res)
}

func (b *BedController) UpdateBed(c echo.Context) error {
	req := bed_req.UpdateBedReq{}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	err := b.Queries.UpdateBed(req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, "update success")
}

func (b *BedController) ChangeBedStt(c echo.Context) error {
	roomName := c.QueryParam("room")
	bedName := c.QueryParam("bed")
	err := b.Queries.ChangeBedStatus(bedName, roomName)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, "Change status success")
}

func (b *BedController) UsageBed(c echo.Context) error {
	req := bed_req.UsageBedReq{}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	err := b.Queries.CreateUsageBed(req, c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, "Handover bed success")
}

func (b *BedController) UnuseBed(c echo.Context) error {
	bed := c.QueryParam("bed")
	room := c.QueryParam("room")
	id := c.QueryParam("id")
	err := b.Queries.UnuseBed(id, bed, room, c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, "unuse bed success")
}

func (b *BedController) DisableBed(c echo.Context) error {
	bed := c.QueryParam("bed")
	err := b.Queries.DisableOrEnableBed(bed, "DISABLED")
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, "disable success")
}
func (b *BedController) EnableBed(c echo.Context) error {
	bed := c.QueryParam("bed")
	err := b.Queries.DisableOrEnableBed(bed, "ENABLE")
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, "enable success")
}

func (b *BedController) GetAllAvailableBed(c echo.Context) error {
	beds, err := b.Queries.GetAvailableBed(c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	if len(beds) == 0 {
		return c.JSON(http.StatusBadRequest, errors.New("No bed data found"))
	}
	var result []res.BedRes
	for _, bed := range beds {
		room, err := b.Queries.GetRoomByOption(bed.IdRoom.String(), "id")
		if err != nil {
			return c.JSON(http.StatusBadRequest, err.Error())
		}
		res := res.BedRes{
			Id:     bed.Id,
			Name:   bed.Name,
			Status: bed.Status,
			Room:   room[0],
		}
		result = append(result, res)
	}
	return c.JSON(http.StatusOK, result)
}

func (b *BedController) GetAllAvailableAndDisableBed(c echo.Context) error {
	beds, err := b.Queries.GetAvailableAndDisableBed(c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	if len(beds) == 0 {
		return c.JSON(http.StatusBadRequest, errors.New("No bed data found"))
	}
	var result []res.BedRes
	for _, bed := range beds {
		room, err := b.Queries.GetRoomByOption(bed.IdRoom.String(), "id")
		if err != nil {
			return c.JSON(http.StatusBadRequest, err.Error())
		}
		res := res.BedRes{
			Id:     bed.Id,
			Name:   bed.Name,
			Status: bed.Status,
			Room:   room[0],
		}
		result = append(result, res)
	}
	return c.JSON(http.StatusOK, result)
}

func (b *BedController) GetAllAvailableBedPagination(c echo.Context) error {
	// Get the page size from query parameters
	size := c.QueryParam("size")
	sizeInt, err := strconv.Atoi(size)
	if err != nil || sizeInt <= 0 {
		sizeInt = 10 // Default page size if not specified or invalid
	}

	// Initialize variables
	var (
		bedsRes  []model.Beds
		beds     []model.Beds
		nextPage []byte
	)

	// Fetch the available beds with pagination

	beds, nextPage, err = b.Queries.GetAvailableBedPagination(nextPage, sizeInt)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	bedsRes = append(bedsRes, beds...)

	// Encode the next page state to string for response
	nextPageStr := ""
	if len(nextPage) > 0 {
		nextPageStr = base64.StdEncoding.EncodeToString(nextPage)
	}

	// Construct the response
	response := map[string]interface{}{
		"nextPage": nextPageStr,
		"beds":     bedsRes,
	}

	return c.JSON(http.StatusOK, response)
}

func (b *BedController) GetBedRecord(c echo.Context) error {
	bed, err := b.Queries.GetBedRecord(c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, bed)
}
