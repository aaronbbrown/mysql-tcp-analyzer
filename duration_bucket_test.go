package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestBucket(t *testing.T) {
	tests := []struct {
		duration time.Duration
		expected int
	}{
		{duration: 50 * time.Millisecond, expected: 0},
		{duration: 100 * time.Millisecond, expected: 1},
		{duration: 101 * time.Millisecond, expected: 1},
		{duration: 250 * time.Millisecond, expected: 2},
		{duration: 500 * time.Millisecond, expected: 5},
		{duration: 2000 * time.Millisecond, expected: 20},
	}

	db := NewDurationBuckets(100 * time.Millisecond)
	for _, test := range tests {
		assert.Equal(t, test.expected, db.bucket(test.duration))
	}
}
