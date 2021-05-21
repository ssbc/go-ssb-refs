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
}

// Some constant identifiers
const (
	RefAlgoFeedSSB1    = "ed25519" // ssb v1 (legacy, crappy encoding)
	RefAlgoMessageSSB1 = "sha256"  // scuttlebutt happend anyway
	RefAlgoBlobSSB1    = RefAlgoMessageSSB1

	RefAlgoCloakedGroup = "cloaked"

	RefAlgoFeedGabby    = "ggfeed-v1" // cbor based chain
	RefAlgoMessageGabby = "ggmsg-v1"

	RefAlgoContentGabby = "gabby-v1-content"
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
		var algo string
		switch split[1] {
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
		return &FeedRef{
			ID:   raw,
			Algo: algo,
		}, nil
	case "%":
		var algo string
		switch split[1] {
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
		return &MessageRef{
			Hash: raw,
			Algo: algo,
		}, nil
	case "&":
		if split[1] != RefAlgoBlobSSB1 {
			return nil, ErrInvalidRefAlgo
		}
		if n := len(raw); n != 32 {
			return nil, newHashLenError(n)
		}
		return &BlobRef{
			Hash: raw,
			Algo: RefAlgoBlobSSB1,
		}, nil
	}

	return nil, ErrInvalidRefType
}

// MessageRef defines the content addressed version of a ssb message, identified it's hash.
type MessageRef struct {
	Hash []byte
	Algo string
}

// Ref prints the full identifieir
func (ref MessageRef) Ref() string {
	return fmt.Sprintf("%%%s.%s", base64.StdEncoding.EncodeToString(ref.Hash), ref.Algo)
}

// ShortRef prints a shortend version
func (ref MessageRef) ShortRef() string {
	return fmt.Sprintf("<%%%s.%s>", base64.StdEncoding.EncodeToString(ref.Hash[:3]), ref.Algo)
}

func (ref MessageRef) Equal(other *MessageRef) bool {
	if other == nil {
		return false
	}

	if ref.Algo != other.Algo {
		return false
	}

	if ref.Hash == nil || other.Hash == nil {
		return true
	}

	return bytes.Equal(ref.Hash, other.Hash)
}

var (
	_ encoding.TextMarshaler   = (*MessageRef)(nil)
	_ encoding.TextUnmarshaler = (*MessageRef)(nil)
)

func (mr MessageRef) MarshalText() ([]byte, error) {
	if len(mr.Hash) == 0 {
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
	*mr = *newRef
	return nil
}

func (r *MessageRef) Scan(raw interface{}) error {
	switch v := raw.(type) {
	case []byte:
		if len(v) != 32 {
			return fmt.Errorf("msgRef/Scan: wrong length: %d", len(v))
		}
		r.Hash = v
		r.Algo = RefAlgoMessageSSB1
	case string:
		mr, err := ParseMessageRef(v)
		if err != nil {
			return fmt.Errorf("msgRef/Scan: failed to serialze from string: %w", err)
		}
		*r = *mr
	default:
		return fmt.Errorf("msgRef/Scan: unhandled type %T", raw)
	}
	return nil
}

func ParseMessageRef(s string) (*MessageRef, error) {
	ref, err := ParseRef(s)
	if err != nil {
		return nil, fmt.Errorf("messageRef: failed to parse ref (%q): %w", s, err)
	}
	newRef, ok := ref.(*MessageRef)
	if !ok {
		return nil, fmt.Errorf("messageRef: not a message! %T", ref)
	}
	return newRef, nil
}

type MessageRefs []*MessageRef

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
		newArr := make([]*MessageRef, len(elems))

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
		newArr := make([]*MessageRef, 1)

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
	ID   []byte
	Algo string
}

func (ref FeedRef) PubKey() ed25519.PublicKey {
	return ref.ID
}

func (ref FeedRef) Ref() string {
	return fmt.Sprintf("@%s.%s", base64.StdEncoding.EncodeToString(ref.ID), ref.Algo)
}

func (ref FeedRef) ShortRef() string {
	return fmt.Sprintf("<@%s.%s>", base64.StdEncoding.EncodeToString(ref.ID[:3]), ref.Algo)
}

func (ref FeedRef) Equal(b *FeedRef) bool {
	// TODO: invset time in shs1.1 to signal the format correctly
	// if ref.Algo != b.Algo {
	// 	return false
	// }
	return bytes.Equal(ref.ID, b.ID)
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
	Hash []byte
	Algo string
}

// Ref returns the BlobRef with the sigil &, it's base64 encoded hash and the used algo (currently only sha256)
func (ref BlobRef) Ref() string {
	return fmt.Sprintf("&%s.%s", base64.StdEncoding.EncodeToString(ref.Hash), ref.Algo)
}

func (ref BlobRef) ShortRef() string {
	return fmt.Sprintf("<&%s.%s>", base64.StdEncoding.EncodeToString(ref.Hash[:3]), ref.Algo)
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

func (ref BlobRef) Equal(b *BlobRef) bool {
	if ref.Algo != b.Algo {
		return false
	}
	return bytes.Equal(ref.Hash, b.Hash)
}

func (br BlobRef) IsValid() error {
	if br.Algo != "sha256" {
		return fmt.Errorf("unknown hash algorithm %q", br.Algo)
	}
	if len(br.Hash) != 32 {
		return fmt.Errorf("expected hash length 32, got %v", len(br.Hash))
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

// ContentRef defines the hashed content of a message
type ContentRef struct {
	Hash []byte
	Algo string
}

func (ref ContentRef) Ref() string {
	return fmt.Sprintf("!%s.%s", base64.StdEncoding.EncodeToString(ref.Hash), ref.Algo)
}

func (ref ContentRef) ShortRef() string {
	return fmt.Sprintf("<!%s.%s>", base64.StdEncoding.EncodeToString(ref.Hash[:3]), ref.Algo)
}

func (ref ContentRef) MarshalBinary() ([]byte, error) {
	switch ref.Algo {
	case RefAlgoContentGabby:
		return append([]byte{0x02}, ref.Hash...), nil
	default:
		return nil, fmt.Errorf("contentRef/Marshal: invalid binref type: %s", ref.Algo)
	}
}

func (ref *ContentRef) UnmarshalBinary(data []byte) error {
	if n := len(data); n != 33 {
		return fmt.Errorf("contentRef: invalid len:%d", n)
	}
	var newRef ContentRef
	newRef.Hash = make([]byte, 32)
	switch data[0] {
	case 0x02:
		newRef.Algo = RefAlgoContentGabby
	default:
		return fmt.Errorf("unmarshal: invalid contentRef type: %x", data[0])
	}
	n := copy(newRef.Hash, data[1:])
	if n != 32 {
		return fmt.Errorf("unmarshal: invalid contentRef size: %d", n)
	}
	*ref = newRef
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
