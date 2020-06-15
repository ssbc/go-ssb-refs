// SPDX-License-Identifier: MIT

package refs

import (
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
)

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

type ErrWrongType struct {
	has, want string
}

func (ewt ErrWrongType) Error() string {
	return fmt.Sprintf("ErrWrongType: want: %s has: %s", ewt.want, ewt.has)
}

var ErrUnuspportedFormat = errors.Errorf("ssb: unsupported format")
