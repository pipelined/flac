package flac_test

import (
	"context"
	"os"
	"reflect"
	"testing"

	"pipelined.dev/audio/flac"
	"pipelined.dev/pipe"
	"pipelined.dev/pipe/mock"
)

const (
	channels   = 2
	bufferSize = 322
	samples    = 330528
	inputFile  = "_testdata/sample.flac"
)

func TestFlacPipe(t *testing.T) {
	in, _ := os.Open(inputFile)
	defer in.Close()

	sink := mock.Sink{}
	p, err := pipe.New(bufferSize,
		pipe.Line{
			Source: flac.Source(in),
			Sink:   sink.Sink(),
		},
	)
	assertEqual(t, "pipe err", err, nil)
	err = pipe.Wait(p.Start(context.Background()))
	assertEqual(t, "wait err", err, nil)
	assertEqual(t, "samples", sink.Samples, samples)

	err = in.Close()
	assertEqual(t, "close err", err, nil)
}

func assertEqual(t *testing.T, name string, result, expected interface{}) {
	t.Helper()
	if !reflect.DeepEqual(expected, result) {
		t.Fatalf("%v\nresult: \t%T\t%+v \nexpected: \t%T\t%+v", name, result, result, expected, expected)
	}
}
