package kubismus

import (
	"sort"
	"time"
)

type kind int

const (
	mCount kind = 1 << iota
	mAverage
	mSum
)

type metric struct {
	name  string
	count int32
	value float64
	cData []float64
	vData []float64
}

type getmetric struct {
	name  string
	mtype kind
	reply chan []float64
}

type metricDef struct {
	Name string
	Type string
}

type sortMetricDef []metricDef

func (a sortMetricDef) Len() int      { return len(a) }
func (a sortMetricDef) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a sortMetricDef) Less(i, j int) bool {
	return a[i].Name < a[j].Name || (a[i].Name == a[j].Name && a[i].Type < a[j].Type)
}

var (
	metricChan     = make(chan metric, 1024)
	getMetricsChan = make(chan getmetric, 16)
	freeListM      = make(chan []float64, 64)
	getNamesChan   = make(chan chan []metricDef)
)

func init() {
	go metricservice()
}

func getMetricNames() []metricDef {
	c := make(chan []metricDef)
	getNamesChan <- c
	return <-c
}

func getMetrics(name string, mtype kind) []float64 {
	c := getmetric{name: name, mtype: mtype, reply: make(chan []float64)}
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

// Metric records a count and total value for a given reading. count should be 1 unless you are providing
// summed data for multiple events as the value. For instance, you can send the total bytes read for 3 files
// at one time.
func Metric(reading string, count int32, value float64) {
	metricChan <- metric{name: reading, count: count, value: value}
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
					v = &metric{name: m.name, count: 0, value: 0.0, cData: make([]float64, 960), vData: make([]float64, 960)}
					metrics[m.name] = v
				}
				v.count += m.count
				v.value += m.value
			}
		case <-tck.C:
			for _, y := range metrics {
				shift(y.cData)
				y.cData[len(y.cData)-1] = float64(y.count)
				shift(y.vData)
				y.vData[len(y.vData)-1] = y.value
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
				r = r[0:len(v.cData)]
				switch gm.mtype {
				case mCount:
					copy(r, v.cData)
				case mAverage:
					for i, c := range v.cData {
						if c == 0.0 || c < 0.000000001 {
							r[i] = v.vData[i]
						} else {
							r[i] = v.vData[i] / c
						}
					}
				case mSum:
					copy(r, v.vData)
				}
				gm.reply <- r
			}
		case gn := <-getNamesChan:
			s := make([]metricDef, 0, 8)
			for x, _ := range metrics {
				s = append(s, metricDef{x, "count"})
				s = append(s, metricDef{x, "average"})
				s = append(s, metricDef{x, "sum"})
			}
			sort.Sort(sortMetricDef(s))
			gn <- s
		}
	}
}
