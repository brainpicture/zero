package zero

import (
	"time"
)

// StatCounter is an event object
type StatCounter struct {
	Name       string
	Count      int64
	Elapsed    int64
	MaxElapsed int64
}

// Stat main type for stat watch
type Stat struct {
	counter     map[string]StatCounter
	counterChan chan StatCounter
}

// Init allow to set stats update time
func (s *Stat) Init(duration time.Duration, cb func(string, StatCounter)) {
	s.counter = map[string]StatCounter{}
	s.counterChan = make(chan StatCounter, 100)
	go func() {
		ticker := time.NewTicker(duration)
		for {
			select {
			case e := <-s.counterChan:
				counter, ok := s.counter[e.Name]
				if !ok {
					counter = StatCounter{Name: e.Name}
				}
				counter.Count += e.Count
				if e.Elapsed != 0 {
					counter.Elapsed += e.Elapsed
					if e.Elapsed > counter.MaxElapsed {
						counter.MaxElapsed = e.Elapsed
					}
				}
				s.counter[e.Name] = counter
			case <-ticker.C:
				for k, v := range s.counter {
					cb(k, v)
				}
				// Clean all counts
				s.counter = map[string]StatCounter{}
			}
		}
	}()
}

// Inc increment
func (s *Stat) Inc(name string) {
	select {
	case s.counterChan <- StatCounter{
		Name:  name,
		Count: 1,
	}:
	default:
	}
}

// Time increment
func (s *Stat) Time(name string, time int64) {
	select {
	case s.counterChan <- StatCounter{
		Name:    name,
		Count:   1,
		Elapsed: time,
	}:
	default:
	}
}

// StatDay return number of the day as int
func StatDay() int64 {
	return Now() / 86400
}

// StatHour return number of hour as int
func StatHour() int64 {
	return Now() / 3600
}

// StatMonth return number of month
func StatMonth() int64 {
	now := time.Now()
	return int64(now.Year()*100) + int64(now.Month())
}
