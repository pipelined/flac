package flac_test

import (
	"context"
	"os"
	"reflect"
	"testing"

	"pipelined.dev/flac"
	"pipelined.dev/pipe"
	"pipelined.dev/pipe/mock"
	"pipelined.dev/signal"
	"pipelined.dev/wav"
)

const (
	channels   = 2
	bufferSize = 322
	samples    = 330528
	inputFile  = "_testdata/sample.flac"
	outputFile = "_testdata/out.wav"
)

func TestFlacPipe(t *testing.T) {
	in, _ := os.Open(inputFile)
	outFile, _ := os.Create(outputFile)

	source := flac.Source{Reader: in}
	processor := &mock.Processor{}
	sink := wav.Sink{
		WriteSeeker: outFile,
		BitDepth:    signal.BitDepth16,
	}
	line, err := pipe.Routing{
		Source:     source.Source(),
		Processors: pipe.Processors(processor.Processor()),
		Sink:       sink.Sink(),
	}.Line(bufferSize)

	p := pipe.New(context.Background(), pipe.WithLines(line))
	err = p.Wait()
	assertEqual(t, "wait err", err, nil)
	assertEqual(t, "samples", processor.Samples, samples)

	err = in.Close()
	assertEqual(t, "close err", err, nil)
}

func assertEqual(t *testing.T, name string, result, expected interface{}) {
	t.Helper()
	if !reflect.DeepEqual(expected, result) {
		t.Fatalf("%v\nresult: \t%T\t%+v \nexpected: \t%T\t%+v", name, result, result, expected, expected)
	}
}
