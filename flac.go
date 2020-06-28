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

// Source decodes PCM from flac data io.Reader.
type Source struct {
	io.Reader
	decoder *flac.Stream
	state   *readState
	ints    signal.Signed
}

// Source returns new Source allocator function.
func (s Source) Source() pipe.SourceAllocatorFunc {
	return s.allocator
}

func (s Source) allocator(bufferSize int) (pipe.Source, pipe.SignalProperties, error) {
	decoder, err := flac.New(s.Reader)
	if err != nil {
		return pipe.Source{}, pipe.SignalProperties{}, fmt.Errorf("error creating FLAC decoder: %w", err)
	}
	s.decoder = decoder
	s.ints = signal.Allocator{
		Channels: int(decoder.Info.NChannels),
		Length:   bufferSize,
		Capacity: bufferSize,
	}.Int32(signal.BitDepth(decoder.Info.BitsPerSample))
	s.state = &readState{}
	return pipe.Source{
			SourceFunc: s.source,
			FlushFunc:  s.flush,
		},
		pipe.SignalProperties{
			SampleRate: signal.SampleRate(decoder.Info.SampleRate),
			Channels:   int(decoder.Info.NChannels),
		}, nil
}

func (s Source) source(floats signal.Floating) (int, error) {
	read := 0
	for read < s.ints.Len() {
		if err := s.state.nextFrame(s.decoder); err != nil {
			if err == io.EOF {
				break
			}
			return 0, fmt.Errorf("error reading FLAC frame: %w", err)
		}

		read = s.state.readFrame(s.ints, read)
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
func (s Source) flush(context.Context) error {
	return s.decoder.Close()
}

type readState struct {
	*frame.Frame
	pos int
}

func (state *readState) nextFrame(stream *flac.Stream) error {
	if state.Frame == nil || state.pos == int(state.Frame.BlockSize) {
		f, err := stream.ParseNext()
		if err != nil {
			return err
		}
		state.Frame = f
		state.pos = 0
	}
	return nil
}

func (state *readState) readFrame(ints signal.Signed, read int) int {
	for state.pos < int(state.Frame.BlockSize) {
		if read == ints.Len() {
			break
		}
		for sf := range state.Frame.Subframes {
			ints.SetSample(read, int64(state.Frame.Subframes[sf].Samples[state.pos]))
			read++
		}
		state.pos++
	}
	return read
}
