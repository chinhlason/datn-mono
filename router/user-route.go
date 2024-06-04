package router

import (
	controller2 "HospitalManager/controller"
	"HospitalManager/db/scylla/scylladb/execute"
	"HospitalManager/middleware"
	"github.com/labstack/echo/v4"
)

func UserRoute(e *echo.Echo, q *execute.Queries) {
	controller := controller2.UserController{
		Queries: q,
	}

	err := controller.CreateAdminAccount()
	if err != nil {
		panic(err)
	}

	c := e.Group("/user")
	c.Use(middleware.SetJWTHeader)
	c.Use(middleware.JWTMiddleware())
	c.Use(middleware.ValidateAndExtractClaims)

	e.POST("/register", controller.Register)
	e.POST("/register-list", controller.RegisterList)
	e.POST("/login", controller.Login)
	e.POST("/log-out", controller.Logout)
	e.GET("/test", controller.Test)
	c.POST("/refresh-token", controller.RefreshToken)
	c.GET("/get-all-doctors", controller.GetAllUsers, middleware.IsADMIN)
	c.PUT("/update", controller.UpdateProfile)
	c.PUT("/change-password", controller.ChangePassword)
	c.PUT("/change-permission", controller.ChangePermission, middleware.IsADMIN)
	c.GET("/profile", controller.GetProfileCurrent)
}
