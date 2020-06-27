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

// Source reads flac data from ReadSeeker.
type Source struct {
	io.Reader
	decoder *flac.Stream
	frame   *frame.Frame
	pos     int
	ints    signal.Signed
}

// Source starts the pump stage.
func (s Source) Source() pipe.SourceAllocatorFunc {
	return s.allocator
}

func (s *Source) allocator(bufferSize int) (pipe.Source, pipe.SignalProperties, error) {
	decoder, err := flac.New(s.Reader)
	if err != nil {
		return pipe.Source{}, pipe.SignalProperties{}, fmt.Errorf("error creating FLAC decoder: %w", err)
	}
	s.decoder = decoder
	channels := int(decoder.Info.NChannels)
	s.ints = signal.Allocator{
		Channels: channels,
		Length:   bufferSize,
		Capacity: bufferSize,
	}.Int32(signal.BitDepth(decoder.Info.BitsPerSample))
	return pipe.Source{
			SourceFunc: s.source,
			FlushFunc:  s.flush,
		},
		pipe.SignalProperties{
			SampleRate: signal.SampleRate(decoder.Info.SampleRate),
			Channels:   channels,
		}, nil
}

func (s *Source) source(floats signal.Floating) (read int, err error) {
outer:
	for {
		if s.frame == nil {
			if s.frame, err = s.decoder.ParseNext(); err != nil {
				if err == io.EOF {
					break // no more bytes available
				} else {
					return 0, fmt.Errorf("error reading FLAC frame: %w", err)
				}
			}
		}

		for s.pos < int(s.frame.BlockSize) {
			if read == s.ints.Len() {
				break outer
			}
			for sf := range s.frame.Subframes {
				s.ints.SetSample(read, int64(s.frame.Subframes[sf].Samples[s.pos]))
				read++
			}
			s.pos++
		}
		s.frame = nil
		s.pos = 0
	}
	if read == 0 {
		return 0, io.EOF
	}
	if read != s.ints.Len() {
		return signal.SignedAsFloating(s.ints.Slice(0, signal.ChannelLength(read, s.ints.Channels())), floats), nil
	}
	return signal.SignedAsFloating(s.ints, floats), nil
}

// Flush closes flac decoder.
func (s *Source) flush(context.Context) error {
	return s.decoder.Close()
}
