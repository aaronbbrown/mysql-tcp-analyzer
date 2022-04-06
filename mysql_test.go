package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewMySQLQuery(t *testing.T) {
	tests := []struct {
		query       string
		fingerprint string
	}{
		{
			query:       "SELECT * FROM foo WHERE bar = 1",
			fingerprint: "select * from foo where bar = ?",
		},
		{
			query:       "BEGIN",
			fingerprint: "begin",
		},
	}

	for _, test := range tests {
		q := NewMySQLQuery(test.query)
		assert.Equal(t, test.fingerprint, q.Fingerprint)
	}
}
