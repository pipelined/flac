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
func (p *Pump) Pump(sourceID string, bufferSize int) (func() ([][]float64, error), int, int, error) {
	decoder, err := flac.New(p)
	if err != nil {
		return nil, 0, 0, err
	}
	p.d = decoder

	sampleRate, numChannels, bitDepth := int(decoder.Info.SampleRate), int(decoder.Info.NChannels), signal.BitDepth(decoder.Info.BitsPerSample)

	// frames needs to be known between calls
	var (
		f   *frame.Frame
		pos int
	)
	// allocate buffer
	intBuf := make([]int, numChannels*bufferSize)
	return func() ([][]float64, error) {
		var err error
		var read int
		for read < len(intBuf) {
			// read next frame if current is finished
			if f == nil {
				if f, err = p.d.ParseNext(); err != nil {
					if err == io.EOF {
						break // no more bytes available
					} else {
						return nil, err
					}
				}
				pos = 0
			}

			// read samples
			for pos < int(f.BlockSize) {
				for _, s := range f.Subframes {
					intBuf[read] = int(s.Samples[pos])
					read++
				}
				pos++
			}
			f = nil
		}
		// nothing was read
		if read == 0 {
			return nil, io.EOF
		}

		// trim and convert the buffer
		b := signal.InterInt{
			Data:        intBuf[:read],
			NumChannels: numChannels,
			BitDepth:    bitDepth,
		}.AsFloat64()

		if b.Size() != bufferSize {
			return b, io.ErrUnexpectedEOF
		}

		return b, nil
	}, sampleRate, numChannels, nil
}
