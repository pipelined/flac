package flac

import (
	"context"
	"fmt"
	"io"

	"github.com/mewkiz/flac"
	"github.com/mewkiz/flac/frame"
	"pipelined.dev/pipe"
	"pipelined.dev/pipe/mutable"
	"pipelined.dev/signal"
)

// Source decodes PCM from flac data io.Reader. Source returns new Source
// allocator function.
func Source(r io.Reader) pipe.SourceAllocatorFunc {
	return func(mctx mutable.Context, bufferSize int) (pipe.Source, error) {
		decoder, err := flac.New(r)
		if err != nil {
			return pipe.Source{}, fmt.Errorf("error creating FLAC decoder: %w", err)
		}
		ints := signal.Allocator{
			Channels: int(decoder.Info.NChannels),
			Length:   bufferSize,
			Capacity: bufferSize,
		}.Int32(signal.BitDepth(decoder.Info.BitsPerSample))

		return pipe.Source{
				SourceFunc: source(decoder, ints),
				FlushFunc:  decoderFlusher(decoder),
				SignalProperties: pipe.SignalProperties{
					SampleRate: signal.Frequency(decoder.Info.SampleRate),
					Channels:   int(decoder.Info.NChannels),
				},
			},
			nil
	}
}

func source(decoder *flac.Stream, ints signal.Signed) pipe.SourceFunc {
	state := &readState{}
	return func(floats signal.Floating) (int, error) {
		read := 0
		for read < ints.Len() {
			if err := state.nextFrame(decoder); err != nil {
				if err == io.EOF {
					break
				}
				return 0, fmt.Errorf("error reading FLAC frame: %w", err)
			}

			read = state.readFrame(ints, read)
		}
		if read == 0 {
			return 0, io.EOF
		}
		if read != ints.Len() {
			return signal.SignedAsFloating(ints.Slice(0, signal.ChannelLength(read, ints.Channels())), floats), nil
		}
		return signal.SignedAsFloating(ints, floats), nil
	}
}

func decoderFlusher(decoder *flac.Stream) pipe.FlushFunc {
	return func(context.Context) error {
		return decoder.Close()
	}
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
