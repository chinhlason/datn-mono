package syclla

import (
	"HospitalManager/db/scylla/cql"
	"HospitalManager/db/scylla/scylladb"
	execute "HospitalManager/db/scylla/scylladb/execute"
	"context"
	"github.com/joho/godotenv"
	"github.com/scylladb/gocqlx/v2/migrate"
	"log"
)

var queries *execute.Queries

func Connect(host string, ksname string) {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()
	manager := scylladb.NewManager(host, ksname)
	err = manager.CreateKeyspace()
	if err != nil {
		panic(err)
	}
	session, err := manager.Connect()
	if err != nil {
		panic(err)
	}
	err = migrate.FromFS(ctx, session, cql.Files)
	if err != nil {
		panic(err)
	}
	queries = execute.New(session, manager.ScyllaKeyspace)
}

func Queries() *execute.Queries { return queries }
