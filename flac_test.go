package flac_test

import (
	"io"
	"os"
	"testing"

	"pipelined.dev/flac"
	"pipelined.dev/signal"

	"github.com/stretchr/testify/assert"
)

var (
	bufferSize  = 512
	flacSamples = 330528
	flacSample  = "_testdata/sample.flac"
)

func TestFlacPipe(t *testing.T) {
	in, err := os.Open(flacSample)
	assert.Nil(t, err)
	pump := flac.Pump{Reader: in}

	pumpFn, sampleRate, numChannles, err := pump.Pump("")
	assert.NotNil(t, pumpFn)
	assert.Nil(t, err)
	t.Logf("SampleRate: %d NumChannels: %d\n", sampleRate, numChannles)

	buf := signal.Float64Buffer(numChannles, bufferSize)
	messages, samples := 0, 0
	for {
		err = pumpFn(buf)
		if err != nil {
			break
		}
		messages++
		if buf != nil {
			samples += len(buf[0])
		}
	}
	assert.Equal(t, io.EOF, err)

	// assert.Equal(t, wavMessages, messages)
	assert.Equal(t, flacSamples, samples)

	err = in.Close()
	assert.Nil(t, err)
}
