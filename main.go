package main

import (
	syclla "HospitalManager/db/scylla"
	"HospitalManager/mqtt"
	"HospitalManager/router"
	"flag"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	scyllaHost := flag.String("scyllaHost", "scylladb:9042", "Scylla listen address")
	scyllaKS := flag.String("scyllaDB", "scylladb", "Scylla keyspace name")

	e := echo.New()

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders: []string{"Origin", "Content-Type", "Accept", "Authorization"},
		//AllowCredentials: true,
	}))

	syclla.Connect(*scyllaHost, *scyllaKS)
	queries := syclla.Queries()
	client := mqtt.Connect()
	router.UserRoute(e, queries)
	router.RoomRoute(e, queries)
	router.BedRoute(e, queries)
	router.RecordRoute(e, queries)
	router.DeviceRoute(e, queries, client)
	router.NoteRoute(e, queries)
	e.Logger.Fatal(e.Start(":8081"))
}
