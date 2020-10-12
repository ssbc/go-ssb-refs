package refs

import (
	"fmt"
	"sort"
)

type ByBranches struct {
	Items []TangledPost

	root   string
	before pointsToMap
}

type pointsToMap map[string][]string

func (byBr *ByBranches) FillLookup() {
	bf := make(pointsToMap, len(byBr.Items))

	for _, m := range byBr.Items {
		branches := m.Branches()

		if branches == nil {
			byBr.root = m.Key().Ref()
			continue
		}

		var refs = make([]string, len(branches))
		for j, br := range branches {
			refs[j] = br.Ref()
		}
		bf[m.Key().Ref()] = refs
	}

	byBr.before = bf
	if byBr.root == "" {
		panic("no root?!")
	}
}

func (bct ByBranches) Len() int {
	return len(bct.Items)
}

func (bct ByBranches) currentIndex(key string) int {
	for idxBr, findBr := range bct.Items {
		if findBr.Key().Ref() == key {
			return idxBr
		}
	}
	return -1
}

func (bct ByBranches) pointsTo(x, y string) bool {
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

func (bct ByBranches) hopsToRoot(key string, hop int) int {
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

func (bct ByBranches) Less(i int, j int) bool {
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

func (bct ByBranches) Swap(i int, j int) {
	bct.Items[i], bct.Items[j] = bct.Items[j], bct.Items[i]
}

type TangledPost interface {
	Key() *MessageRef

	Root() *MessageRef
	Branches() []*MessageRef
}
