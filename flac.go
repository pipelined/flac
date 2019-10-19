package flac

import (
	"fmt"
	"io"

	"github.com/mewkiz/flac"
	"github.com/mewkiz/flac/frame"

	"github.com/pipelined/signal"
)

type (
	// Pump reads flac data from ReadSeeker.
	Pump struct {
		io.Reader
		decoder *flac.Stream
	}
)

// Pump starts the pump process once executed, wav attributes are accessible.
func (p *Pump) Pump(sourceID string) (func(b signal.Float64) error, signal.SampleRate, int, error) {
	decoder, err := flac.New(p.Reader)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("error creating flac decoder: %w", err)
	}
	p.decoder = decoder

	sampleRate := signal.SampleRate(decoder.Info.SampleRate)
	numChannels := int(decoder.Info.NChannels)
	bitDepth := signal.BitDepth(decoder.Info.BitsPerSample)

	// frames needs to be known between calls
	var (
		frame *frame.Frame
		pos   int // position in frame
	)
	ints := signal.InterInt{
		NumChannels: numChannels,
		BitDepth:    bitDepth,
	}
	return func(b signal.Float64) error {
		if ints.Size() != b.Size() {
			ints.Data = make([]int, numChannels*b.Size())
		}
		var (
			read int
		)
		for read < len(ints.Data) {
			// read next frame if current is finished
			if frame == nil {
				if frame, err = p.decoder.ParseNext(); err != nil {
					if err == io.EOF {
						break // no more bytes available
					} else {
						return fmt.Errorf("error reading FLAC frame: %w", err)
					}
				}
				pos = 0
			}

			// read samples
			for pos < int(frame.BlockSize) {
				for _, s := range frame.Subframes {
					ints.Data[read] = int(s.Samples[pos])
					read++
				}
				pos++
			}
			frame = nil
		}
		if read == 0 {
			return io.EOF
		}

		// trim buffer.
		if read != len(ints.Data) {
			ints.Data = ints.Data[:read]
			for i := range b {
				b[i] = b[i][:ints.Size()]
			}
		}

		// copy to the output.
		ints.CopyToFloat64(b)

		return nil
	}, sampleRate, numChannels, nil
}

// Flush closes flac decoder.
func (p *Pump) Flush() error {
	return p.decoder.Close()
}
