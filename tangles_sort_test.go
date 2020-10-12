package refs

import (
	"math/rand"
	"sort"
	"testing"
)

func TestBranchHelperHops(t *testing.T) {
	var msgs = []fakeMessage{
		{key: "p1", order: 1, branches: nil},
		{key: "p2", order: 2, branches: []string{"p1"}},
		{key: "p3", order: 3, branches: []string{"p2"}},
		{key: "p4", order: 4, branches: []string{"p3"}},
		{key: "p5", order: 5, branches: []string{"p4"}},
	}

	// stupid interface wrapping
	tp := make([]TangledPost, len(msgs))
	for i, m := range msgs {
		tp[i] = TangledPost(m)

		t.Log(i, m.key, m.Key().Ref())
	}

	sorter := ByBranches{Items: tp}
	sorter.FillLookup()

	for i := len(msgs) - 1; i >= 0; i-- {
		if h := sorter.hopsToRoot(msgs[i].Key().Ref(), 0); h != i {
			t.Error("wrong p1", h)
		}

	}
}

func TestBranchCausalitySimple(t *testing.T) {
	var msgs = []fakeMessage{
		{key: "p1", order: 1, branches: nil},
		{key: "p2", order: 3, branches: []string{"b1"}},
		{key: "p3", order: 6, branches: []string{"b2"}},

		{key: "b1", order: 2, branches: []string{"p1"}},
		{key: "b2", order: 5, branches: []string{"p2", "s1"}},
		{key: "b3", order: 8, branches: []string{"p3", "s2"}},

		{key: "s1", order: 4, branches: []string{"p1"}},
		{key: "s2", order: 7, branches: []string{"b2"}},
		// {key: "s3", order: 9, branches: []string{"p3", "s2"}},
	}

	rand.Shuffle(len(msgs), func(i, j int) {
		msgs[i], msgs[j] = msgs[j], msgs[i]
	})

	// stupid interface wrapping
	tp := make([]TangledPost, len(msgs))
	for i, m := range msgs {
		tp[i] = TangledPost(m)

		t.Log(i, m.key, m.Key().Ref())
	}

	sorter := ByBranches{Items: tp}
	sorter.FillLookup()
	sort.Sort(sorter)

	for i, m := range tp {

		fm := m.(fakeMessage)
		t.Log(fm.key, fm.order)
		if fm.order != i+1 {
			t.Error(fm.key, "not sorted")
			// TODO: there is no tiebreak on the numbers of replies but it's nearly correct
			// i _think_ new replies should go lower to start off new branches
			// and not disrupt existing flow too much but need to make more cases to show this
		}
	}

}

type fakeMessage struct {
	key string

	root     string // same for all
	branches []string

	order int // test index
}

func (fm fakeMessage) Key() *MessageRef {
	return &MessageRef{
		Hash: []byte(fm.key),
		Algo: "fake",
	}
}

func (fm fakeMessage) Root() *MessageRef {
	return &MessageRef{
		Hash: []byte(fm.root),
		Algo: "fake",
	}
}

func (fm fakeMessage) Branches() []*MessageRef {
	n := len(fm.branches)
	if n == 0 {
		return nil
	}

	brs := make([]*MessageRef, n)
	for i, b := range fm.branches {
		brs[i] = &MessageRef{
			Hash: []byte(b),
			Algo: "fake",
		}
	}
	return brs
}
