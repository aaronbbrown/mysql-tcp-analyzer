// this expects a file in the format of the output of the following command piped into stdin:
// tshark -r mysql.pcap -Y mysql -Tjson -e tcp.analysis.lost_segment -e tcp.analysis.ack_lost_segment -e frame.number -e frame.time_relative -e tcp.stream -e mysql.command -e mysql.query -e mysql.payload -e mysql.response_code
package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/davecgh/go-spew/spew"
)

type rawsource struct {
	Layers Layers `json:"layers"`
}

type rawframe struct {
	Source rawsource `json:"_source"`
}

func main() {
	var rawframes []rawframe
	r := bufio.NewReader(os.Stdin)
	dec := json.NewDecoder(r)

	for {
		if err := dec.Decode(&rawframes); err == io.EOF {
			break
		} else if err != nil {
			log.Fatal(err)
		}
	}

	fp := NewFrameParser()

	if err := fp.ParseRawFrames(rawframes); err != nil {
		log.Fatal(err)
	}

	mode := flag.String("mode", "debug", "mode (debug, count-tags, queries-for-tag, tags-for-fingerprint)")

	// for queries-for-tag
	key := flag.String("key", "", "key")
	value := flag.String("value", "", "value")

	// for query-for-fingerprint
	fingerprint := flag.String("fingerprint", "", "fingerprint")

	flag.Parse()

	switch *mode {
	case "debug":
		spew.Dump(fp.Transactions)
		spew.Dump(fp.Frames)

	case "count-tags":
		tags := fp.Frames.CountByTag()
		for k, v := range tags {
			fmt.Println(k, v)
		}

	case "queries-for-tag":
		queries := fp.Frames.QueriesForTag(*key, *value)
		for q, count := range queries {
			fmt.Println(count, "\t", q)
		}

	case "tags-for-fingerprint":
		fmt.Println("Fingerprint: ", *fingerprint)
		tags := fp.Frames.TagsForFingerprint(*fingerprint)
		for tag, count := range tags {
			fmt.Println(count, "\t", tag)
		}

	case "transactions":
		for _, t := range fp.Transactions.Transactions {
			fmt.Println("---")
			fmt.Println("Total Duration: ", t.TotalDuration())
			fmt.Println("Query Duration: ", t.QueryDuration())
			fmt.Println("Waste Duration: ", t.WasteDuration())
			fmt.Println("Waste Percentage: ", t.WastePercentage())
			fmt.Println("Transaction Fingerprint:")
			fmt.Println(t.Fingerprint())
			fmt.Println()
			fmt.Println("Example (Fingerprinted):")
			for _, f := range t.Frames {
				if f.MySQLQuery.Fingerprint == "" {
					fmt.Println(f.MySQLQuery.Query)
					continue
				}

				fmt.Println(f.MySQLQuery.Fingerprint)
			}
		}

	case "normalized-transactions":
		nts := NewNormalizedTransactions()
		for _, t := range fp.Transactions.Transactions {
			nts.Add(*t)
		}

		b, err := json.MarshalIndent(&nts, "", " ")
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(string(b))
	}
}
