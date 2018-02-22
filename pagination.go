package zero

// Pagination used to describe pagination input
type Pagination struct {
	Offset       int
	ObjectID     int64
	Count        int
	Skip         int
	AllCount     int
	NextObjectID int64
}

// PaginationWrap used to wrap result of pagination
type PaginationWrap struct {
	Items    interface{} `json:"items"`
	NextFrom string      `json:"next_from,omitempty"`
}

// Wrap will return object of PaginationWrap
func (p *Pagination) Wrap(items interface{}) *PaginationWrap {
	return &PaginationWrap{
		Items:    items,
		NextFrom: p.GetNextFrom(),
	}
}

// GetNextFrom calculates next_from using pagination params
func (p *Pagination) GetNextFrom() string {
	if p.AllCount > 0 && p.Offset+p.Count+p.Skip < p.AllCount {
		return J(p.Offset+p.Count+p.Skip, ":", p.NextObjectID)
	}
	return ""
}

// Slice slices int64 list using pagination
func (p *Pagination) Slice(items []int64) []int64 {
	items = items[p.Offset:]
	for k, v := range items {
		if v == p.ObjectID {
			p.Skip = k
		}
	}
	toNum := p.Skip + p.Count
	if toNum > len(items) {
		toNum = len(items)
	}
	if toNum == 0 {
		return items
	}
	p.NextObjectID = items[toNum-1]
	return items[p.Skip:toNum]
}

// From return from value for redis
func (p *Pagination) From() int64 {
	return int64(p.Offset)
}

// To return to value for redis
func (p *Pagination) To() int64 {
	return int64(p.Offset + p.Count)
}
