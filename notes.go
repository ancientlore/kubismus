package kubismus

import (
	"sort"
)

// note defines the information stored for a note
type note struct {
	Name  string `json:"key"`
	Value string `json:"value"`
}

// sortNote specifies how to sort a slice of notes
type sortNote []note

func (a sortNote) Len() int           { return len(a) }
func (a sortNote) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a sortNote) Less(i, j int) bool { return a[i].Name < a[j].Name }

var (
	noteChan     = make(chan note, 16)
	getNotesChan = make(chan chan []note)
	freeList     = make(chan []note, 16)
)

// init initializes the notes system
func init() {
	go noteService()
}

// Note logs a specific value to show in a table.
func Note(name, value string) {
	noteChan <- note{Name: name, Value: value}
}

// GetNotes returns the currently defined notes.
func GetNotes() []note {
	c := make(chan []note)
	getNotesChan <- c
	return <-c
}

// ReleaseNotes returns the slice to the leaky buffer, if possible.
func ReleaseNotes(n []note) {
	// Reuse buffer if there's room.
	select {
	case freeList <- n:
		// Buffer on free list; nothing more to do.
	default:
		// Free list full, just carry on.
	}
}

// noteservice process note-related requests
func noteService() {
	notes := make(map[string]string)
	for {
		select {
		case n := <-noteChan:
			if n.Name != "" {
				notes[n.Name] = n.Value
			}
		case gn := <-getNotesChan:
			var r []note
			// Grab a buffer if available; allocate if not.
			select {
			case r = <-freeList:
				// Got one; nothing more to do but slice it.
				r = r[0:0]
			default:
				// None free, so allocate a new one.
				r = make([]note, 0, 32)
			}
			for k, v := range notes {
				r = append(r, note{Name: k, Value: v})
			}
			sort.Sort(sortNote(r))
			gn <- r
		}
	}
}
