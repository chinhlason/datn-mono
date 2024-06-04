package execute

import "github.com/scylladb/gocqlx/v2"

type Queries struct {
	session  gocqlx.Session
	keyspace string
}

func New(session gocqlx.Session, keyspace string) *Queries {
	return &Queries{session: session, keyspace: keyspace}
}
