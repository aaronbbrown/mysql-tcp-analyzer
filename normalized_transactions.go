package main

import (
	"encoding/json"
	"sort"
	"time"
)

type NormalizedTransactions struct {
	// Transactions is a map of normalized transactions, keyed by transaction fingerprint
	Transactions map[string]*NormalizedTransaction `json:"transactions"`
}

type NormalizedTransaction struct {
	Fingerprint          []string
	Example              []string
	tags                 map[string]bool
	queryDurations       []time.Duration
	transactionDurations []time.Duration
	wasteDurations       []time.Duration
}

func NewNormalizedTransactions() NormalizedTransactions {
	return NormalizedTransactions{
		Transactions: make(map[string]*NormalizedTransaction),
	}
}

func (nts *NormalizedTransactions) Add(transaction Transaction) {
	nt := NewNormalizedTransaction()
	fingerprint := transaction.Fingerprint()
	if _, ok := nts.Transactions[fingerprint]; !ok {
		nt.Fingerprint = transaction.FingerprintSlice(true)
		nt.Example = transaction.FingerprintSlice(false)
		nts.Transactions[fingerprint] = &nt
	}

	for _, frame := range transaction.Frames {
		for k, v := range frame.MySQLQuery.Tags {
			nts.Transactions[fingerprint].tags[k+":"+v] = true
		}
	}

	nts.Transactions[fingerprint].queryDurations = append(nts.Transactions[fingerprint].queryDurations, transaction.QueryDuration())
	nts.Transactions[fingerprint].transactionDurations = append(nts.Transactions[fingerprint].transactionDurations, transaction.TotalDuration())
	nts.Transactions[fingerprint].wasteDurations = append(nts.Transactions[fingerprint].wasteDurations, transaction.WasteDuration())
}

func (nts *NormalizedTransactions) MarshalJSON() ([]byte, error) {
	data := make([]NormalizedTransaction, 0, len(nts.Transactions))
	for _, transaction := range nts.Transactions {
		data = append(data, *transaction)
	}
	return json.Marshal(data)
}

func NewNormalizedTransaction() NormalizedTransaction {
	return NormalizedTransaction{
		tags: make(map[string]bool),
	}
}

func (nt *NormalizedTransaction) MarshalJSON() ([]byte, error) {
	var queryStatistics TimeStatistics
	var transactionStatistics TimeStatistics
	var wasteStatistics TimeStatistics

	data := struct {
		Fingerprint           []string        `json:"fingerprint"`
		Example               []string        `json:"example_query"`
		WastePercentage       float64         `json:"waste_percentage"`
		Tags                  []string        `json:"tags"`
		QueryStatistics       *TimeStatistics `json:"query_statistics"`
		TransactionStatistics *TimeStatistics `json:"transaction_statistics"`
		WasteStatistics       *TimeStatistics `json:"waste_statistics"`
	}{
		Fingerprint: nt.Fingerprint,
		Example:     nt.Example,
	}

	queryStatistics, err := NewTimeStatistics(nt.queryDurations)
	if err != nil {
		return nil, err
	}
	data.QueryStatistics = &queryStatistics

	transactionStatistics, err = NewTimeStatistics(nt.transactionDurations)
	if err != nil {
		return nil, err
	}
	data.TransactionStatistics = &transactionStatistics

	wasteStatistics, err = NewTimeStatistics(nt.wasteDurations)
	if err != nil {
		return nil, err
	}
	data.WasteStatistics = &wasteStatistics

	data.WastePercentage = float64(wasteStatistics.Mean) / float64(transactionStatistics.Mean) * 100

	for tag := range nt.tags {
		data.Tags = append(data.Tags, tag)
	}
	sort.Strings(data.Tags)

	return json.Marshal(data)
}
