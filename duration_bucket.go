package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type DurationBuckets struct {
	// interval is the bucket size
	interval time.Duration
	buckets  []DurationBucket
	// the most recent bucket index
	// buckets must be added in order
	index         int
	closedStreams map[int]bool
	seenStreams   map[int]bool
}

func NewDurationBuckets(interval time.Duration) DurationBuckets {
	return DurationBuckets{
		interval:      interval,
		index:         0,
		closedStreams: make(map[int]bool),
		seenStreams:   make(map[int]bool),
	}
}

func (db *DurationBuckets) AddFrame(frame Frame) error {
	idx := db.bucket(frame.TimeRelative)
	if idx < db.index {
		return errors.New("frame is in the past")
	}

	if len(db.buckets) == 0 {
		db.buckets = append(db.buckets, NewDurationBucket(DurationBucket{}, db.closedStreams))
	}

	// expand the bucket and move the bucket pointer
	for i := db.index; i < idx; i++ {
		//		fmt.Printf("Adding bucket i: %d db.index: %d, idx: %d, len(db.buckets): %d, frame.TimeRelative: %s\n", i, db.index, idx, len(db.buckets), frame.TimeRelative)
		db.buckets = append(db.buckets, NewDurationBucket(db.buckets[i], db.closedStreams))
	}
	db.index = idx
	db.buckets[idx].AddFrame(frame, db.seenStreams, db.closedStreams)

	return nil
}

func (db *DurationBuckets) TSV() string {
	var builder strings.Builder

	builder.WriteString("Time,Concurrent,New,Closed\n")

	for idx, bucket := range db.buckets {
		builder.WriteString((time.Duration(idx) * db.interval).String())
		builder.WriteString("\t")
		builder.WriteString(strconv.Itoa(bucket.CountConcurrent()))
		builder.WriteString("\t")
		builder.WriteString(strconv.Itoa(bucket.CountNew()))
		builder.WriteString("\t")
		builder.WriteString(strconv.Itoa(bucket.CountClosed()))
		builder.WriteString("\n")
	}

	return builder.String()
}

// bucket returns the bucket index for the given duration
func (db *DurationBuckets) bucket(d time.Duration) int {
	return int(d / db.interval)
}

type DurationBucket struct {
	// streams is a map used to count the number of streams that fall into this bucket
	// key is tcp stream
	streams map[int]bool

	newStreams    map[int]bool
	closedStreams map[int]bool
}

func NewDurationBucket(prevdb DurationBucket, closedStreams map[int]bool) DurationBucket {
	db := DurationBucket{
		streams:       make(map[int]bool),
		newStreams:    make(map[int]bool),
		closedStreams: make(map[int]bool),
	}

	// copy the previous streams over since they are still active
	// until the fin/rst is received
	for stream := range prevdb.streams {
		// this stream has already been closed, don't copy it over
		// this is a hack because something else is wonky with the stream
		// closing detection
		if _, ok := closedStreams[stream]; ok {
			fmt.Println("Found stream that was closed but was in the previous bucket", stream)
			continue
		}

		db.streams[stream] = true
	}

	return db
}

func (db *DurationBucket) AddFrame(frame Frame, seenStreams map[int]bool, closedStreams map[int]bool) {
	// if the TCP connection was closed, reset, or mysql quit was received
	// make the stream as closed
	if frame.TCPFin || frame.TCPReset || frame.MySQLCommand == 1 {
		delete(db.streams, frame.TCPStream)

		if _, ok := closedStreams[frame.TCPStream]; !ok {
			db.closedStreams[frame.TCPStream] = true
		}
		closedStreams[frame.TCPStream] = true
		return
	}

	// this stream has already been closed
	if _, ok := closedStreams[frame.TCPStream]; ok {
		return
	}

	if _, ok := seenStreams[frame.TCPStream]; !ok {
		db.newStreams[frame.TCPStream] = true
		seenStreams[frame.TCPStream] = true
	}

	db.streams[frame.TCPStream] = true
}

func (db *DurationBucket) CountConcurrent() int {
	return len(db.streams)
}

func (db *DurationBucket) CountNew() int {
	return len(db.newStreams)
}

func (db *DurationBucket) CountClosed() int {
	return len(db.closedStreams)
}
