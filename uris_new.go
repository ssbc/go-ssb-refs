// SPDX-FileCopyrightText: 2021 Henry Bubert
//
// SPDX-License-Identifier: MIT

package refs

import (
	"encoding/base64"
	"fmt"
	"net"
	"net/url"

	"golang.org/x/crypto/ed25519"
)

// URIOption allow to customaize an experimental ssb-uri
type URIOption func(e *ExperimentalURI) error

// MSAddr adds a multiserver address to an experimental URI
func MSAddr(hostAndPort string, pubKey ed25519.PublicKey) URIOption {
	return func(e *ExperimentalURI) error {

		host, port, err := net.SplitHostPort(hostAndPort)
		if err != nil {
			return err
		}

		// cant import go.mindeco.de/ssb-multiserver
		msAddr := fmt.Sprintf("net:%s:%s~shs:%s", host, port, base64.StdEncoding.EncodeToString(pubKey))
		e.params.Set("msaddr", msAddr)
		e.params.Set("action", "add-pub")
		return nil
	}
}

// RoomInvite adds a rooms2 invite to an experimental URI
func RoomInvite(code string) URIOption {
	return func(e *ExperimentalURI) error {
		e.params.Set("invite", code)
		e.params.Set("action", "join-room")
		return nil
	}
}

// RoomAlias adds a rooms2 alias to an experimental URI
func RoomAlias(roomID, userID, alias, signature string) URIOption {
	return func(e *ExperimentalURI) error {
		e.params.Set("roomID", roomID)
		e.params.Set("userID", userID)
		e.params.Set("action", "consume-alias")
		return nil
	}
}

// NewExperimentalURI constructs an experimental ssb uri from a slice of options
func NewExperimentalURI(opts ...URIOption) (*url.URL, error) {
	var e ExperimentalURI

	e.params = make(url.Values)

	for i, opt := range opts {
		err := opt(&e)
		if err != nil {
			return nil, fmt.Errorf("NewExperimentalURI: option %d failed: %w", i, err)
		}
	}

	var u url.URL
	u.Scheme = "ssb"
	u.RawQuery = e.params.Encode()
	u.Opaque = "experimental"

	return &u, nil
}
