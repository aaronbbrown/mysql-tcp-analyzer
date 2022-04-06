package main

import (
	"encoding/json"
	"time"

	"github.com/montanaflynn/stats"
)

type TimeStatistics struct {
	Min   time.Duration
	Mean  time.Duration
	P95   time.Duration
	P99   time.Duration
	Max   time.Duration
	Sum   time.Duration
	Count int
}

func NewTimeStatistics(durations []time.Duration) (TimeStatistics, error) {
	var ts TimeStatistics
	data := stats.LoadRawData(durations)

	ts.Count = len(durations)

	min, err := data.Min()
	if err != nil {
		return ts, err
	}
	ts.Min = time.Duration(min)

	mean, err := data.Mean()
	if err != nil {
		return ts, err
	}
	ts.Mean = time.Duration(mean)

	p95, err := data.Percentile(95)
	if err != nil {
		return ts, err
	}
	ts.P95 = time.Duration(p95)

	p99, err := data.Percentile(99)
	if err != nil {
		return ts, err
	}
	ts.P99 = time.Duration(p99)

	max, err := data.Max()
	if err != nil {
		return ts, err
	}
	ts.Max = time.Duration(max)

	sum, err := data.Sum()
	if err != nil {
		return ts, err
	}
	ts.Sum = time.Duration(sum)

	return ts, nil
}

func (nt *TimeStatistics) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Min   float64 `json:"min"`
		Mean  float64 `json:"mean"`
		P95   float64 `json:"p95"`
		P99   float64 `json:"p99"`
		Max   float64 `json:"max"`
		Sum   float64 `json:"sum"`
		Count int     `json:"count"`
	}{
		Min:   float64(nt.Min) / float64(time.Millisecond),
		Mean:  float64(nt.Mean) / float64(time.Millisecond),
		P95:   float64(nt.P95) / float64(time.Millisecond),
		P99:   float64(nt.P99) / float64(time.Millisecond),
		Max:   float64(nt.Max) / float64(time.Millisecond),
		Sum:   float64(nt.Sum) / float64(time.Millisecond),
		Count: nt.Count,
	})
}
