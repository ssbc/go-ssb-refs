package refs

import (
	"fmt"
	"sort"
)

type ByPrevious struct {
	TangleName string

	Items []TangledPost

	root   string
	before pointsToMap
}

type pointsToMap map[string][]string

func (byBr *ByPrevious) FillLookup() {
	bf := make(pointsToMap, len(byBr.Items))

	for _, m := range byBr.Items {
		_, prev := m.Tangle(byBr.TangleName)

		if prev == nil {
			byBr.root = m.Key().Ref()
			continue
		}

		var refs = make([]string, len(prev))
		for j, br := range prev {
			refs[j] = br.Ref()
		}
		bf[m.Key().Ref()] = refs
	}

	byBr.before = bf
	if byBr.root == "" {
		panic("no root?!")
	}
}

func (bct ByPrevious) Len() int {
	return len(bct.Items)
}

func (bct ByPrevious) currentIndex(key string) int {
	for idxBr, findBr := range bct.Items {
		if findBr.Key().Ref() == key {
			return idxBr
		}
	}
	return -1
}

func (bct ByPrevious) pointsTo(x, y string) bool {
	pointsTo, has := bct.before[x]
	if !has {
		return false
	}

	for _, candidate := range pointsTo {
		if candidate == y {
			return true
		}
		if bct.pointsTo(candidate, y) {
			return true
		}
	}
	return false
}

func (bct ByPrevious) hopsToRoot(key string, hop int) int {
	if key == bct.root {
		return hop
	}

	pointsTo, ok := bct.before[key]
	if !ok {
		panic("untangled message which isnt root:" + key)
	}

	fmt.Printf("toRoot(%s) h:%d (ptrs:%d)\n", key, hop, len(pointsTo))

	var found []int // collect all paths for tie-breaking
	for _, candidate := range pointsTo {
		if candidate == bct.root {
			found = append(found, hop+1)
			continue
		}

		fmt.Println("\tlooking for:", candidate)
		if h := bct.hopsToRoot(candidate, hop+1); h > 0 {
			// TODO: fill up cache of these results
			found = append(found, h)
		}
	}

	if len(found) < 1 {
		panic("not pointing to root?")
	}
	sort.Ints(found)
	return found[len(found)-1]
}

func (bct ByPrevious) Less(i int, j int) bool {
	msgI, msgJ := bct.Items[i], bct.Items[j]
	keyI, keyJ := msgI.Key().Ref(), msgJ.Key().Ref()

	if bct.pointsTo(keyI, keyJ) {
		return false
	}

	hops_i, hops_j := bct.hopsToRoot(keyI, 0), bct.hopsToRoot(keyJ, 0)
	if hops_i < hops_j {
		return true
	}
	fmt.Println(keyI, hops_i)
	fmt.Println(keyJ, hops_j)

	return false
}

func (bct ByPrevious) Swap(i int, j int) {
	bct.Items[i], bct.Items[j] = bct.Items[j], bct.Items[i]
}

type TangledPost interface {
	Key() *MessageRef

	Tangle(name string) (root *MessageRef, prev MessageRefs)
}
