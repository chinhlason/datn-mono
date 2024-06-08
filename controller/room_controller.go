package controller

import (
	"HospitalManager/db/scylla/scylladb/execute"
	"HospitalManager/dto/req/room_req"
	"HospitalManager/dto/res"
	"HospitalManager/model"
	"github.com/labstack/echo/v4"
	"net/http"
	"strconv"
)

type RoomController struct {
	Queries *execute.Queries
}

func (r *RoomController) CreateRoom(c echo.Context) error {
	var reqs []room_req.CreateRoomReq
	if err := c.Bind(&reqs); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	var bugs []room_req.CreateRoomReq
	for _, req := range reqs {
		err := r.Queries.InsertRoom(req)
		if err != nil {
			bugs = append(bugs, req)
		}
	}
	if len(bugs) > 0 {
		return c.JSON(http.StatusBadRequest, res.Response{
			Message: "Cannot insert room",
			Data:    bugs,
		})
	}
	return c.JSON(http.StatusOK, "Create Success")
}

func (r *RoomController) mapToRoomRes(room model.Rooms, userId string) res.Room {
	user, _ := r.Queries.GetUserByOption(userId, "id")
	return res.Room{
		Id:            room.Id,
		Name:          room.Name,
		BedNumber:     room.BedNumber,
		PatientNumber: room.PatientNumber,
		CreateAt:      room.CreateAt,
		UpdateAt:      room.UpdateAt,
		Doctor:        user[0],
	}
}

func (r *RoomController) GetAllByCurrDoctor(c echo.Context) error {
	var res []res.Room
	rooms, err := r.Queries.SelectAllRoomByCurrDoctor(c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	if len(rooms) == 0 {
		return c.JSON(http.StatusOK, nil)
	}
	for _, room := range rooms {
		user, _ := r.Queries.GetUserByOption(room.IdDoctor.String(), "id")
		roomRes := r.mapToRoomRes(room, user[0].Id.String())
		res = append(res, roomRes)
	}
	return c.JSON(http.StatusOK, res)
}

func (r *RoomController) GetAllByAdmin(c echo.Context) error {
	rooms, err := r.Queries.GetAllRooms()
	var res []res.Room
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	if len(rooms) == 0 {
		return c.JSON(http.StatusOK, nil)
	}
	for _, room := range rooms {
		roomRes := r.mapToRoomRes(room, room.IdDoctor.String())
		res = append(res, roomRes)
	}
	return c.JSON(http.StatusOK, res)
}

func (r *RoomController) GetByOption(c echo.Context) error {
	value := c.QueryParam("value")
	option := c.QueryParam("option")
	var res []res.Room
	rooms, err := r.Queries.GetRoomByOption(value, option)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	if len(rooms) == 0 {
		return c.JSON(http.StatusBadRequest, nil)
	}
	for _, room := range rooms {
		roomRes := r.mapToRoomRes(room, room.IdDoctor.String())
		res = append(res, roomRes)
	}
	return c.JSON(http.StatusOK, res)
}

func (r *RoomController) GetRoomByName(c echo.Context) error {
	name := c.QueryParam("name")
	var res []res.Room
	rooms, err := r.Queries.GetRoomByOption(name, "name")
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	if len(rooms) == 0 {
		return c.JSON(http.StatusBadRequest, nil)
	}
	for _, room := range rooms {
		roomRes := r.mapToRoomRes(room, room.IdDoctor.String())
		res = append(res, roomRes)
	}
	return c.JSON(http.StatusOK, res)
}

func (r *RoomController) UpdateRoom(c echo.Context) error {
	var req room_req.UpdateRoomReq
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	err := r.Queries.UpdateRoom(req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, "update success")
}

func (r *RoomController) HandOver(c echo.Context) error {
	roomName := c.QueryParam("room")
	doctorCode := c.QueryParam("doctorcode")
	err := r.Queries.HandoverRoom(roomName, doctorCode)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, "Handover success")
}

func (r *RoomController) GetAllShortRecord(c echo.Context) error {
	roomName := c.QueryParam("room")
	records, err := r.Queries.GetAllRecordInRoom(roomName)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, records)
}
func (r *RoomController) GetAllShortRecordPagi(c echo.Context) error {
	page := c.QueryParam("page")
	size := c.QueryParam("size")
	pageInt, _ := strconv.Atoi(page)
	sizeInt, _ := strconv.Atoi(size)
	roomName := c.QueryParam("room")
	records, err, maxPage := r.Queries.GetAllRecordInRoomPagination(roomName, pageInt, sizeInt)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	maxPageStr := strconv.Itoa(maxPage)
	return c.JSON(http.StatusOK, res.Response{
		Message: maxPageStr,
		Data:    records,
	})
}

func (r *RoomController) HandoverRoom(c echo.Context) error {
	idRoom := c.QueryParam("room")
	idDoctor := c.QueryParam("doctor")
	err := r.Queries.HandoverRoomForNormalDoctor(idRoom, idDoctor)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}
	return c.JSON(http.StatusOK, "handover room successfully!")
}

func (r *RoomController) RoomDetailByAdmin(c echo.Context) error {
	room := c.QueryParam("room")
	detail, err := r.Queries.GetRoomDetail(room)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, detail)
}

func (r *RoomController) DeleteMember(c echo.Context) error {
	idRoom := c.QueryParam("room")
	doctor := c.QueryParam("doctor")
	err := r.Queries.DeleteDoctorFromRoom(idRoom, doctor)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, "delete successfully")
}
