package kubismus

import (
	"sort"
)

type note struct {
	Name  string `json:"key"`
	Value string `json:"value"`
}

type sortNote []note

func (a sortNote) Len() int           { return len(a) }
func (a sortNote) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a sortNote) Less(i, j int) bool { return a[i].Name < a[j].Name }

var (
	noteChan     = make(chan note)
	getNotesChan = make(chan chan []note)
)

func init() {
	go noteservice()
}

// Note logs a specific value to show in a table.
func Note(name, value string) {
	noteChan <- note{Name: name, Value: value}
}

func getNotes() []note {
	c := make(chan []note)
	getNotesChan <- c
	return <-c
}

func noteservice() {
	notes := make(map[string]string)
	for {
		select {
		case n := <-noteChan:
			if n.Name != "" {
				notes[n.Name] = n.Value
			}
		case gn := <-getNotesChan:
			r := make([]note, 0, 32)
			for k, v := range notes {
				r = append(r, note{Name: k, Value: v})
			}
			sort.Sort(sortNote(r))
			gn <- r
		}
	}
}

// LogCount records a count for a given reading. Values are added within a time interval.
func LogCount(reading string, value int32) {

}

// LogAverage records a value for a given reading. Values are averaged within a time interval.
func LogAverage(reading string, value float64) {

}
