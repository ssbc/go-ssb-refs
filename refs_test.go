// SPDX-License-Identifier: MIT

package refs

import (
	"encoding/json"
	"errors"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseRef(t *testing.T) {
	a := assert.New(t)
	var tcases = []struct {
		ref  string
		err  error
		want Ref
	}{
		{"", ErrInvalidRef, nil},
		{"xxxx", ErrInvalidRefType, nil},
		{"+xxx.foo", ErrInvalidRefType, nil},
		{"@xxx.foo", ErrInvalidHash, nil},

		{"%wayTooShort.sha256", ErrInvalidHash, nil},
		{"&tooShort.sha256", newHashLenError(6), nil},
		{"@tooShort.ed25519", newFeedRefLenError(6), nil},
		{"&c29tZU5vbmVTZW5zZQo=.sha256", newHashLenError(14), nil},

		{"@ye+QM09iPcDJD6YvQYjoQc7sLF/IFhmNbEqgdzQo3lQ=.ed25519", nil, FeedRef{
			id:   [32]byte{201, 239, 144, 51, 79, 98, 61, 192, 201, 15, 166, 47, 65, 136, 232, 65, 206, 236, 44, 95, 200, 22, 25, 141, 108, 74, 160, 119, 52, 40, 222, 84},
			algo: RefAlgoFeedSSB1,
		}},

		// {"@ye+QM09iPcDJD6YvQYjoQc7sLF/IFhmNbEqgdzQo3lQ=.bamboo?", nil, &FeedRef{
		// 	id:   [32]byte{201, 239, 144, 51, 79, 98, 61, 192, 201, 15, 166, 47, 65, 136, 232, 65, 206, 236, 44, 95, 200, 22, 25, 141, 108, 74, 160, 119, 52, 40, 222, 84},
		// 	algo: RefAlgoFeed?????,
		// }},

		{"ssb:feed/gabbygrove-v1/ye-QM09iPcDJD6YvQYjoQc7sLF_IFhmNbEqgdzQo3lQ=", nil, FeedRef{
			id:   [32]byte{201, 239, 144, 51, 79, 98, 61, 192, 201, 15, 166, 47, 65, 136, 232, 65, 206, 236, 44, 95, 200, 22, 25, 141, 108, 74, 160, 119, 52, 40, 222, 84},
			algo: RefAlgoFeedGabby,
		}},

		{"&84SSLNv5YdDVTdSzN2V1gzY5ze4lj6tYFkNyT+P28Qs=.sha256", nil, BlobRef{
			hash: [32]byte{243, 132, 146, 44, 219, 249, 97, 208, 213, 77, 212, 179, 55, 101, 117, 131, 54, 57, 205, 238, 37, 143, 171, 88, 22, 67, 114, 79, 227, 246, 241, 11},
			algo: RefAlgoBlobSSB1,
		}},

		{"%2jDrrJEeG7PQcCLcisISqarMboNpnwyfxLnwU1ijOjc=.sha256", nil, MessageRef{
			hash: [32]byte{218, 48, 235, 172, 145, 30, 27, 179, 208, 112, 34, 220, 138, 194, 18, 169, 170, 204, 110, 131, 105, 159, 12, 159, 196, 185, 240, 83, 88, 163, 58, 55},
			algo: RefAlgoMessageSSB1,
		}},

		// {`ssb:message/cloaked/vof09Dhy3YUat1ylIUVGaCjotAFxE8iGbF6QxLlCWWc=`, nil, MessageRef{
		// 	hash: [32]byte{190, 135, 244, 244, 56, 114, 221, 133, 26, 183, 92, 165, 33, 69, 70, 104, 40, 232, 180, 1, 113, 19, 200, 134, 108, 94, 144, 196, 185, 66, 89, 103},
		// 	algo: RefAlgoCloakedGroup,
		// }},

		{"ssb:message/gabbygrove-v1/2jDrrJEeG7PQcCLcisISqarMboNpnwyfxLnwU1ijOjc=", nil, MessageRef{
			hash: [32]byte{218, 48, 235, 172, 145, 30, 27, 179, 208, 112, 34, 220, 138, 194, 18, 169, 170, 204, 110, 131, 105, 159, 12, 159, 196, 185, 240, 83, 88, 163, 58, 55},
			algo: RefAlgoMessageGabby,
		}},
	}
	for i, tc := range tcases {
		r, err := ParseRef(tc.ref)
		if tc.err == nil {
			if !a.NoError(err, "got error on test %d (%v)", i+1, tc.ref) {
				continue
			}
			input := a.Equal(tc.ref, tc.want.String(), "test %d input<>output failed", i+1)
			want := a.Equal(tc.want.String(), r.String(), "test %d re-encode failed", i+1)
			if !input || !want {
				t.Logf("%+v", r)
			}
			t.Log(i+1, r.ShortSigil())
		} else {
			a.True(errors.Is(err, tc.err), "%d wrong error: %s", i+1, err)
		}
	}
}

func TestAnyRef(t *testing.T) {
	a := assert.New(t)
	r := require.New(t)
	var tcases = []struct {
		ref  string
		want Ref
	}{

		{"@ye+QM09iPcDJD6YvQYjoQc7sLF/IFhmNbEqgdzQo3lQ=.ed25519", FeedRef{
			id:   [32]byte{201, 239, 144, 51, 79, 98, 61, 192, 201, 15, 166, 47, 65, 136, 232, 65, 206, 236, 44, 95, 200, 22, 25, 141, 108, 74, 160, 119, 52, 40, 222, 84},
			algo: RefAlgoFeedSSB1,
		}},

		{"%AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=.sha256", MessageRef{
			hash: [32]byte{},
			algo: RefAlgoMessageSSB1,
		}},

		{"&84SSLNv5YdDVTdSzN2V1gzY5ze4lj6tYFkNyT+P28Qs=.sha256", BlobRef{
			hash: [32]byte{243, 132, 146, 44, 219, 249, 97, 208, 213, 77, 212, 179, 55, 101, 117, 131, 54, 57, 205, 238, 37, 143, 171, 88, 22, 67, 114, 79, 227, 246, 241, 11},
			algo: RefAlgoBlobSSB1,
		}},

		{"%2jDrrJEeG7PQcCLcisISqarMboNpnwyfxLnwU1ijOjc=.sha256", MessageRef{
			hash: [32]byte{218, 48, 235, 172, 145, 30, 27, 179, 208, 112, 34, 220, 138, 194, 18, 169, 170, 204, 110, 131, 105, 159, 12, 159, 196, 185, 240, 83, 88, 163, 58, 55},
			algo: RefAlgoMessageSSB1,
		}},

		{"ssb:message/bendybutt-v1/2jDrrJEeG7PQcCLcisISqarMboNpnwyfxLnwU1ijOjc=", MessageRef{
			hash: [32]byte{218, 48, 235, 172, 145, 30, 27, 179, 208, 112, 34, 220, 138, 194, 18, 169, 170, 204, 110, 131, 105, 159, 12, 159, 196, 185, 240, 83, 88, 163, 58, 55},
			algo: RefAlgoMessageBendyButt,
		}},

		{"ssb:message/gabbygrove-v1/2jDrrJEeG7PQcCLcisISqarMboNpnwyfxLnwU1ijOjc=", MessageRef{
			hash: [32]byte{218, 48, 235, 172, 145, 30, 27, 179, 208, 112, 34, 220, 138, 194, 18, 169, 170, 204, 110, 131, 105, 159, 12, 159, 196, 185, 240, 83, 88, 163, 58, 55},
			algo: RefAlgoMessageGabby,
		}},
	}
	for i, tc := range tcases {

		var testPost Post
		testPost.Type = "test"
		testPost.Text = strconv.Itoa(i)
		testPost.Mentions = []Mention{
			NewMention(tc.want, "test"),
		}

		body, err := json.Marshal(testPost)
		r.NoError(err)

		var gotPost Post
		err = json.Unmarshal(body, &gotPost)
		r.NoError(err, "case %d unmarshal", i)

		r.Len(gotPost.Mentions, 1)
		a.Equal(tc.want.String(), gotPost.Mentions[0].Link.String(), "test %d re-encode failed", i)

		if i == 2 {
			br, ok := gotPost.Mentions[0].Link.IsBlob()
			a.True(ok, "not a blob?")
			a.NotNil(br)
		}
	}
}

func TestParseBranches(t *testing.T) {
	r := require.New(t)

	var got struct {
		Refs MessageRefs `json:"refs"`
	}
	var input = []byte(`{
		"refs": "%HG1p299uO2nCenG6YwR3DG33lLpcALAS/PI6/BP5dB0=.sha256"
	}`)

	err := json.Unmarshal(input, &got)
	r.NoError(err)
	r.Equal(1, len(got.Refs))
	r.Equal(got.Refs[0].String(), "%HG1p299uO2nCenG6YwR3DG33lLpcALAS/PI6/BP5dB0=.sha256")

	var asArray = []byte(`{
		"refs": [
			"%hCM+q/zsq8vseJKwIAAJMMdsAmWeSfG9cs8ed3uOXCc=.sha256",
			"%yJAzwPO7HSjvHRp7wrVGO4sbo9GHSwKk0BXOSiUr+bo=.sha256"
		]
	}`)

	err = json.Unmarshal(asArray, &got)
	require.NoError(t, err)
	r.Equal(2, len(got.Refs))
	r.Equal(got.Refs[0].String(), `%hCM+q/zsq8vseJKwIAAJMMdsAmWeSfG9cs8ed3uOXCc=.sha256`)
	r.Equal(got.Refs[1].String(), `%yJAzwPO7HSjvHRp7wrVGO4sbo9GHSwKk0BXOSiUr+bo=.sha256`)

	var empty = []byte(`{
		"refs": []
	}`)

	err = json.Unmarshal(empty, &got)
	require.NoError(t, err)
	r.Equal(0, len(got.Refs))
}
