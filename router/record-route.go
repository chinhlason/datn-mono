package router

import (
	controller2 "HospitalManager/controller"
	"HospitalManager/db/scylla/scylladb/execute"
	"HospitalManager/middleware"
	"github.com/labstack/echo/v4"
)

func RecordRoute(e *echo.Echo, q *execute.Queries) {
	controller := controller2.RecordController{
		Queries: q,
	}

	c := e.Group("/record")
	//c.Use(middleware.SetJWTHeader)
	c.Use(middleware.JWTMiddleware())
	c.Use(middleware.ValidateAndExtractClaims)

	c.POST("/create", controller.CreateRecord)
	c.GET("/get", controller.GetRecordById)
	c.GET("/get-all-total", controller.GetAllTotalRecord)
	c.GET("/get-all-total-admin", controller.GetTotalByAdmin, middleware.IsADMIN)
	c.GET("/get-all-pending", controller.GetPendingRecord)
	c.GET("/search", controller.SearchTotalRecord)
	c.GET("/search-admin", controller.SearchTotalByAdmin, middleware.IsADMIN)
	c.GET("/search-pending", controller.SearchPendingRecord)
	c.GET("/statistical", controller.Statistical)
	c.PUT("/leave", controller.LeaveHospital)
}
