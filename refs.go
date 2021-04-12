// SPDX-License-Identifier: MIT

// Package refs strives to offer a couple of types and corresponding encoding code to help other go-based ssb projects to talk about message, feed and blob references without pulling in all of go-ssb and it's network and database code.
package refs

import (
	"bytes"
	"encoding"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"golang.org/x/crypto/ed25519"
)

// Ref is the abstract interface all reference types should implement.
type Ref interface {
	Ref() string      // returns the full reference
	ShortRef() string // returns a shortend prefix

	Algo() RefAlgo
}

type RefAlgo string

// Some constant identifiers
const (
	RefAlgoFeedSSB1    RefAlgo = "ed25519" // ssb v1 (legacy, crappy encoding)
	RefAlgoMessageSSB1 RefAlgo = "sha256"  // scuttlebutt happend anyway
	RefAlgoBlobSSB1    RefAlgo = RefAlgoMessageSSB1

	RefAlgoCloakedGroup RefAlgo = "cloaked"

	RefAlgoFeedGabby    RefAlgo = "ggfeed-v1" // cbor based chain
	RefAlgoMessageGabby RefAlgo = "ggmsg-v1"

	RefAlgoContentGabby RefAlgo = "gabby-v1-content"
)

func ParseRef(str string) (Ref, error) {
	if len(str) == 0 {
		return nil, ErrInvalidRef
	}

	split := strings.Split(str[1:], ".")
	if len(split) < 2 {
		return nil, ErrInvalidRef
	}

	raw, err := base64.StdEncoding.DecodeString(split[0])
	if err != nil {
		return nil, fmt.Errorf("ssb-ref: b64 decode failed (%s): %w", err, ErrInvalidHash)
	}

	switch string(str[0]) {
	case "@":
		var algo RefAlgo
		switch RefAlgo(split[1]) {
		case RefAlgoFeedSSB1:
			algo = RefAlgoFeedSSB1
		case RefAlgoFeedGabby:
			algo = RefAlgoFeedGabby
		default:
			return nil, ErrInvalidRefAlgo
		}
		if n := len(raw); n != 32 {
			return nil, newFeedRefLenError(n)
		}
		newRef := FeedRef{algo: algo}
		copy(newRef.id[:], raw)
		return newRef, nil
	case "%":
		var algo RefAlgo
		switch RefAlgo(split[1]) {
		case RefAlgoMessageSSB1:
			algo = RefAlgoMessageSSB1
		case RefAlgoMessageGabby:
			algo = RefAlgoMessageGabby
		case RefAlgoCloakedGroup:
			algo = RefAlgoCloakedGroup
		default:
			return nil, ErrInvalidRefAlgo
		}
		if n := len(raw); n != 32 {
			return nil, newHashLenError(n)
		}
		newMsg := MessageRef{algo: algo}
		copy(newMsg.hash[:], raw)
		return newMsg, nil
	case "&":
		if RefAlgo(split[1]) != RefAlgoBlobSSB1 {
			return nil, ErrInvalidRefAlgo
		}
		if n := len(raw); n != 32 {
			return nil, newHashLenError(n)
		}
		newBlob := BlobRef{algo: RefAlgoBlobSSB1}
		copy(newBlob.hash[:], raw)
		return newBlob, nil
	}

	return nil, ErrInvalidRefType
}

// MessageRef defines the content addressed version of a ssb message, identified it's hash.
type MessageRef struct {
	hash [32]byte
	algo RefAlgo
}

// Ref prints the full identifieir
func (ref MessageRef) Ref() string {
	return fmt.Sprintf("%%%s.%s", base64.StdEncoding.EncodeToString(ref.hash[:]), ref.algo)
}

// ShortRef prints a shortend version
func (ref MessageRef) ShortRef() string {
	return fmt.Sprintf("<%%%s.%s>", base64.StdEncoding.EncodeToString(ref.hash[:3]), ref.algo)
}

func (ref MessageRef) Algo() RefAlgo {
	return ref.algo
}

func (ref MessageRef) Equal(other MessageRef) bool {
	if ref.algo != other.algo {
		return false
	}

	return bytes.Equal(ref.hash[:], other.hash[:])
}

var (
	_ encoding.TextMarshaler   = (*MessageRef)(nil)
	_ encoding.TextUnmarshaler = (*MessageRef)(nil)
)

func (mr MessageRef) MarshalText() ([]byte, error) {
	if len(mr.hash) == 0 {
		return []byte{}, nil
	}
	return []byte(mr.Ref()), nil
}

func (mr *MessageRef) UnmarshalText(text []byte) error {
	if len(text) == 0 {
		*mr = MessageRef{}
		return nil
	}
	newRef, err := ParseMessageRef(string(text))
	if err != nil {
		return fmt.Errorf("message(%s): unmarshal failed: %w", string(text), err)
	}
	*mr = newRef
	return nil
}

func (r *MessageRef) Scan(raw interface{}) error {
	switch v := raw.(type) {
	case []byte:
		if len(v) != 32 {
			return fmt.Errorf("msgRef/Scan: wrong length: %d", len(v))
		}
		copy(r.hash[:], v)
		r.algo = RefAlgoMessageSSB1
	case string:
		mr, err := ParseMessageRef(v)
		if err != nil {
			return fmt.Errorf("msgRef/Scan: failed to serialze from string: %w", err)
		}
		*r = mr
	default:
		return fmt.Errorf("msgRef/Scan: unhandled type %T", raw)
	}
	return nil
}

func ParseMessageRef(s string) (MessageRef, error) {
	ref, err := ParseRef(s)
	if err != nil {
		return MessageRef{}, fmt.Errorf("messageRef: failed to parse ref (%q): %w", s, err)
	}
	newRef, ok := ref.(MessageRef)
	if !ok {
		return MessageRef{}, fmt.Errorf("messageRef: not a message! %T", ref)
	}
	return newRef, nil
}

type MessageRefs []MessageRef

func (mr *MessageRefs) String() string {
	var s []string
	for _, r := range *mr {
		s = append(s, r.Ref())
	}
	return strings.Join(s, ", ")
}

func (mr *MessageRefs) UnmarshalJSON(text []byte) error {
	if len(text) == 0 {
		*mr = nil
		return nil
	}

	if bytes.Equal([]byte("[]"), text) || bytes.Equal([]byte("null"), text) {
		*mr = nil
		return nil
	}

	if bytes.HasPrefix(text, []byte("[")) && bytes.HasSuffix(text, []byte("]")) {

		elems := bytes.Split(text[1:len(text)-1], []byte(","))
		newArr := make([]MessageRef, len(elems))

		for i, e := range elems {
			var err error
			r := strings.TrimSpace(string(e))
			r = r[1 : len(r)-1] // remove quotes
			newArr[i], err = ParseMessageRef(r)
			if err != nil {
				return fmt.Errorf("messageRefs %d unmarshal failed: %w", i, err)
			}
		}

		*mr = newArr

	} else {
		newArr := make([]MessageRef, 1)

		var err error
		newArr[0], err = ParseMessageRef(string(text[1 : len(text)-1]))
		if err != nil {
			fmt.Println(string(text))
			return fmt.Errorf("messageRefs single unmarshal failed: %w", err)
		}

		*mr = newArr
	}
	return nil
}

// FeedRef defines a publickey as ID with a specific algorithm (currently only ed25519)
type FeedRef struct {
	id   [32]byte
	algo RefAlgo
}

func (ref FeedRef) PubKey() ed25519.PublicKey {
	return ref.id[:]
}

func (ref FeedRef) Ref() string {
	return fmt.Sprintf("@%s.%s", base64.StdEncoding.EncodeToString(ref.id[:]), ref.algo)
}

func (ref FeedRef) ShortRef() string {
	return fmt.Sprintf("<@%s.%s>", base64.StdEncoding.EncodeToString(ref.id[:3]), ref.algo)
}

func (ref FeedRef) Algo() RefAlgo {
	return ref.algo
}

func (ref FeedRef) Equal(b FeedRef) bool {
	// TODO: invset time in shs1.1 to signal the format correctly
	// if ref.Algo != b.Algo {
	// 	return false
	// }
	return bytes.Equal(ref.id[:], b.id[:])
}

func (ref FeedRef) Copy() *FeedRef {
	newRef, err := ParseFeedRef(ref.Ref())
	if err != nil {
		panic(fmt.Errorf("failed to copy existing ref: %w", err))
	}
	return newRef
}

var (
	_ encoding.TextMarshaler   = (*FeedRef)(nil)
	_ encoding.TextUnmarshaler = (*FeedRef)(nil)
)

func (fr FeedRef) MarshalText() ([]byte, error) {
	return []byte(fr.Ref()), nil
}

func (fr *FeedRef) UnmarshalText(text []byte) error {
	if len(text) == 0 {
		*fr = FeedRef{}
		return nil
	}
	newRef, err := ParseFeedRef(string(text))
	if err != nil {
		return err
	}
	*fr = *newRef
	return nil
}

func (r *FeedRef) Scan(raw interface{}) error {
	switch v := raw.(type) {
	// TODO: add an extra byte/flag bits to denote algo and types

	// case []byte:
	// 	if len(v) != 32 {
	// 		return fmt.Errorf("feedRef/Scan: wrong length: %d", len(v))
	// 	}
	// 	(*r).ID = v
	// 	(*r).Algo = "ed25519"

	case string:
		fr, err := ParseFeedRef(v)
		if err != nil {
			return fmt.Errorf("feedRef/Scan: failed to serialize from string: %w", err)
		}
		*r = *fr
	default:
		return fmt.Errorf("feedRef/Scan: unhandled type %T (see TODO)", raw)
	}
	return nil
}

// ParseFeedRef uses ParseRef and checks that it returns a *FeedRef
func ParseFeedRef(s string) (*FeedRef, error) {
	ref, err := ParseRef(s)
	if err != nil {
		return nil, fmt.Errorf("feedRef: couldn't parse %q: %w", s, err)
	}
	newRef, ok := ref.(*FeedRef)
	if !ok {
		return nil, fmt.Errorf("feedRef: not a feed! %T", ref)
	}
	return newRef, nil
}

// BlobRef defines a static binary attachment reference, identified it's hash.
type BlobRef struct {
	hash [32]byte
	algo RefAlgo
}

// Ref returns the BlobRef with the sigil &, it's base64 encoded hash and the used algo (currently only sha256)
func (ref BlobRef) Ref() string {
	return fmt.Sprintf("&%s.%s", base64.StdEncoding.EncodeToString(ref.hash[:]), ref.algo)
}

func (ref BlobRef) ShortRef() string {
	return fmt.Sprintf("<&%s.%s>", base64.StdEncoding.EncodeToString(ref.hash[:3]), ref.algo)
}

func (ref BlobRef) Algo() RefAlgo {
	return ref.algo
}

// ParseBlobRef uses ParseRef and checks that it returns a *BlobRef
func ParseBlobRef(s string) (*BlobRef, error) {
	ref, err := ParseRef(s)
	if err != nil {
		return nil, fmt.Errorf("blobRef: failed to parse %q: %w", s, err)
	}
	newRef, ok := ref.(*BlobRef)
	if !ok {
		return nil, fmt.Errorf("blobRef: not a blob! %T", ref)
	}
	return newRef, nil
}

func (ref BlobRef) Equal(b BlobRef) bool {
	if ref.algo != b.algo {
		return false
	}
	return bytes.Equal(ref.hash[:], b.hash[:])
}

func (br BlobRef) IsValid() error {
	if br.algo != RefAlgoBlobSSB1 {
		return fmt.Errorf("unknown hash algorithm %q", br.algo)
	}
	if len(br.hash) != 32 {
		return fmt.Errorf("expected hash length 32, got %v", len(br.hash))
	}
	return nil
}

// MarshalText encodes the BlobRef using Ref()
func (br BlobRef) MarshalText() ([]byte, error) {
	return []byte(br.Ref()), nil
}

// UnmarshalText uses ParseBlobRef
func (br *BlobRef) UnmarshalText(text []byte) error {
	if len(text) == 0 {
		*br = BlobRef{}
		return nil
	}
	newBR, err := ParseBlobRef(string(text))
	if err != nil {
		return fmt.Errorf(" BlobRef/UnmarshalText failed: %w", err)
	}
	*br = *newBR
	return nil
}

type AnyRef struct {
	r       Ref
	channel string
}

func (ar AnyRef) ShortRef() string {
	if ar.r == nil {
		panic("empty ref")
	}
	return ar.r.ShortRef()
}

func (ar AnyRef) Ref() string {
	if ar.r == nil {
		panic("empty ref")
	}
	return ar.r.Ref()
}

func (ar AnyRef) IsBlob() (*BlobRef, bool) {
	br, ok := ar.r.(*BlobRef)
	return br, ok
}

func (ar AnyRef) IsFeed() (*FeedRef, bool) {
	r, ok := ar.r.(*FeedRef)
	return r, ok
}

func (ar AnyRef) IsMessage() (*MessageRef, bool) {
	r, ok := ar.r.(*MessageRef)
	return r, ok
}

func (ar AnyRef) IsChannel() (string, bool) {
	ok := ar.channel != ""
	return ar.channel, ok
}

func (ar *AnyRef) MarshalJSON() ([]byte, error) {
	if ar.r == nil {
		if ar.channel != "" {
			return []byte(`"` + ar.channel + `"`), nil
		}
		return nil, ErrInvalidRef
	}
	refStr := ar.Ref()
	return []byte(`"` + refStr + `"`), nil
}

func (ar *AnyRef) UnmarshalJSON(b []byte) error {
	if string(b[0:2]) == `"#` {
		ar.channel = string(b[1 : len(b)-1])
		return nil
	}
	if len(b) < 53 {
		fmt.Printf("anyRef? %q %q\n", string(b), string(b[0:2]))
		return fmt.Errorf("ssb/anyRef: not a ref?: %w", ErrInvalidRef)
	}

	var refStr string
	err := json.Unmarshal(b, &refStr)
	if err != nil {
		return fmt.Errorf("ssb/anyRef: not a valid JSON string (%w)", err)
	}

	newRef, err := ParseRef(refStr)
	if err != nil {
		return err
	}

	ar.r = newRef
	return nil
}

var (
	_ json.Marshaler   = (*AnyRef)(nil)
	_ json.Unmarshaler = (*AnyRef)(nil)
	_ Ref              = (*AnyRef)(nil)
)
