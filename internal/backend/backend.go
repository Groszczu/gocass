package backend

import (
	"github.com/gocql/gocql"
	"github.com/scylladb/gocqlx/v2"
)

const Keyspace = "shop"

func Session(hosts ...string) (gocqlx.Session, error) {
	cluster := gocql.NewCluster(hosts...)
	cluster.Keyspace = Keyspace

	return  gocqlx.WrapSession(cluster.CreateSession())
}
