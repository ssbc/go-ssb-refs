// SPDX-FileCopyrightText: 2022 Henry Bubert
//
// SPDX-License-Identifier: MIT

package refs

import (
	"bytes"
	"fmt"
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// just message, feed, blob. Canonical and Experimental
func TestParseSimpleURIs(t *testing.T) {

	type tcase struct {
		name  string
		input string

		want  CanonicalURI
		kind  Kind
		sigil string

		err error // should it error?
	}

	var cases = []tcase{
		{
			name:  "canon message (ssb v1)",
			input: "ssb:message/sha256/g3hPVPDEO1Aj_uPl0-J2NlhFB2bbFLIHlty-YuqFZ3w=",
			want: CanonicalURI{ref: MessageRef{
				hash: [32]byte{131, 120, 79, 84, 240, 196, 59, 80, 35, 254, 227, 229, 211, 226, 118, 54, 88, 69, 7, 102, 219, 20, 178, 7, 150, 220, 190, 98, 234, 133, 103, 124},
				algo: RefAlgoMessageSSB1,
			}},
			sigil: `%g3hPVPDEO1Aj/uPl0+J2NlhFB2bbFLIHlty+YuqFZ3w=.sha256`,
			kind:  KindMessage,
		},

		{
			name:  "canon message (bendy)",
			input: "ssb:message/bendybutt-v1/PR2-btDEO1AjXuPl0TJ2N_hFB2bbFLIHlty0VF1nctw=",
			want: CanonicalURI{ref: MessageRef{
				hash: [32]uint8{0x3d, 0x1d, 0xbe, 0x6e, 0xd0, 0xc4, 0x3b, 0x50, 0x23, 0x5e, 0xe3, 0xe5, 0xd1, 0x32, 0x76, 0x37, 0xf8, 0x45, 0x7, 0x66, 0xdb, 0x14, 0xb2, 0x7, 0x96, 0xdc, 0xb4, 0x54, 0x5d, 0x67, 0x72, 0xdc},
				algo: RefAlgoMessageBendyButt,
			}},
			sigil: `%PR2+btDEO1AjXuPl0TJ2N/hFB2bbFLIHlty0VF1nctw=.bendybutt-v1`,
			kind:  KindMessage,
		},

		{
			name:  "canon feed (classic)",
			input: "ssb:feed/ed25519/-oaWWDs8g73EZFUMfW37R_ULtFEjwKN_DczvdYihjbU=",
			want: CanonicalURI{ref: FeedRef{
				id:   [32]byte{0xfa, 0x86, 0x96, 0x58, 0x3b, 0x3c, 0x83, 0xbd, 0xc4, 0x64, 0x55, 0xc, 0x7d, 0x6d, 0xfb, 0x47, 0xf5, 0xb, 0xb4, 0x51, 0x23, 0xc0, 0xa3, 0x7f, 0xd, 0xcc, 0xef, 0x75, 0x88, 0xa1, 0x8d, 0xb5},
				algo: RefAlgoFeedSSB1,
			}},
			sigil: `@+oaWWDs8g73EZFUMfW37R/ULtFEjwKN/DczvdYihjbU=.ed25519`,
			kind:  KindFeed,
		},

		{
			name:  "canon feed (bendy butt)",
			input: "ssb:feed/bendybutt-v1/APaWWDs8g73EZFUMfW37RBULtFEjwKNbDczvdYiRXtA=",
			want: CanonicalURI{ref: FeedRef{
				id:   [32]uint8{0x0, 0xf6, 0x96, 0x58, 0x3b, 0x3c, 0x83, 0xbd, 0xc4, 0x64, 0x55, 0xc, 0x7d, 0x6d, 0xfb, 0x44, 0x15, 0xb, 0xb4, 0x51, 0x23, 0xc0, 0xa3, 0x5b, 0xd, 0xcc, 0xef, 0x75, 0x88, 0x91, 0x5e, 0xd0},
				algo: RefAlgoFeedBendyButt,
			}},
			sigil: `@APaWWDs8g73EZFUMfW37RBULtFEjwKNbDczvdYiRXtA=.bendybutt-v1`,
			kind:  KindFeed,
		},

		{
			name:  "canon blob",
			input: "ssb:blob/sha256/sbBmsB7XWvmIzkBzreYcuzPpLtpeCMDIs6n_OJGSC1U=",
			want: CanonicalURI{ref: BlobRef{
				hash: [32]byte{0xb1, 0xb0, 0x66, 0xb0, 0x1e, 0xd7, 0x5a, 0xf9, 0x88, 0xce, 0x40, 0x73, 0xad, 0xe6, 0x1c, 0xbb, 0x33, 0xe9, 0x2e, 0xda, 0x5e, 0x8, 0xc0, 0xc8, 0xb3, 0xa9, 0xff, 0x38, 0x91, 0x92, 0xb, 0x55},
				algo: RefAlgoBlobSSB1,
			}},
			sigil: `&sbBmsB7XWvmIzkBzreYcuzPpLtpeCMDIs6n/OJGSC1U=.sha256`,
			kind:  KindBlob,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := require.New(t)
			a := assert.New(t)

			got, err := ParseURI(tc.input)
			if tc.err == nil {
				r.NoError(err)
				a.Equal(tc.kind, got.Kind(), "wrong kind")
				a.Equal(tc.want, got)
				var (
					ref Ref
					ok  bool
				)
				switch tc.kind {
				case KindFeed:
					ref, ok = got.Feed()
					r.True(ok)
				case KindMessage:
					ref, ok = got.Message()
					r.True(ok)
				case KindBlob:
					ref, ok = got.Blob()
					r.True(ok)
				default:
					t.Fatal("oops? unhandled kind")
				}

				a.Equal(tc.sigil, ref.Sigil(), "wrong sigil")

				a.Equal(tc.input, got.String(), "did not turn back into the uri")
			} else {
				r.EqualError(err, tc.err.Error())
			}
		})
	}
}

func ExampleNewExperimentalURI() {
	check := func(err error) {
		if err != nil {
			log.Fatal(err)
		}
	}
	fmt.Println() // emptyline so that Output: block looks nicer

	var testRef FeedRef
	copy(testRef.id[:], bytes.Repeat([]byte("A"), 32))

	// msaddr
	msaddr, err := NewExperimentalURI(
		MSAddr("host:port", testRef.PubKey()),
	)
	check(err)
	fmt.Println("simple multiserver address:")
	fmt.Println(msaddr.String())

	// room invite
	roomInvite, err := NewExperimentalURI(
		MSAddr("host:port", testRef.PubKey()),
		RoomInvite("some-code"),
	)
	check(err)
	fmt.Println("rooms2 invite:")
	fmt.Println(roomInvite.String())

	// room alias
	roomAlias, err := NewExperimentalURI(
		MSAddr("host:port", testRef.PubKey()),
		RoomAlias("roomID", "userID", "alias", "sig"),
	)
	check(err)
	fmt.Println("rooms2 alias:")
	fmt.Println(roomAlias.String())

	// Output:
	// simple multiserver address:
	// ssb:experimental?action=add-pub&msaddr=net%3Ahost%3Aport~shs%3AQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUE%3D
	// rooms2 invite:
	// ssb:experimental?action=join-room&invite=some-code&msaddr=net%3Ahost%3Aport~shs%3AQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUE%3D
	// rooms2 alias:
	// ssb:experimental?action=consume-alias&msaddr=net%3Ahost%3Aport~shs%3AQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUE%3D&roomID=roomID&userID=userID
}
