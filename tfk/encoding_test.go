// SPDX-FileCopyrightText: 2022 Henry Bubert
//
// SPDX-License-Identifier: CC0-1.0

package tfk_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	refs "go.mindeco.de/ssb-refs"
	"go.mindeco.de/ssb-refs/tfk"
)

func mustMakeFeed(t *testing.T, hash []byte, algo refs.RefAlgo) refs.FeedRef {
	mr, err := refs.NewFeedRefFromBytes(hash, algo)
	if err != nil {
		t.Fatal(err)
	}
	return mr
}

func TestFormatFeedRef(t *testing.T) {
	type testcase struct {
		name string
		in   refs.FeedRef
		out  []byte
		err  error
	}

	tcs := []testcase{
		{
			name: "ed25519",
			in:   mustMakeFeed(t, seq(0, 32), "ed25519"),
			out:  append([]byte{tfk.TypeFeed, tfk.FormatFeedEd25519}, seq(0, 32)...),
		},
		{
			name: "gabby",
			in:   mustMakeFeed(t, seq(0, 32), "gabbygrove-v1"),
			out:  append([]byte{tfk.TypeFeed, tfk.FormatFeedGabbyGrove}, seq(0, 32)...),
		},
		{
			name: "bamboo",
			in:   mustMakeFeed(t, seq(0, 32), "bamboo"),
			out:  append([]byte{tfk.TypeFeed, tfk.FormatFeedBamboo}, seq(0, 32)...),
		},
		{
			name: "metafeed",
			in:   mustMakeFeed(t, seq(0, 32), "bendybutt-v1"),
			out:  append([]byte{tfk.TypeFeed, tfk.FormatFeedBendyButt}, seq(0, 32)...),
		},
		{
			name: "tooShort",
			out:  nil,
			err:  tfk.ErrTooShort,
		},
		{
			name: "unknown-algo",
			in:   mustMakeFeed(t, seq(0, 32), "the-future"),
			out:  append([]byte{tfk.TypeFeed, 42}, seq(0, 32)...),
			err:  tfk.ErrUnhandledFormat,
		},
	}

	for _, tc := range tcs {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			var f tfk.Feed
			err := f.UnmarshalBinary(tc.out)
			if tc.err != nil {
				require.Equal(t, tc.err.Error(), err.Error())
				return
			}
			require.NoError(t, err)

			feedRef, err := f.Feed()
			require.NoError(t, err)

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

func mustMakeMessage(t *testing.T, hash []byte, algo refs.RefAlgo) refs.MessageRef {
	mr, err := refs.NewMessageRefFromBytes(hash, algo)
	if err != nil {
		t.Fatal(err)
	}
	return mr
}

func TestFormatMessageRef(t *testing.T) {
	type testcase struct {
		name string
		in   refs.MessageRef
		out  []byte
		err  error
	}

	tcs := []testcase{
		{
			name: "sha256",
			in:   mustMakeMessage(t, seq(0, 32), "sha256"),

			out: append([]byte{tfk.TypeMessage, tfk.FormatMessageSHA256}, seq(0, 32)...),
		},
		{
			name: "gabby",
			in:   mustMakeMessage(t, seq(0, 32), "gabbygrove-v1"),

			out: append([]byte{tfk.TypeMessage, tfk.FormatMessageGabbyGrove}, seq(0, 32)...),
		},
		{
			name: "bamboo",
			in:   mustMakeMessage(t, seq(0, 64), "bamboo"),
			out:  append([]byte{tfk.TypeMessage, tfk.FormatMessageBamboo}, seq(0, 64)...),
		},
		{
			name: "metafeed",
			in:   mustMakeMessage(t, seq(0, 32), "bendybutt-v1"),
			out:  append([]byte{tfk.TypeMessage, tfk.FormatMessageMetaFeed}, seq(0, 32)...),
		},
		{
			name: "tooShort",

			out: nil,
			err: tfk.ErrTooShort,
		},
		{
			name: "unknown-algo",
			in:   mustMakeMessage(t, seq(0, 32), "the-future"),

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
				return
			}

			require.NoError(t, err)

			msgRef, err := m.Message()
			require.NoError(t, err)

			require.True(t, msgRef.Equal(tc.in), "got %s and %s", msgRef.String(), tc.in.String())

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
