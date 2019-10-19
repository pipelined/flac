package flac_test

import (
	"io"
	"os"
	"testing"

	"github.com/pipelined/flac"
	"github.com/pipelined/signal"
	"github.com/pipelined/wav"

	"github.com/stretchr/testify/assert"
)

var (
	bufferSize  = 512
	flacSamples = 330528
	flacSample  = "_testdata/sample.flac"
	wav1        = "_testdata/out1.wav"
)

func TestFlacPipe(t *testing.T) {
	tests := []struct {
		inPath  string
		outPath string
	}{
		{
			inPath:  flacSample,
			outPath: wav1,
		},
	}

	for _, test := range tests {
		inFile, err := os.Open(test.inPath)
		assert.Nil(t, err)
		pump := flac.Pump{Reader: inFile}

		outFile, err := os.Create(test.outPath)
		assert.Nil(t, err)
		sink := wav.Sink{
			WriteSeeker: outFile,
			BitDepth:    16,
		}

		pumpFn, sampleRate, numChannles, err := pump.Pump("")
		assert.NotNil(t, pumpFn)
		assert.Nil(t, err)
		t.Logf("SampleRate: %d NumChannels: %d\n", sampleRate, numChannles)

		sinkFn, err := sink.Sink("", sampleRate, numChannles)
		assert.NotNil(t, sinkFn)
		assert.Nil(t, err)

		buf := signal.Float64Buffer(numChannles, bufferSize)
		messages, samples := 0, 0
		for {
			err = pumpFn(buf)
			if err != nil {
				break
			}
			_ = sinkFn(buf)
			messages++
			if buf != nil {
				samples += len(buf[0])
			}
		}
		assert.Equal(t, io.EOF, err)

		// assert.Equal(t, wavMessages, messages)
		assert.Equal(t, flacSamples, samples)

		err = sink.Flush("")
		assert.Nil(t, err)

		err = inFile.Close()
		assert.Nil(t, err)
		err = outFile.Close()
		assert.Nil(t, err)
	}
}
