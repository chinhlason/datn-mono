package router

import (
	controller2 "HospitalManager/controller"
	"HospitalManager/db/scylla/scylladb/execute"
	"HospitalManager/middleware"
	"github.com/labstack/echo/v4"
)

func RoomRoute(e *echo.Echo, q *execute.Queries) {
	controller := controller2.RoomController{
		Queries: q,
	}

	c := e.Group("/room")
	c.Use(middleware.SetJWTHeader)
	c.Use(middleware.JWTMiddleware())
	c.Use(middleware.ValidateAndExtractClaims)

	c.POST("/create", controller.CreateRoom)
	c.GET("/get-all", controller.GetAllByCurrDoctor)
	c.GET("/get-all-by-admin", controller.GetAllByAdmin, middleware.IsADMIN)
	c.GET("/get", controller.GetByOption, middleware.IsADMIN)
	c.GET("/get-by-name", controller.GetRoomByName)
	c.GET("/get-detail", controller.GetAllShortRecord)
	c.GET("/get-detail-pagi", controller.GetAllShortRecordPagi)
	c.PUT("/update", controller.UpdateRoom)
	c.PUT("/handover", controller.HandOver, middleware.IsADMIN)
}
