package flac

import (
	"io"

	"github.com/mewkiz/flac"
	"github.com/mewkiz/flac/frame"

	"github.com/pipelined/signal"
)

type (
	// Pump reads flac data from ReadSeeker.
	Pump struct {
		io.Reader
		d *flac.Stream
	}
)

// Pump starts the pump process once executed, wav attributes are accessible.
func (p *Pump) Pump(sourceID string) (func(bufferSize int) ([][]float64, error), int, int, error) {
	decoder, err := flac.New(p)
	if err != nil {
		return nil, 0, 0, err
	}
	p.d = decoder

	sampleRate, numChannels, bitDepth := int(decoder.Info.SampleRate), int(decoder.Info.NChannels), signal.BitDepth(decoder.Info.BitsPerSample)

	// frames needs to be known between calls
	var (
		frame             *frame.Frame
		pos               int // position in frame
		ints              []int
		currentBufferSize int
	)
	return func(bufferSize int) ([][]float64, error) {
		if currentBufferSize != bufferSize {
			currentBufferSize = bufferSize
			ints = make([]int, numChannels*bufferSize)
		}
		var (
			err  error
			read int
		)
		for read < len(ints) {
			// read next frame if current is finished
			if frame == nil {
				if frame, err = p.d.ParseNext(); err != nil {
					if err == io.EOF {
						break // no more bytes available
					} else {
						return nil, err
					}
				}
				pos = 0
			}

			// read samples
			for pos < int(frame.BlockSize) {
				for _, s := range frame.Subframes {
					ints[read] = int(s.Samples[pos])
					read++
				}
				pos++
			}
			frame = nil
		}
		// nothing was read
		if read == 0 {
			return nil, io.EOF
		}

		// trim and convert the buffer
		b := signal.InterInt{
			Data:        ints[:read],
			NumChannels: numChannels,
			BitDepth:    bitDepth,
		}.AsFloat64()

		if b.Size() != bufferSize {
			return b, io.ErrUnexpectedEOF
		}

		return b, nil
	}, sampleRate, numChannels, nil
}

// Flush closes flac decoder.
func (p *Pump) Flush() error {
	return p.d.Close()
}
