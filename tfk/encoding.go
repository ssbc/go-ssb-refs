// SPDX-FileCopyrightText: 2022 Henry Bubert
//
// SPDX-License-Identifier: CC0-1.0

package tfk

import (
	"encoding"
	"fmt"

	refs "github.com/ssbc/go-ssb-refs"
)

// Encode returns type-format-key bytes for supported references.
// Currently only *refs.MessageRef and *refs.FeedRef
func Encode(r refs.Ref) ([]byte, error) {
	var mb encoding.BinaryMarshaler

	switch tv := r.(type) {

	case *refs.MessageRef:
		m, err := MessageFromRef(*tv)
		if err != nil {
			return nil, err
		}
		mb = m

	case refs.MessageRef:
		m, err := MessageFromRef(tv)
		if err != nil {
			return nil, err
		}
		mb = m

	case refs.FeedRef:
		f, err := FeedFromRef(tv)
		if err != nil {
			return nil, err
		}
		mb = f

	default:
		return nil, fmt.Errorf("ssb/tfk: unhandled reference type: %s (%T)", r.Algo(), r)
	}

	return mb.MarshalBinary()
}
