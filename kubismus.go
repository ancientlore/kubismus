/*
	kubismus makes it simple to embed a status page in your web service. Using D3 and Cubism, events
	are stored and then rendered on a dynamic display.
*/
package kubismus

import (
	"encoding/json"
	"github.com/ancientlore/kubismus/static"
	"html/template"
	"log"
	"net/http"
	"path"
	"strconv"
	"strings"
)

const (
	DefaultPath = "/kubismus/"
)

var (
	mux  *http.ServeMux     // ServeMux to multiplex request paths
	tmpl *template.Template // rendering templates
	pg   page               // page data for template
)

type page struct {
	Title    string   // Monitor page title
	Image    string   // Optional monitor page image
	Readings []string // List of the names for each reading that gets a graph
}

// init sets up the templates and http handlers
func init() {
	b := static.Lookup("/tpl/index.html")
	if b == nil {
		log.Fatal("Unable to find template")
	}
	tmpl = template.Must(template.New("kubismus").Parse(string(b)))

	pg.Title = "Kubismus"
	pg.Image = "web/kubismus36.png"
	pg.Readings = make([]string, 0)

	mux = http.NewServeMux()
	mux.Handle("/", http.HandlerFunc(index))
	mux.Handle("/web/", http.HandlerFunc(static.ServeHTTP))
	mux.Handle("/json/notes", http.HandlerFunc(jsonNotes))
	mux.Handle("/json/metrics", http.HandlerFunc(jsonNames))
	mux.Handle("/json/metrics/", http.HandlerFunc(jsonMetrics))
}

// index handles the template rendering
func index(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" || r.URL.Path == "/index" {
		pg.Readings = getMetricNames()
		tmpl.ExecuteTemplate(w, "kubismus", pg)
	} else {
		http.NotFound(w, r)
	}
}

func jsonNotes(w http.ResponseWriter, r *http.Request) {
	notes := getNotes()
	defer releaseNotes(notes)
	e := json.NewEncoder(w)
	if e == nil {
		http.Error(w, "Unable to create json encoder", 500)
		return
	}
	err := e.Encode(notes)
	if err != nil {
		http.Error(w, err.Error(), 500)
	}
}

func jsonNames(w http.ResponseWriter, r *http.Request) {
	n := getMetricNames()
	e := json.NewEncoder(w)
	if e == nil {
		http.Error(w, "Unable to create json encoder", 500)
		return
	}
	err := e.Encode(n)
	if err != nil {
		http.Error(w, err.Error(), 500)
	}
}

func jsonMetrics(w http.ResponseWriter, r *http.Request) {
	name := path.Base(r.URL.Path)
	// assume step = 1000ms for this
	start, err1 := strconv.ParseInt(r.URL.Query().Get("start"), 10, 64)
	stop, err2 := strconv.ParseInt(r.URL.Query().Get("stop"), 10, 64)
	if err1 != nil || err2 != nil {
		http.Error(w, "Invalid start or stop", 500)
		return
	}
	count := int((stop - start) / 1000)
	m := getMetrics(name)
	if m == nil {
		http.Error(w, "No metric named "+name, 500)
		return
	}
	defer releaseMetrics(m)
	if count < 0 || count > len(m) {
		http.Error(w, "Invalid start or stop range", 500)
		return
	}
	mout := m[len(m)-count:]
	//log.Printf("Count %d len %d data %v", count, len(m), mout)
	e := json.NewEncoder(w)
	if e == nil {
		http.Error(w, "Unable to create json encoder", 500)
		return
	}
	err := e.Encode(mout)
	if err != nil {
		http.Error(w, err.Error(), 500)
	}
}

// Setup sets the basic parameters for the graph page.
func Setup(title, image string) {
	pg.Title = title
	pg.Image = image
}

// ServeHTTP servers HTTP requests for kubismus.
func ServeHTTP(w http.ResponseWriter, r *http.Request) {
	mux.ServeHTTP(w, r)
}

// HandleHTTP registers an HTTP handler for kubismus on the default path.
// It is still necessary to invoke http.Serve(), typically in a go statement.
func HandleHTTP() {
	http.Handle(DefaultPath, http.StripPrefix(strings.TrimSuffix(DefaultPath, "/"), http.HandlerFunc(ServeHTTP)))
}
