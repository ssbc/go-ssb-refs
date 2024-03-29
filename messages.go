// SPDX-FileCopyrightText: 2022 Henry Bubert
//
// SPDX-License-Identifier: MIT

package refs

import (
	"encoding/json"
	"fmt"
	"time"
)

// Value describes a signed entry on a classical ssb feed.
// The name 'value' comes from seeing them in (hashed) 'key' and 'value' pairs from database query results.
type Value struct {
	Previous  *MessageRef     `json:"previous"`
	Author    FeedRef         `json:"author"`
	Sequence  int64           `json:"sequence"`
	Timestamp Millisecs       `json:"timestamp"`
	Hash      string          `json:"hash"`
	Content   json.RawMessage `json:"content"`
	Signature string          `json:"signature"`

	Meta map[string]interface{} `json:"meta,omitempty"`
}

// Message allows accessing message aspects without known the feed type
type Message interface {
	Key() MessageRef
	Previous() *MessageRef

	Seq() int64

	Claimed() time.Time
	Received() time.Time

	Author() FeedRef
	ContentBytes() []byte

	ValueContent() *Value
	ValueContentJSON() json.RawMessage
}

// Contact represents a (un)follow or (un)block message
type Contact struct {
	Type      string  `json:"type"`
	Contact   FeedRef `json:"contact"`
	Following bool    `json:"following"`
	Blocking  bool    `json:"blocking"`
}

// NewContactFollow returns a initialzed follow message
func NewContactFollow(who FeedRef) Contact {
	return Contact{
		Type:      "contact",
		Contact:   who,
		Following: true,
	}
}

// NewContactBlock returns a initialzed block message
func NewContactBlock(who FeedRef) Contact {
	return Contact{
		Type:     "contact",
		Contact:  who,
		Blocking: true,
	}
}

// UnmarshalJSON implements JSON deserialization of type:contact
func (c *Contact) UnmarshalJSON(b []byte) error {
	if len(b) > 0 && b[0] == '"' {
		return ErrWrongType{want: "contact", has: "private.box?"}
	}

	var potential map[string]interface{}
	err := json.Unmarshal(b, &potential)
	if err != nil {
		return fmt.Errorf("contact: map stage failed: %w", err)
	}

	t, ok := potential["type"].(string)
	if !ok {
		return ErrMalfromedMsg{"contact: no type on message", nil}
	}

	if t != "contact" {
		return ErrWrongType{want: "contact", has: t}
	}

	newC := new(Contact)

	contact, ok := potential["contact"].(string)
	if !ok {
		return ErrMalfromedMsg{"contact: no string contact field on type:contact", potential}
	}

	newC.Contact, err = ParseFeedRef(contact)
	if err != nil {
		return fmt.Errorf("contact: failed to parse contact field: %w", err)
	}

	newC.Following, _ = potential["following"].(bool)
	newC.Blocking, _ = potential["blocking"].(bool)

	if newC.Following && newC.Blocking {
		return fmt.Errorf("invalid contact message")
	}

	*c = *newC
	return nil
}

// About represents metadata updates about a feed like name, descriptin or image
type About struct {
	Type        string   `json:"type"`
	About       FeedRef  `json:"about"`
	Name        string   `json:"name,omitempty"`
	Description string   `json:"description,omitempty"`
	Image       *BlobRef `json:"image,omitempty"`
}

// NewAboutName creats a new message to update one's name
func NewAboutName(who FeedRef, name string) *About {
	return &About{
		Type:  "about",
		About: who,
		Name:  name,
	}
}

// NewAboutImage creats a new message to update one's image
func NewAboutImage(who FeedRef, img *BlobRef) *About {
	return &About{
		Type:  "about",
		About: who,
		Image: img,
	}
}

// UnmarshalJSON implements JSON deserialization of type:about
func (a *About) UnmarshalJSON(b []byte) error {
	var priv string
	err := json.Unmarshal(b, &priv)
	if err == nil {
		return ErrWrongType{want: "about", has: "private.box?"}
	}

	var potential map[string]interface{}
	err = json.Unmarshal(b, &potential)
	if err != nil {
		return fmt.Errorf("about: map stage failed: %w", err)
	}

	t, ok := potential["type"].(string)
	if !ok {
		return ErrMalfromedMsg{"about: no type on message", nil}
	}

	if t != "about" {
		return ErrWrongType{want: "about", has: t}
	}

	newA := new(About)

	about, ok := potential["about"].(string)
	if !ok {
		return ErrMalfromedMsg{"about: no string about field on type:about", potential}
	}

	newA.About, err = ParseFeedRef(about)
	if err != nil {
		return fmt.Errorf("about: who?: %w", err)
	}

	if newName, ok := potential["name"].(string); ok {
		newA.Name = newName
	}
	if newDesc, ok := potential["description"].(string); ok {
		newA.Description = newDesc
	}

	var newImgBlob string
	if img, ok := potential["image"].(string); ok {
		newImgBlob = img
	}
	if imgObj, ok := potential["image"].(map[string]interface{}); ok {
		lnk, ok := imgObj["link"].(string)
		if ok {
			newImgBlob = lnk
		}
	}
	if newImgBlob != "" {
		br, err := ParseBlobRef(newImgBlob)
		if err != nil {
			return fmt.Errorf("about: invalid image: %q: %w", newImgBlob, err)
		}
		newA.Image = &br
	}

	*a = *newA
	return nil
}

// Typed helps to quickly get the type of a message
type Typed struct {
	Value
	Content struct {
		Type string `json:"type"`
	} `json:"content"`
}

// ValuePost helps to deserialze a type:post message
type ValuePost struct {
	Value
	Content Post `json:"content"`
}

// NewPost creates a new Post with the text field set to the passed string
func NewPost(text string) Post {
	return Post{
		Type: "post",
		Text: text,
	}
}

// Post represents a textual (markdown) message with some metadata.
type Post struct {
	Type     string      `json:"type"`
	Text     string      `json:"text"`
	Root     *MessageRef `json:"root,omitempty"`
	Branch   MessageRefs `json:"branch,omitempty"`
	Mentions []Mention   `json:"mentions,omitempty"`

	Tangles Tangles `json:"tangles,omitempty"`

	// Recipients of a message
	Recps MessageRefs `json:"recps,omitempty"`
}

// Tangles represent a set of tangle information ala ssb-tangles v2 ( https://gitlab.com/tangle-js/tangle-graph )
// for general information about tangles, you might want to read:
// https://github.com/cn-uofbasel/ssbdrv/blob/1f7e6b11373ef6f73415f0e9c62f1ade29739251/doc/tangle.md
type Tangles map[string]TanglePoint

// TanglePoint represent a single reference point to a common message (root)
// and the _previous_ messages that were seen at the time
// in aggregate with other such messages this creates a _happend before_ relation betwee them.
type TanglePoint struct {
	Root     *MessageRef `json:"root"`
	Previous MessageRefs `json:"previous"`
}

// Mention can link feeds/authors by name, channels or other messages.
type Mention struct {
	Link AnyRef `json:"link,omitempty"`
	Name string `json:"name,omitempty"`
}

// NewMention creates a mention:name field that should be added to a message, like a post.
func NewMention(r Ref, name string) Mention {
	return Mention{Link: AnyRef{r: r}, Name: name}
}

// ValueVote is a convenience wrapper if the content is wrapped in a value
type ValueVote struct {
	Value
	Content Vote `json:"content"`
}

// Vote represents a 'like' message
type Vote struct {
	Type string `json:"type"`
	Vote struct {
		Expression string      `json:"expression"`
		Link       *MessageRef `json:"link"`
		Value      int         `json:"value"`
	} `json:"vote"`
}

// OldPubMessage is posted after a legacy invite is used to advertise a pub to your followers
type OldPubMessage struct {
	Type    string     `json:"type"`
	Address OldAddress `json:"address"`
}

// OldAddress uses an ssb.FeedRef as a key
// this is not strictly necessary since secret-handshake only cares about the 32bytes of public key (and not the feed type)
// newer type:address uses multiserver and not this nested object
type OldAddress struct {
	Host string  `json:"host"`
	Port int     `json:"port"`
	Key  FeedRef `json:"key"`
}

// KeyValueRaw uses json.RawMessage for the content portion, this helps of the content needs to be deserialzed manually or not at all
type KeyValueRaw struct {
	Key_      MessageRef `json:"key"` // Key_ is using the underline here to not conflict with the refs.Message interface (for history ceasons)
	Value     Value      `json:"value"`
	Timestamp Millisecs  `json:"timestamp"`
}

// KeyValueAsMap helps if there are no expectations about the content of a message
type KeyValueAsMap struct {
	Key       MessageRef `json:"key"`
	Value     Value      `json:"value"`
	Timestamp Millisecs  `json:"timestamp"`
}

var _ Message = (*KeyValueRaw)(nil)

// Seq implements the refs.Message interface
func (kvr KeyValueRaw) Seq() int64 {
	return kvr.Value.Sequence
}

// Key keyimplements the refs.Message interface
func (kvr KeyValueRaw) Key() MessageRef {
	return kvr.Key_
}

// Author implements the refs.Message interface
func (kvr KeyValueRaw) Author() FeedRef {
	return kvr.Value.Author
}

// Previous implements the refs.Message interface
func (kvr KeyValueRaw) Previous() *MessageRef {
	return kvr.Value.Previous
}

// Claimed implements the refs.Message interface
func (kvr KeyValueRaw) Claimed() time.Time {
	return time.Time(kvr.Value.Timestamp)
}

// Received implements the refs.Message interface
func (kvr KeyValueRaw) Received() time.Time {
	return time.Time(kvr.Timestamp)
}

// ContentBytes implements the refs.Message interface
func (kvr KeyValueRaw) ContentBytes() []byte {
	return kvr.Value.Content
}

// ValueContent implements the refs.Message interface
func (kvr KeyValueRaw) ValueContent() *Value {
	return &kvr.Value
}

// ValueContentJSON implements the refs.Message interface
func (kvr KeyValueRaw) ValueContentJSON() json.RawMessage {
	jsonB, err := json.Marshal(kvr.ValueContent())
	if err != nil {
		panic(err.Error())
	}

	return jsonB
}
