package refs

import (
	"math/rand"
	"sort"
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

func TestBranchSequential(t *testing.T) {
	var msgs = []fakeMessage{
		{key: "p1", order: 1, prev: nil},
		{key: "p2", order: 2, prev: []string{"p1"}},
		{key: "p3", order: 3, prev: []string{"p2"}},
		{key: "p4", order: 4, prev: []string{"p3"}},
	}

	rand.Shuffle(len(msgs), func(i, j int) {
		msgs[i], msgs[j] = msgs[j], msgs[i]
	})

	// stupid interface wrapping
	tp := make([]TangledPost, len(msgs))
	for i, m := range msgs {
		tp[i] = TangledPost(m)
		// t.Log(i, m.key, m.Key().Ref())
	}

	sorter := ByPrevious{Items: tp}
	sorter.FillLookup()
	sort.Sort(sorter)

	for i, m := range tp {

		fm := m.(fakeMessage)
		// t.Log(i, fm.key, fm.order)

		if fm.order > i+1 {
			t.Errorf("%s has the wrong order", fm.key)
		}
	}
}

func TestBranchConcurrent(t *testing.T) {
	var msgs = []fakeMessage{
		{key: "p1", order: 1, prev: nil},
		{key: "a1", order: 3, prev: []string{"p1"}},
		{key: "b1", order: 3, prev: []string{"p1"}},
	}

	rand.Shuffle(len(msgs), func(i, j int) {
		msgs[i], msgs[j] = msgs[j], msgs[i]
	})

	// stupid interface wrapping
	tp := make([]TangledPost, len(msgs))
	for i, m := range msgs {
		tp[i] = TangledPost(m)
		// t.Log(i, m.key, m.Key().Ref())
	}

	sorter := ByPrevious{Items: tp}
	sorter.FillLookup()
	sort.Sort(sorter)

	for i, m := range tp {

		fm := m.(fakeMessage)
		atLeast := fm.order - 1
		// t.Log(i, fm.key, atLeast)
		if atLeast < i {
			t.Errorf("%s has the wrong order (atLeast:%d i:%d)", fm.key, atLeast, i)
		}
	}
}

func TestBranchMerge(t *testing.T) {
	var msgs = []fakeMessage{
		{key: "p1", order: 1, prev: nil},
		{key: "a1", order: 2, prev: []string{"p1"}},
		{key: "b1", order: 3, prev: []string{"p1"}},
		{key: "p2", order: 4, prev: []string{"a1", "b1"}},
	}

	rand.Shuffle(len(msgs), func(i, j int) {
		msgs[i], msgs[j] = msgs[j], msgs[i]
	})

	// stupid interface wrapping
	tp := make([]TangledPost, len(msgs))
	for i, m := range msgs {
		tp[i] = TangledPost(m)
		// t.Log(i, m.key, m.Key().Ref())
	}

	sorter := ByPrevious{Items: tp}
	sorter.FillLookup()
	sort.Sort(sorter)

	for i, m := range tp {

		fm := m.(fakeMessage)
		// t.Log(i, fm.key, fm.order)

		if fm.order > i+1 {
			t.Errorf("%s has the wrong order", fm.key)
		}
	}
}

func XTestBranchCausalityLong(t *testing.T) {
	var msgs = []fakeMessage{
		{key: "p1", order: 1, prev: nil},
		{key: "p2", order: 3, prev: []string{"b1"}},
		{key: "p3", order: 6, prev: []string{"b2"}},

		{key: "b1", order: 2, prev: []string{"p1"}},
		{key: "b2", order: 5, prev: []string{"p2", "s1"}},
		{key: "b3", order: 8, prev: []string{"p3", "s2"}},

		{key: "s1", order: 4, prev: []string{"p1"}},
		{key: "s2", order: 7, prev: []string{"b2"}},
		// {key: "s3", order: 9, prev: []string{"p3", "s2"}},
	}

	rand.Shuffle(len(msgs), func(i, j int) {
		msgs[i], msgs[j] = msgs[j], msgs[i]
	})

	// stupid interface wrapping
	tp := make([]TangledPost, len(msgs))
	for i, m := range msgs {
		tp[i] = TangledPost(m)

		// t.Log(i, m.key, m.Key().Ref())
	}

	sorter := ByPrevious{Items: tp}
	sorter.FillLookup()
	sort.Sort(sorter)

	for i, m := range tp {

		fm := m.(fakeMessage)
		// t.Log(i, fm.key, fm.order)
		if fm.order != i+1 {
			t.Error(fm.key, "not sorted")
			// TODO: there is no tiebreak on the numbers of replies but it's nearly correct
			// i _think_ new replies should go lower to start off new prev
			// and not disrupt existing flow too much but need to make more cases to show this
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
