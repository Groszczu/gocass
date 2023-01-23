package models

import (
	"github.com/gocql/gocql"
)

type UUID = gocql.UUID

func EmptyUUID() UUID {
	return UUID{}
}

func RandomUUID() UUID {
	uuid, err := gocql.RandomUUID()
	if err != nil {
		panic("Unable to generate UUID")
	}
	return uuid
}
