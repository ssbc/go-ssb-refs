// SPDX-License-Identifier: MIT

package refs

import (
	"encoding/json"
	stderr "errors"
	"fmt"

	"github.com/pkg/errors"
)

// Common errors for invalid references
var (
	ErrInvalidRef     = stderr.New("ssb: Invalid Ref")
	ErrInvalidRefType = stderr.New("ssb: Invalid Ref Type")
	ErrInvalidRefAlgo = stderr.New("ssb: Invalid Ref Algo")
	ErrInvalidSig     = stderr.New("ssb: Invalid Signature")
	ErrInvalidHash    = stderr.New("ssb: Invalid Hash")
)

// ErrRefLen is returned when a parsed reference was too short.
type ErrRefLen struct {
	algo string
	n    int
}

func (e ErrRefLen) Error() string {
	return fmt.Sprintf("ssb: Invalid reference len for %s: %d", e.algo, e.n)
}

// NewFeedRefLenError returns a new ErrRefLen error for a feed
func newFeedRefLenError(n int) error {
	return ErrRefLen{algo: RefAlgoFeedSSB1, n: n}
}

func newHashLenError(n int) error {
	return ErrRefLen{algo: RefAlgoMessageSSB1, n: n}
}

// IsMessageUnusable checks if an error is ErrWrongType, ErrMalfromedMsg or *json.SyntaxError
func IsMessageUnusable(err error) bool {
	cause := errors.Cause(err)
	_, is := cause.(ErrWrongType)
	if is {
		return true
	}
	_, is = cause.(ErrMalfromedMsg)
	if is {
		return true
	}
	_, is = cause.(*json.SyntaxError)
	return is
}

// ErrMalfromedMsg is returned if a message has invalid values
type ErrMalfromedMsg struct {
	reason string
	m      map[string]interface{}
}

func (emm ErrMalfromedMsg) Error() string {
	s := "ErrMalfromedMsg: " + emm.reason
	if emm.m != nil {
		s += fmt.Sprintf(" %+v", emm.m)
	}
	return s
}

// ErrWrongType is returned if a certain type:value was expected on a message.
type ErrWrongType struct {
	has, want string
}

func (ewt ErrWrongType) Error() string {
	return fmt.Sprintf("ErrWrongType: want: %s has: %s", ewt.want, ewt.has)
}

var ErrUnuspportedFormat = fmt.Errorf("ssb: unsupported format")
