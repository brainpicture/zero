package zero

import "sort"

type SortInt64 []int64

func (p SortInt64) Len() int           { return len(p) }
func (p SortInt64) Less(i, j int) bool { return p[i] < p[j] }
func (p SortInt64) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

func (p SortInt64) Sort() {
	sort.Sort(p)
}

func (p SortInt64) RevSort() {
	sort.Sort(sort.Reverse(p))
}
