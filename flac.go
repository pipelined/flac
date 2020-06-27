package flac

import (
	"context"
	"fmt"
	"io"

	"github.com/mewkiz/flac"
	"github.com/mewkiz/flac/frame"

	"pipelined.dev/pipe"
	"pipelined.dev/signal"
)

type (
	// Pump reads flac data from ReadSeeker.
	Pump struct {
		io.Reader
		decoder *flac.Stream
	}
)

// Pump starts the pump stage.
func (p *Pump) Pump() pipe.SourceAllocatorFunc {
	return func(bufferSize int) (pipe.Source, pipe.SignalProperties, error) {
		decoder, err := flac.New(p.Reader)
		if err != nil {
			return pipe.Source{}, pipe.SignalProperties{}, fmt.Errorf("error creating FLAC decoder: %w", err)
		}
		p.decoder = decoder
		channels := int(decoder.Info.NChannels)
		var (
			frame *frame.Frame // frames needs to be known between calls
			pos   int          // position of last read sample within frame
		)
		ints := signal.Allocator{
			Channels: channels,
			Length:   bufferSize,
			Capacity: bufferSize,
		}.Int32(signal.BitDepth(decoder.Info.BitsPerSample))
		return pipe.Source{
				SourceFunc: func(floats signal.Floating) (int, error) {
					var read int
				outer:
					for {
						if frame == nil {
							if frame, err = p.decoder.ParseNext(); err != nil {
								if err == io.EOF {
									break // no more bytes available
								} else {
									return 0, fmt.Errorf("error reading FLAC frame: %w", err)
								}
							}
						}

						for pos < int(frame.BlockSize) {
							if read == ints.Len() {
								break outer
							}
							for sf := range frame.Subframes {
								ints.SetSample(read, int64(frame.Subframes[sf].Samples[pos]))
								read++
							}
							pos++
						}
						frame = nil
						pos = 0
					}
					if read == 0 {
						return 0, io.EOF
					}
					if read != ints.Len() {
						return signal.SignedAsFloating(ints.Slice(0, signal.ChannelLength(read, ints.Channels())), floats), nil
					}
					return signal.SignedAsFloating(ints, floats), nil
				},
				FlushFunc: p.flush,
			},
			pipe.SignalProperties{
				SampleRate: signal.SampleRate(decoder.Info.SampleRate),
				Channels:   channels,
			}, nil
	}
}

// Flush closes flac decoder.
func (p *Pump) flush(context.Context) error {
	return p.decoder.Close()
}
