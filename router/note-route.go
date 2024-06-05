package router

import "C"
import (
	controller2 "HospitalManager/controller"
	"HospitalManager/db/scylla/scylladb/execute"
	"HospitalManager/middleware"
	"github.com/labstack/echo/v4"
)

func NoteRoute(e *echo.Echo, q *execute.Queries) {
	controller := controller2.NoteController{
		Queries: q,
	}

	c := e.Group("/note")
	//c.Use(middleware.SetJWTHeader)
	c.Use(middleware.JWTMiddleware())
	c.Use(middleware.ValidateAndExtractClaims)

	c.POST("/create", controller.CreateNote)
	c.DELETE("/delete", controller.DeleteNote)
	c.PUT("/update", controller.UpdateNote)
	c.GET("/get", controller.GetAll)
}
