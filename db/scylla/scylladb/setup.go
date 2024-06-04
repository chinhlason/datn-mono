package scylladb

import (
	"fmt"
	"github.com/gocql/gocql"
	"github.com/scylladb/gocqlx/v2"
)

type Manager struct {
	ScyllaHost     string
	ScyllaKeyspace string
}

func NewManager(host string, ksname string) *Manager {
	return &Manager{
		ScyllaHost:     host,
		ScyllaKeyspace: ksname,
	}
}

func (m *Manager) Connect() (gocqlx.Session, error) {
	return m.connect(m.ScyllaKeyspace, m.ScyllaHost)
}

func (m *Manager) CreateKeyspace() error {
	session, err := m.connect("system", m.ScyllaHost)
	if err != nil {
		return err
	}
	defer session.Close()
	stmt := fmt.Sprintf(`CREATE KEYSPACE IF NOT EXISTS %s WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1}`, m.ScyllaKeyspace)

	return session.ExecStmt(stmt)
}

func (m *Manager) connect(keyspace string, host string) (gocqlx.Session, error) {
	c := gocql.NewCluster(host)
	c.Keyspace = keyspace
	return gocqlx.WrapSession(c.CreateSession())
}
