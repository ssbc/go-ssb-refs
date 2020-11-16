package refs

import (
	"testing"
)

func TestBranchHelperHops(t *testing.T) {
	var msgs = []fakeMessage{
		{key: "p1", order: 1, prev: nil},
		{key: "p2", order: 2, prev: []string{"p1"}},
		{key: "p3", order: 3, prev: []string{"p2"}},
		{key: "p4", order: 4, prev: []string{"p3"}},
		{key: "p5", order: 5, prev: []string{"p4"}},
	}

	// stupid interface wrapping
	tp := make([]TangledPost, len(msgs))
	for i, m := range msgs {
		tp[i] = TangledPost(m)

		// t.Log(i, m.key, m.Key().Ref())
	}

	sorter := ByPrevious{Items: tp}
	sorter.FillLookup()

	for i := len(msgs) - 1; i >= 0; i-- {
		if h := sorter.hopsToRoot(msgs[i].Key().Ref(), 0); h != i {
			t.Error("wrong p1", h)
		}
	}
}

type fakeMessage struct {
	key string

	root string // same for all
	prev []string

	order int // test index
}

func (fm fakeMessage) Key() *MessageRef {
	return &MessageRef{
		Hash: []byte(fm.key),
		Algo: "fake",
	}
}

func (fm fakeMessage) Tangle(_ string) (*MessageRef, MessageRefs) {
	root := &MessageRef{
		Hash: []byte(fm.root),
		Algo: "fake",
	}

	n := len(fm.prev)
	if n == 0 {
		return root, nil
	}

	brs := make(MessageRefs, n)
	for i, b := range fm.prev {
		brs[i] = &MessageRef{
			Hash: []byte(b),
			Algo: "fake",
		}
	}
	return root, brs
}
