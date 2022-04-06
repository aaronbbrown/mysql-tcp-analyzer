package main

import "time"

type Layers map[string][]string

type Frames []*Frame

type Frame struct {
	Number       int
	TimeRelative time.Duration
	TCPStream    int
	MySQLCommand int
	MySQLQuery   MySQLQuery
}

func (f *Frames) CountByTag() map[string]int {
	result := make(map[string]int)
	for _, frame := range *f {
		for k, v := range frame.MySQLQuery.Tags {
			result[k+":"+v] += 1
		}
	}
	return result
}

func (f *Frames) QueriesForTag(key, value string) map[string]int {
	result := make(map[string]int)

	for _, frame := range *f {
		if frame.MySQLQuery.Tags[key] == value {
			result[frame.MySQLQuery.Fingerprint] += 1
		}
	}

	return result
}

func (f *Frames) TagsForFingerprint(fingerprint string) map[string]int {
	result := make(map[string]int)

	for _, frame := range *f {
		if frame.MySQLQuery.Fingerprint != fingerprint {
			continue
		}

		for k, v := range frame.MySQLQuery.Tags {
			result[k+":"+v] += 1
		}
	}

	return result
}
