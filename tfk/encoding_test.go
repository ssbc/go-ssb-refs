package tfk_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	refs "go.mindeco.de/ssb-refs"
	"go.mindeco.de/ssb-refs/tfk"
)

func TestFormatFeedRef(t *testing.T) {
	type testcase struct {
		name string
		in   *refs.FeedRef
		out  []byte
		err  error
	}

	tcs := []testcase{
		{
			name: "ed25519",
			in: &refs.FeedRef{
				Algo: "ed25519",
				ID:   seq(0, 32),
			},
			out: append([]byte{tfk.TypeFeed, tfk.FormatFeedEd25519}, seq(0, 32)...),
		},
		{
			name: "gabby",
			in: &refs.FeedRef{
				Algo: "ggfeed-v1",
				ID:   seq(0, 32),
			},
			out: append([]byte{tfk.TypeFeed, tfk.FormatFeedGabbyGrove}, seq(0, 32)...),
		},
		{
			name: "tooShort",
			in: &refs.FeedRef{
				Algo: "tooShort",
				ID:   nil,
			},
			out: nil,
			err: tfk.ErrTooShort,
		},
		{
			name: "unknown-algo",
			in: &refs.FeedRef{
				Algo: "the-future",
				ID:   seq(0, 32),
			},
			out: append([]byte{tfk.TypeFeed, 42}, seq(0, 32)...),
			err: tfk.ErrUnhandledFormat,
		},
	}

	for _, tc := range tcs {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			var f tfk.Feed
			err := f.UnmarshalBinary(tc.out)
			if tc.err != nil {
				require.Equal(t, tc.err.Error(), err.Error())
				require.Nil(t, f.Feed())
				return
			}
			require.NoError(t, err)

			feedRef := f.Feed()
			require.NotNil(t, feedRef)
			require.True(t, feedRef.Equal(tc.in))

			encoded, err := f.MarshalBinary()
			if tc.err != nil {
				require.Equal(t, tc.err, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, encoded, tc.out)
			}
		})
	}
}

func TestFormatMessageRef(t *testing.T) {
	type testcase struct {
		name string
		in   *refs.MessageRef
		out  []byte
		err  error
	}

	tcs := []testcase{
		{
			name: "sha256",
			in: &refs.MessageRef{
				Algo: "sha256",
				Hash: seq(0, 32),
			},
			out: append([]byte{tfk.TypeMessage, tfk.FormatMessageSHA256}, seq(0, 32)...),
		},
		{
			name: "gabby",
			in: &refs.MessageRef{
				Algo: "ggmsg-v1",
				Hash: seq(0, 32),
			},
			out: append([]byte{tfk.TypeMessage, tfk.FormatMessageGabbyGrove}, seq(0, 32)...),
		},
		{
			name: "tooShort",
			in: &refs.MessageRef{
				Algo: "tooShort",
				Hash: nil,
			},
			out: nil,
			err: tfk.ErrTooShort,
		},
		{
			name: "unknown-algo",
			in: &refs.MessageRef{
				Algo: "the-future",
				Hash: seq(0, 32),
			},
			out: append([]byte{tfk.TypeMessage, 42}, seq(0, 32)...),
			err: tfk.ErrUnhandledFormat,
		},
	}

	for _, tc := range tcs {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			var m tfk.Message
			err := m.UnmarshalBinary(tc.out)
			if tc.err != nil {
				require.Equal(t, tc.err, err)
				require.Nil(t, m.Message())
				return
			}

			require.NoError(t, err)

			msgRef := m.Message()
			require.NotNil(t, msgRef)
			require.True(t, msgRef.Equal(*tc.in), "got %s and %s", msgRef.Ref(), tc.in.Ref())

			encoded, err := m.MarshalBinary()
			if tc.err != nil {
				require.Equal(t, tc.err, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, encoded, tc.out)
			}
		})
	}
}

// utils

func seq(start, end int) []byte {
	out := make([]byte, end-start)
	for i := range out {
		out[i] = byte(i + start)
	}
	return out
}
