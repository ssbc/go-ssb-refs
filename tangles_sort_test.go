package refs_test

import (
	"math/rand"
	"sort"
	"testing"

	"github.com/stretchr/testify/require"

	refs "go.mindeco.de/ssb-refs"
)

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
	tp := make([]refs.TangledPost, len(msgs))
	for i, m := range msgs {
		tp[i] = refs.TangledPost(m)
		// t.Log(i, m.key, m.Key().Ref())
	}

	sorter := refs.ByPrevious{Items: tp}
	sorter.FillLookup()
	sort.Sort(sorter)

	for i, m := range tp {

		fm := m.(fakeMessage)
		// t.Log(i, fm.key, fm.order)

		if fm.order > i+1 {
			t.Errorf("%s has the wrong order", fm.key)
		}
	}

	h := sorter.Heads()
	require.Len(t, h, 1)
	require.EqualValues(t, string(h[0].Hash), "p4", "wrong head")
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
	tp := make([]refs.TangledPost, len(msgs))
	for i, m := range msgs {
		tp[i] = refs.TangledPost(m)
		// t.Log(i, m.key, m.Key().Ref())
	}

	sorter := refs.ByPrevious{Items: tp}
	sorter.FillLookup()
	sort.Sort(sorter)

	for i, m := range tp {

		fm := m.(fakeMessage)

		// a1 and b1 should be at the same level at least
		atLeast := fm.order - 1
		// t.Log(i, fm.key, atLeast)
		if atLeast < i {
			t.Errorf("%s has the wrong order (atLeast:%d i:%d)", fm.key, atLeast, i)
		}
	}

	h := sorter.Heads()
	require.Len(t, h, 2, "wrong count of heads")
	// for >2 this should be sorted by key
	var headKeys []string
	for _, m := range h {
		headKeys = append(headKeys, string(m.Hash))
	}
	if headKeys[0] == "a1" && headKeys[1] == "b1" {
		t.Log("version x")
	} else if headKeys[1] == "a1" && headKeys[0] == "b1" {
		t.Log("version y")
	} else {
		t.Errorf("actual heads: %v", headKeys)
	}
}

func TestBranchMerge(t *testing.T) {
	var msgs = []fakeMessage{
		{key: "p1", order: 1, prev: nil},
		{key: "a1", order: 3, prev: []string{"p1"}},
		{key: "b1", order: 3, prev: []string{"p1"}},
		{key: "p2", order: 4, prev: []string{"a1", "b1"}},
	}

	rand.Shuffle(len(msgs), func(i, j int) {
		msgs[i], msgs[j] = msgs[j], msgs[i]
	})

	// stupid interface wrapping
	tp := make([]refs.TangledPost, len(msgs))
	for i, m := range msgs {
		tp[i] = refs.TangledPost(m)
		// t.Log(i, m.key, m.Key().Ref())
	}

	sorter := refs.ByPrevious{Items: tp}
	sorter.FillLookup()
	sort.Sort(sorter)

	for i, m := range tp {

		fm := m.(fakeMessage)
		// a1 and b1 should be at the same level at least
		atLeast := fm.order - 1
		t.Log(i, fm.key, atLeast)

		if atLeast < i {
			t.Errorf("%d: has the wrong order", i)
		}
	}

	h := sorter.Heads()
	require.Len(t, h, 1)
	require.Equal(t, "p2", string(h[0].Hash))
}

func TestBranchMergeOpen(t *testing.T) {
	// 1:       A1
	//         /|\
	//        / | \
	//       /  |  \
	// 2:   B1  C1  D1
	//       \ /
	// 3:     C2
	var msgs = []fakeMessage{
		{key: "a1", order: 1, prev: nil},
		{key: "b1", order: 2, prev: []string{"a1"}},
		{key: "c1", order: 2, prev: []string{"a1"}},
		{key: "d1", order: 2, prev: []string{"a1"}},
		{key: "c2", order: 3, prev: []string{"b1", "c1"}},
	}

	rand.Shuffle(len(msgs), func(i, j int) {
		msgs[i], msgs[j] = msgs[j], msgs[i]
	})

	// stupid interface wrapping
	tp := make([]refs.TangledPost, len(msgs))
	for i, m := range msgs {
		tp[i] = refs.TangledPost(m)
		// t.Log(i, m.key, m.Key().Ref())
	}

	sorter := refs.ByPrevious{Items: tp}
	sorter.FillLookup()
	sort.Sort(sorter)

	h := sorter.Heads()
	require.Len(t, h, 2)
	var keys []string
	for _, hs := range h {
		keys = append(keys, string(hs.Hash))
	}
	sort.Strings(keys)
	require.Equal(t, []string{"c2", "d1"}, keys)
}

func TestBranchMergeOpenTwo(t *testing.T) {
	// 1:       A1
	//         /|\
	//        / | \
	//       /  |  \
	// 2:   B1  C1  D1
	//       \ /     \
	// 3:     C2     B2
	var msgs = []fakeMessage{
		{key: "a1", order: 1, prev: nil},
		{key: "b1", order: 2, prev: []string{"a1"}},
		{key: "c1", order: 2, prev: []string{"a1"}},
		{key: "d1", order: 2, prev: []string{"a1"}},
		{key: "c2", order: 3, prev: []string{"b1", "c1"}},
		{key: "b2", order: 3, prev: []string{"d1"}},
	}

	rand.Shuffle(len(msgs), func(i, j int) {
		msgs[i], msgs[j] = msgs[j], msgs[i]
	})

	// stupid interface wrapping
	tp := make([]refs.TangledPost, len(msgs))
	for i, m := range msgs {
		tp[i] = refs.TangledPost(m)
		// t.Log(i, m.key, m.Key().Ref())
	}

	sorter := refs.ByPrevious{Items: tp}
	sorter.FillLookup()
	sort.Sort(sorter)

	h := sorter.Heads()
	require.Len(t, h, 2)
	var keys []string
	for _, hs := range h {
		keys = append(keys, string(hs.Hash))
	}
	sort.Strings(keys)
	require.Equal(t, []string{"b2", "c2"}, keys)
}

func TestBranchMergeMulti(t *testing.T) {
	// 1:       A1
	//         /|\
	//        / | \
	//       /  |  \
	// 2:   B1  C1  D1
	//       \ /   /
	// 3:     C2  /
	//         \ /
	// 4:       A2
	var msgs = []fakeMessage{
		{key: "a1", order: 1, prev: nil},
		{key: "b1", order: 2, prev: []string{"a1"}},
		{key: "c1", order: 2, prev: []string{"a1"}},
		{key: "d1", order: 2, prev: []string{"a1"}},
		{key: "c2", order: 3, prev: []string{"b1", "c1"}},
		{key: "a2", order: 4, prev: []string{"c2", "d1"}},
	}

	rand.Shuffle(len(msgs), func(i, j int) {
		msgs[i], msgs[j] = msgs[j], msgs[i]
	})

	// stupid interface wrapping
	tp := make([]refs.TangledPost, len(msgs))
	for i, m := range msgs {
		tp[i] = refs.TangledPost(m)
		// t.Log(i, m.key, m.Key().Ref())
	}

	sorter := refs.ByPrevious{Items: tp}
	sorter.FillLookup()
	sort.Sort(sorter)

	h := sorter.Heads()
	require.Len(t, h, 1)
	require.Equal(t, "a2", string(h[0].Hash))
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
	tp := make([]refs.TangledPost, len(msgs))
	for i, m := range msgs {
		tp[i] = refs.TangledPost(m)

		// t.Log(i, m.key, m.Key().Ref())
	}

	sorter := refs.ByPrevious{Items: tp}
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

func (fm fakeMessage) Key() *refs.MessageRef {
	return &refs.MessageRef{
		Hash: []byte(fm.key),
		Algo: "fake",
	}
}

func (fm fakeMessage) Tangle(_ string) (*refs.MessageRef, refs.MessageRefs) {
	root := &refs.MessageRef{
		Hash: []byte(fm.root),
		Algo: "fake",
	}

	n := len(fm.prev)
	if n == 0 {
		return root, nil
	}

	brs := make(refs.MessageRefs, n)
	for i, b := range fm.prev {
		brs[i] = &refs.MessageRef{
			Hash: []byte(b),
			Algo: "fake",
		}
	}
	return root, brs
}
