package router

import (
	controller2 "HospitalManager/controller"
	"HospitalManager/db/scylla/scylladb/execute"
	"HospitalManager/middleware"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/labstack/echo/v4"
)

func DeviceRoute(e *echo.Echo, q *execute.Queries, client mqtt.Client) {
	controller := controller2.DeviceController{
		Queries: q,
		Mqtt:    client,
	}

	c := e.Group("/device")
	c.Use(middleware.SetJWTHeader)
	c.Use(middleware.JWTMiddleware())
	c.Use(middleware.ValidateAndExtractClaims)

	c.POST("/add", controller.AddDevices)
	c.PUT("/update", controller.UpdateDevice)
	c.GET("/get-by-option", controller.GetDeviceByOption)
	c.GET("/in-storage", controller.GetDeviceInStorage)
	c.GET("/get-in-use", controller.GetInUse)
	c.GET("/get-all", controller.GetAllDevices)
	c.POST("/use-device", controller.UseDevice)
	c.PUT("/unuse-device", controller.UnuseDevice)
	c.PUT("/disable", controller.DisableDevice)
	c.PUT("/enable", controller.EnableDevice)
	c.POST("/on-off", controller.OnOffDevice)
}
