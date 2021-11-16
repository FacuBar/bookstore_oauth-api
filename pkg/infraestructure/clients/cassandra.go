package clients

import (
	"os"

	"github.com/gocql/gocql"
)

func ConnectDb() *gocql.Session {
	cluster := gocql.NewCluster(os.Getenv("CASSANDRA_ADDRESS"))
	cluster.Keyspace = os.Getenv("CASSANDRA_KEYSPACE")
	cluster.Consistency = gocql.Quorum

	session, err := cluster.CreateSession()
	if err != nil {
		panic(err)
	}

	return session
}
