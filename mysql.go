package main

import (
	"strings"
	"time"

	"github.com/percona/go-mysql/query"
)

type MySQLQuery struct {
	rawQuery    string
	Query       string
	Tags        map[string]string
	Fingerprint string
	Duration    time.Duration
}

func NewMySQLQuery(rawquery string) MySQLQuery {
	result := MySQLQuery{
		rawQuery: rawquery,
		Query:    rawquery,
		Tags:     make(map[string]string),
	}

	n := strings.LastIndex(rawquery, " /*")
	if n >= 0 {
		result.Query = rawquery[0:n]               // strip off the comment
		comment := rawquery[n+3 : len(rawquery)-2] // strip /* and */
		split := strings.Split(comment, ",")

		for _, pair := range split {
			kv := strings.Split(pair, ":")
			if len(kv) > 1 {
				// skip these tags, they are not useful or have high cardinality
				switch kv[0] {
				case "request_id":
					continue
				case "server":
					continue
				case "application":
					continue
				case "deployed_to":
					continue
				}

				result.Tags[kv[0]] = kv[1]
			}
		}
	}

	fingerprint := query.Fingerprint(result.rawQuery)
	if fingerprint == "" {
		if strings.HasPrefix(strings.ToUpper(rawquery), "BEGIN") {
			fingerprint = "begin"
		}
		if strings.HasPrefix(strings.ToUpper(rawquery), "COMMIT") {
			fingerprint = "commit"
		}
		if strings.HasPrefix(strings.ToUpper(rawquery), "ROLLBACK") {
			fingerprint = "rollback"
		}
		fingerprint = "E_NO_FINGERPRINT"
	}
	result.Fingerprint = fingerprint

	return result
}
