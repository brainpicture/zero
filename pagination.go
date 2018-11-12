package zero

import (
	"strings"
)

// Pagination used to describe pagination input
type Pagination struct {
	offset       int64
	from         string
	nextFrom     string
	ObjectID     int64
	Count        int
	Skip         int
	NewOffset    int64 // used to override offset and count
	AllCount     int
	NextObjectID int64
	Reverse      bool
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

// WrapCustom will return object of PaginationWrap with some additional fields
func (p *Pagination) WrapCustom(items interface{}, extra H) H {
	resp := H{
		"items": items,
	}
	nextFrom := p.GetNextFrom()
	if nextFrom != "" {
		resp["next_from"] = nextFrom
	}
	for k, v := range extra {
		resp[k] = v
	}
	return resp
}

// SetNextFrom sets next_from pagination value based in interface values passed
func (p *Pagination) SetNextFrom(items ...interface{}) {
	nextFrom := []string{}
	for _, v := range items {
		nextFrom = append(nextFrom, J(v))
	}
	p.nextFrom = strings.Join(nextFrom, ";")
}

// SetOffset will set new offset value
func (p *Pagination) SetOffset(val int64) {
	p.offset = val
}

// SetOffsetInt will set new offset value
func (p *Pagination) SetOffsetInt(val int) {
	p.offset = int64(val)
}

// GetNextFrom calculates next_from using pagination params
func (p *Pagination) GetNextFrom() string {
	if p.nextFrom != "" {
		return p.nextFrom
	}
	if p.NewOffset != 0 { // new offset is used
		if p.NewOffset == -1 {
			return ""
		}
		if p.NextObjectID != 0 {
			return J(p.NewOffset, ":", p.NextObjectID)
		} else {
			return J(p.NewOffset)
		}
	}
	if p.AllCount > 0 && p.offset+int64(p.Count+p.Skip) < int64(p.AllCount) {
		return J(p.offset+int64(p.Count+p.Skip), ":", p.NextObjectID)
	}
	return ""
}

// Slice slices int64 list using pagination
func (p *Pagination) Slice(items []int64) []int64 {
	items = items[p.offset:]
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

// Offset return offset value for as int64
func (p *Pagination) Offset() int64 {
	return p.offset
}

// OffsetInt return offset value for as integer
func (p *Pagination) OffsetInt() int {
	return int(p.offset)
}

// From return from value
func (p *Pagination) From() string {
	return p.from
}

// To return to value for redis
func (p *Pagination) To() int64 {
	return int64(p.offset + int64(p.Count))
}

// GetCursor returns cursor value for redis
func (p *Pagination) GetCursor() uint64 {
	return uint64(p.offset)
}

// SetCursor sets cursor value from redis
func (p *Pagination) SetCursor(cursor uint64) {
	if cursor == 0 {
		p.NewOffset = -1 // nothing more
	} else {
		p.NewOffset = int64(cursor)
	}
}
