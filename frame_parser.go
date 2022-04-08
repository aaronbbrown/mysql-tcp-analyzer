package main

import (
	"strconv"
	"time"
)

type transactionId struct {
	// the index of the first frame in the transaction
	Index int
	// the number of levels of nesting for nested transactions
	NestingLevels int
}

type FrameParser struct {
	Frames       Frames
	Transactions Transactions
	// a buffer of TCP Stream IDs that have not yet seen a mysql response
	// with the key being the stream ID and the value the index of the original
	// frame in the Frames slice
	unRespondedStreams     map[int]int
	openTransactionStreams map[int]*transactionId
}

func NewFrameParser() FrameParser {
	return FrameParser{
		Transactions:           NewTransactions(),
		unRespondedStreams:     make(map[int]int),
		openTransactionStreams: make(map[int]*transactionId),
	}
}

func (fp *FrameParser) ParseRawFrames(rawframes []rawframe) error {
	for index, rawframe := range rawframes {
		frame, err := fp.parseLayers(rawframe.Source.Layers, index)
		if err != nil {
			return err
		}
		fp.Frames = append(fp.Frames, frame)
	}

	// clean up any transactions that have not had a response
	// this is probably the case when a connection was killed or
	// the tcpdump ended before the transaction was completed
	for _, txid := range fp.openTransactionStreams {
		fp.Transactions.Delete(txid.Index)
	}

	return nil
}

func (fp *FrameParser) parseLayers(layers Layers, index int) (*Frame, error) {
	var frame Frame

	if val, ok := layers["frame.number"]; ok {
		number, err := strconv.Atoi(val[0])
		if err != nil {
			return &frame, err
		}
		frame.Number = number
	}

	if val, ok := layers["tcp.stream"]; ok {
		stream, err := strconv.Atoi(val[0])
		if err != nil {
			return &frame, err
		}
		frame.TCPStream = stream
	}

	if val, ok := layers["frame.time_relative"]; ok {
		tr, err := strconv.ParseFloat(val[0], 64)
		trint := int64(tr * 1e9)
		if err != nil {
			return &frame, err
		}
		frame.TimeRelative = time.Duration(trint) * time.Nanosecond
	}

	if val, ok := layers["mysql.command"]; ok {
		command, err := strconv.Atoi(val[0])
		if err != nil {
			return &frame, err
		}
		frame.MySQLCommand = command
	}

	if val, ok := layers["tcp.flags.fin"]; ok {
		var err error
		frame.TCPFin, err = strconv.ParseBool(val[0])
		if err != nil {
			return &frame, err
		}
	}

	if val, ok := layers["tcp.flags.reset"]; ok {
		var err error
		frame.TCPReset, err = strconv.ParseBool(val[0])
		if err != nil {
			return &frame, err
		}
	}

	if val, ok := layers["mysql.query"]; ok {
		frame.MySQLQuery = NewMySQLQuery(val[0])
		// add it to the list of unacknowledged queries
		fp.unRespondedStreams[frame.TCPStream] = index

		if frame.MySQLQuery.Fingerprint == "begin" {
			txid, ok := fp.openTransactionStreams[frame.TCPStream]
			if ok {
				// this is a nested transaction
				txid.NestingLevels++
			} else {
				// this is a new transaction
				txid := transactionId{Index: index}
				fp.openTransactionStreams[frame.TCPStream] = &txid
				transaction := NewTransaction(txid.Index)
				fp.Transactions.Add(&transaction)
			}
		}

		if txid, ok := fp.openTransactionStreams[frame.TCPStream]; ok {
			// in a transaction, so add it to the transaction obj
			if err := fp.Transactions.AddFrame(txid.Index, &frame); err != nil {
				return &frame, err
			}

			if frame.MySQLQuery.Fingerprint == "commit" || frame.MySQLQuery.Fingerprint == "rollback" {
				txid.NestingLevels--

				if txid.NestingLevels < 0 {
					// exited the transaction, so remove it from the list and record timing
					delete(fp.openTransactionStreams, frame.TCPStream)
				}
			}
		}
	}

	// select queries get a payload back
	if val, ok := layers["mysql.payload"]; ok {
		if len(val) > 0 {
			if idx, ok := fp.unRespondedStreams[frame.TCPStream]; ok {
				took := time.Duration(frame.TimeRelative - fp.Frames[idx].TimeRelative)
				fp.Frames[idx].MySQLQuery.Duration = took
				delete(fp.unRespondedStreams, frame.TCPStream)
			}
		}
	}

	// non-select queries just get a response code
	if val, ok := layers["mysql.response_code"]; ok {
		if len(val) > 0 {
			if idx, ok := fp.unRespondedStreams[frame.TCPStream]; ok {
				took := time.Duration(frame.TimeRelative - fp.Frames[idx].TimeRelative)
				fp.Frames[idx].MySQLQuery.Duration = took
				delete(fp.unRespondedStreams, frame.TCPStream)
			}
		}
	}

	lost := false
	if val, ok := layers["tcp.analysis.lost_segment"]; ok && len(val) > 0 {
		if val[0] > "0" {
			lost = true
		}
	}
	if val, ok := layers["tcp.analysis.ack_lost_segment"]; ok && len(val) > 0 {
		if val[0] > "0" {
			lost = true
		}
	}

	if lost {
		if txid, ok := fp.openTransactionStreams[frame.TCPStream]; ok {
			// transaction got lost in the data, so remove it from the list
			delete(fp.openTransactionStreams, frame.TCPStream)
			fp.Transactions.Delete(txid.Index)
		}
	}

	return &frame, nil
}
