package zero

import (
	"strings"
)

// Pagination used to describe pagination input
type Pagination struct {
	srv          *Server
	offset       int64
	from         string
	nextFrom     string
	ObjectID     int64
	Count        int
	Skip         int
	NextOffset   int64 // used to override offset and count
	PrevOffset   int64 // used to automaticly set PrevFrom using int
	AllCount     int
	NextObjectID int64
	Reverse      bool
}

// PaginationWrap used to wrap result of pagination
type PaginationWrap struct {
	Items    interface{} `json:"items"`
	NextFrom string      `json:"next_from,omitempty"`
	PrevFrom string      `json:"prev_from,omitempty"`
}

// Wrap will return object of PaginationWrap
func (p *Pagination) Wrap(items interface{}) *PaginationWrap {
	pagWrap := PaginationWrap{
		Items: items,
	}
	if p.Reverse {
		pagWrap.PrevFrom = p.GetNextFrom()
		pagWrap.NextFrom = p.GetPrevFrom()
	} else {
		pagWrap.NextFrom = p.GetNextFrom()
		pagWrap.PrevFrom = p.GetPrevFrom()
	}
	return &pagWrap
}

// CountMax will send error if count bigger than limit
func (p *Pagination) CountMax(limit int) {
	if p.Count > limit {
		p.srv.Err("count_field", "count field is invalid, too big")
	}
}

// WrapCustom will return object of PaginationWrap with some additional fields
func (p *Pagination) WrapCustom(items interface{}, extra H) H {
	resp := H{
		"items": items,
	}

	if p.Reverse {
		prevFrom := p.GetNextFrom()
		if prevFrom != "" {
			resp["prev_from"] = prevFrom
		}
		nextFrom := p.GetPrevFrom()
		if nextFrom != "" {
			resp["next_from"] = nextFrom
		}
	} else {
		nextFrom := p.GetNextFrom()
		if nextFrom != "" {
			resp["next_from"] = nextFrom
		}
		prevFrom := p.GetPrevFrom()
		if prevFrom != "" {
			resp["prev_from"] = prevFrom
		}
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
	if p.NextOffset != 0 { // new offset is used
		if p.NextOffset == -1 {
			return ""
		}
		if p.NextObjectID != 0 {
			return J(p.NextOffset, ":", p.NextObjectID)
		} else {
			return J(p.NextOffset)
		}
	}
	if p.AllCount > 0 && p.offset+int64(p.Count+p.Skip) < int64(p.AllCount) {
		return J(p.offset+int64(p.Count+p.Skip), ":", p.NextObjectID)
	}
	return ""
}

// GetPrevFrom will return ultimately prev_from build using PrevOffset value
func (p *Pagination) GetPrevFrom() string {
	if p.PrevOffset != 0 { // new offset is used
		return J(p.PrevOffset)
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
		p.NextOffset = -1 // nothing more
	} else {
		p.NextOffset = int64(cursor)
	}
}
