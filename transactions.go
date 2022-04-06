package main

import (
	"fmt"
	"strings"
	"time"
)

type Transaction struct {
	id     int
	Frames []*Frame
}

type Transactions struct {
	// Transactions is a map of transactions, keyed by an id
	Transactions map[int]*Transaction
}

func NewTransactions() Transactions {
	return Transactions{Transactions: make(map[int]*Transaction)}
}

func (t *Transactions) Add(transaction *Transaction) {
	t.Transactions[transaction.id] = transaction
}

func (t *Transactions) Delete(id int) {
	delete(t.Transactions, id)
}

func (t *Transactions) AddFrame(id int, frame *Frame) error {
	if _, ok := t.Transactions[id]; !ok {
		return fmt.Errorf("Transaction %d not found. Try creating a new one", id)
	}

	t.Transactions[id].AddFrame(frame)

	return nil
}

func NewTransaction(id int) Transaction {
	return Transaction{id: id}
}

func (t *Transaction) AddFrame(frame *Frame) {
	t.Frames = append(t.Frames, frame)
}

// TotalDuration returns how long the transaction actually took based on the Frames
func (t *Transaction) TotalDuration() time.Duration {
	// add up the start times of all the frames, plus the duration of the commit itself
	return time.Duration(t.Frames[len(t.Frames)-1].TimeRelative-t.Frames[0].TimeRelative) + t.Frames[len(t.Frames)-1].MySQLQuery.Duration
}

// QueryDuration returns the sum of the durations of the queries in the transaction
func (t *Transaction) QueryDuration() time.Duration {
	var total time.Duration
	for _, frame := range t.Frames {
		total = time.Duration(total + frame.MySQLQuery.Duration)
	}
	return total
}

// WasteDuration returns how much time was wasted within the transaction
// it is the delta between total duration and query duration
func (t *Transaction) WasteDuration() time.Duration {
	return time.Duration(t.TotalDuration() - t.QueryDuration())
}

func (t *Transaction) WastePercentage() int {
	return 100 - int(float64(t.QueryDuration())/float64(t.TotalDuration())*100)
}

// Queries returns the query for the transaction
func (t *Transaction) Queries() []string {
	var result []string
	for _, frame := range t.Frames {
		result = append(result, frame.MySQLQuery.Query)
	}
	return result
}

// Fingerprint attempts to normalize the transaction by removing duplicate query
func (t *Transaction) Fingerprint() string {
	var builder strings.Builder
	// use a map to deduplicate
	fingerprints := make(map[string]bool)
	for _, frame := range t.Frames {
		if _, ok := fingerprints[frame.MySQLQuery.Fingerprint]; !ok {
			builder.WriteString(frame.MySQLQuery.Fingerprint)
			builder.WriteString("\n")
			fingerprints[frame.MySQLQuery.Fingerprint] = true
		}
	}

	return builder.String()
}

func (t *Transaction) FingerprintSlice(deduplicate bool) []string {
	var result []string
	// use a map to deduplicate
	fingerprints := make(map[string]bool)

	for _, frame := range t.Frames {
		if !deduplicate {
			result = append(result, frame.MySQLQuery.Fingerprint)
			continue
		}

		if _, ok := fingerprints[frame.MySQLQuery.Fingerprint]; !ok {
			result = append(result, frame.MySQLQuery.Fingerprint)
		}
		fingerprints[frame.MySQLQuery.Fingerprint] = true
	}

	return result
}
