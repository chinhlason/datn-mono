package router

import (
	controller2 "HospitalManager/controller"
	"HospitalManager/db/scylla/scylladb/execute"
	"HospitalManager/middleware"
	"github.com/labstack/echo/v4"
)

func BedRoute(e *echo.Echo, q *execute.Queries) {
	controller := controller2.BedController{
		Queries: q,
	}
	c := e.Group("/bed")
	c.Use(middleware.SetJWTHeader)
	c.Use(middleware.JWTMiddleware())
	c.Use(middleware.ValidateAndExtractClaims)

	c.POST("/insert", controller.InsertBeds)
	c.GET("/get", controller.GetAllBedFromRoom)
	c.GET("/get-by-status", controller.GetBedByStatus)
	c.GET("/get-available", controller.GetAllAvailableBed)
	c.GET("/storage", controller.GetAllAvailableAndDisableBed)
	c.GET("/get-available-page", controller.GetAllAvailableBedPagination)
	c.GET("/get-used-bed", controller.GetBedRecord)
	c.PUT("/update", controller.UpdateBed)
	c.PUT("/change-stt", controller.ChangeBedStt)
	c.POST("/usage-bed", controller.UsageBed)
	c.PUT("/unusage-bed", controller.UnuseBed)
	c.PUT("/disable", controller.DisableBed)
	c.PUT("/enable", controller.EnableBed)
}
