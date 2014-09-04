package kubismus

import (
	"sort"
	"time"
)

type metric struct {
	name  string
	value float64
	count int32
	isAvg bool
	data  []float64
}

type getmetric struct {
	name  string
	reply chan []float64
}

var (
	metricChan     = make(chan metric, 1024)
	getMetricsChan = make(chan getmetric, 16)
	freeListM      = make(chan []float64, 64)
	getNamesChan   = make(chan chan []string)
)

func init() {
	go metricservice()
}

func getMetricNames() []string {
	c := make(chan []string)
	getNamesChan <- c
	return <-c
}

func getMetrics(name string) []float64 {
	c := getmetric{name: name, reply: make(chan []float64)}
	getMetricsChan <- c
	return <-c.reply
}

func releaseMetrics(m []float64) {
	// Reuse buffer if there's room.
	select {
	case freeListM <- m:
		// Buffer on free list; nothing more to do.
	default:
		// Free list full, just carry on.
	}
}

func shift(a []float64) {
	for i := 0; i < len(a)-1; i++ {
		a[i] = a[i+1]
	}
}

// Count records a count for a given reading. Values are added within a time interval.
func Count(reading string, value int32) {
	metricChan <- metric{name: reading, value: float64(value), isAvg: false}
}

// Average records a value for a given reading. Values are averaged within a time interval.
func Average(reading string, value float64) {
	metricChan <- metric{name: reading, value: value, isAvg: true}
}

func metricservice() {
	metrics := make(map[string]*metric)
	tck := time.NewTicker(1 * time.Second)
	for {
		select {
		case m := <-metricChan:
			if m.name != "" {
				v, ok := metrics[m.name]
				if !ok {
					v = &metric{name: m.name, isAvg: m.isAvg, count: 0, value: 0.0, data: make([]float64, 960)}
					metrics[m.name] = v
				}
				if v.isAvg == true {
					v.count++
				}
				v.value += m.value
			}
		case <-tck.C:
			for _, y := range metrics {
				ans := y.value
				if y.count > 0 {
					ans /= float64(y.count)
				}
				shift(y.data)
				y.data[len(y.data)-1] = ans
				y.count = 0
				y.value = 0.0
			}
		case gm := <-getMetricsChan:
			var r []float64
			// Grab a buffer if available; allocate if not.
			select {
			case r = <-freeListM:
				// Got one; nothing more to do but slice it.
			default:
				// None free, so allocate a new one.
				r = make([]float64, 960)
			}
			v, ok := metrics[gm.name]
			if !ok {
				gm.reply <- nil
			} else {
				r = r[0:len(v.data)]
				copy(r, v.data)
				gm.reply <- r
			}
		case gn := <-getNamesChan:
			s := make([]string, 0, 8)
			for x, _ := range metrics {
				s = append(s, x)
			}
			sort.Strings(s)
			gn <- s
		}
	}

	/*
		var r int64
		var bytes int64
		var data Status
		r = 0
		bytes = 0
		tck := time.NewTicker(1 * time.Second)
		for {
			select {
			case b := <-x:
				r += 1
				bytes += b
			case <-tck.C:
				shift(data.Requests[:])
				data.Requests[0] = r
				shift(data.Bytes[:])
				data.Bytes[0] = bytes
				if r != 0 {
					log.Print(r, " req/s, ", bytes, " bytes/s")
					r = 0
					bytes = 0
				}
			case x := <-reqStatusChan:
				b, err := json.Marshal(data)
				if err != nil {
					log.Panic("cannot format json")
				}
				x.Reply <- b
			}
		}
	*/
}
