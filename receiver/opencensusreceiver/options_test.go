package opencensusreceiver

import (
	"reflect"
	"testing"
)

func TestNoopOption(t *testing.T) {
	plainReceiver := new(Receiver)

	subjectReceiver := new(Receiver)
	opts := []Option{noopOption(0), noopOption(0)}
	for _, opt := range opts {
		opt.withReceiver(subjectReceiver)
	}

	if !reflect.DeepEqual(plainReceiver, subjectReceiver) {
		t.Fatalf("noopOption has side effects\nGot:  %+v\nWant: %+v", subjectReceiver, plainReceiver)
	}
}
