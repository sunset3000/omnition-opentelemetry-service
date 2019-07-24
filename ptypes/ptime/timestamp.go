package ptime

import (
	"time"

	ptypes "github.com/gogo/protobuf/types"
)

// TimeToTimestamp converts a time.Time to a Timestamp pointer.
func TimeToTimestamp(t time.Time) *ptypes.Timestamp {
	if ts, err := ptypes.TimestampProto(t); err == nil {
		return ts
	}
	return nil
}
