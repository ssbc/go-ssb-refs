// Package tfk implements the type-format-key encoding for SSB references.
//
// See https://github.com/ssbc/envelope-spec/ ... encoding/tfk.md
package tfk

import "errors"

// Type are the type-format-key type values
const (
	TypeFeed uint8 = iota
	TypeMessage
	TypeBlob
	TypeDiffieHellmanKey
)

// These are the type-format-key feed format values
const (
	FormatFeedEd25519 uint8 = iota
	FormatFeedGabbyGrove
)

// These are the type-format-key message format values
const (
	FormatMessageSHA256 uint8 = iota
	FormatMessageGabbyGrove
)

// Common errors
var (
	ErrTooShort        = errors.New("ssb/tfk: data too short")
	ErrWrongType       = errors.New("ssb/tfk: unexpected type value")
	ErrUnhandledFormat = errors.New("ssb/tfk: unhandled format value")
)
