package tfk

import (
	"fmt"

	refs "go.mindeco.de/ssb-refs"
)

type Message struct{ value }

func MessageFromRef(r *refs.MessageRef) (*Message, error) {
	var m Message
	m.tipe = TypeMessage

	// TODO: bamboo
	if n := len(r.Hash); n != 32 {
		return nil, fmt.Errorf("ssb/tfk: unexpected value length %d: %w", n, ErrTooShort)
	}

	m.key = make([]byte, 32)
	copy(m.key, r.Hash)

	switch r.Algo {
	case refs.RefAlgoMessageSSB1:
		m.format = FormatMessageSHA256
	case refs.RefAlgoMessageGabby:
		m.format = FormatMessageGabbyGrove
	default:
		return nil, fmt.Errorf("format value: %x: %w", m.format, ErrUnhandledFormat)
	}
	return &m, nil
}

// MarshalBinary returns the type-format-key encoding for a message.
func (msg *Message) MarshalBinary() ([]byte, error) {
	if msg.tipe != TypeMessage {
		return nil, ErrWrongType
	}
	if msg.format > 2 {
		return nil, ErrUnhandledFormat
	}
	return msg.value.MarshalBinary()
}

// UnmarshalBinary takes some data, unboxes the t-f-k
// and does some validity checks to make sure it's an understood message reference.
func (msg *Message) UnmarshalBinary(data []byte) error {
	err := msg.value.UnmarshalBinary(data)
	if err != nil {
		msg.broken = true
		return err
	}

	if msg.tipe != TypeMessage {
		msg.broken = true
		return ErrWrongType
	}

	if msg.format > 2 {
		msg.broken = true
		return ErrUnhandledFormat
	}

	// TODO: add bamboo
	if n := len(msg.key); n != 32 {
		if n < 32 {
			msg.broken = true
			return fmt.Errorf("ssb/tfk/message: unexpected key length: %d: %w", n, ErrTooShort)
		}
		msg.key = msg.key[:32]
	}
	return nil
}

// Message retruns the ssb-ref type after a successfull unmarshal.
// It returns a new copy to discourage tampering with the internal values of this reference.
func (msg Message) Message() *refs.MessageRef {
	if msg.broken {
		return nil
	}
	var algo string
	switch msg.format {
	case FormatMessageSHA256:
		algo = refs.RefAlgoMessageSSB1
	case FormatMessageGabbyGrove:
		algo = refs.RefAlgoMessageGabby
	}
	// copy key bytes so that tfk can be re-used?!
	hash := make([]byte, len(msg.key))
	copy(hash, msg.key)
	return &refs.MessageRef{
		Algo: algo,
		Hash: hash,
	}
}
