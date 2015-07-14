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
	"strconv"
	"strings"
)

// github.com/ancientlore/binder is used to package the web files into the executable.
//go:generate binder -package static -o static/files.go web/* tpl/*

const (
	DefaultPath = "/kubismus/" // The default path on the URL to get to the Kubismus display
)

var (
	mux  *http.ServeMux     // ServeMux to multiplex request paths
	tmpl *template.Template // rendering templates
	pg   page               // page data for template
)

// page represents data that is rendered in the HTML
type page struct {
	Title    string      // Monitor page title
	Image    string      // Optional monitor page image
	Readings []metricDef // List of the names for each reading that gets a graph
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
	pg.Readings = make([]metricDef, 0)

	mux = http.NewServeMux()
	mux.Handle("/", http.HandlerFunc(index))
	mux.Handle("/web/", http.HandlerFunc(static.ServeHTTP))
	mux.Handle("/json/notes", http.HandlerFunc(jsonNotes))
	mux.Handle("/json/metrics/list", http.HandlerFunc(jsonDefs))
	mux.Handle("/json/metrics", http.HandlerFunc(jsonMetrics))
}

// index handles the template rendering
func index(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" || r.URL.Path == "/index" {
		pg.Readings = getMetricDefs()
		tmpl.ExecuteTemplate(w, "kubismus", pg)
	} else {
		http.NotFound(w, r)
	}
}

// jsonNotes handles returning note data in JSON format
func jsonNotes(w http.ResponseWriter, r *http.Request) {
	notes := GetNotes()
	defer ReleaseNotes(notes)
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

// jsonDefs handles returning the definitions of the metrics in JSON format
func jsonDefs(w http.ResponseWriter, r *http.Request) {
	n := getMetricDefs()
	defer releaseMetricDefs(n)
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

// jsonMetrics handles returning metric values in JSON format
func jsonMetrics(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	var op Op
	switch r.URL.Query().Get("op") {
	case "count":
		op = COUNT
	case "average":
		op = AVERAGE
	case "sum":
		op = SUM
	default:
		http.Error(w, "Invalid type, must be \"count\", \"average\", or \"sum\"", 500)
		return
	}
	// assume step = 1000ms for this
	start, err1 := strconv.ParseInt(r.URL.Query().Get("start"), 10, 64)
	stop, err2 := strconv.ParseInt(r.URL.Query().Get("stop"), 10, 64)
	if err1 != nil || err2 != nil {
		http.Error(w, "Invalid start or stop, must be time value in milliseconds", 500)
		return
	}
	count := int((stop - start) / 1000)
	m := GetMetrics(name, op)
	if m == nil {
		http.Error(w, "No metric named \""+name+"\" found", 500)
		return
	}
	defer ReleaseMetrics(m)
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

// HttpRequestMetric returns a handler that logs a metric for the incoming content length
// after invoking the handler h
func HttpRequestMetric(reading string, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(w, r)
		Metric(reading, 1, float64(r.ContentLength))
	})
}

// HttpResponseMetric returns a handler that logs a metric for the response content length
// after invoking the handler h
func HttpResponseMetric(reading string, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(w, r)
		l, e := strconv.Atoi(w.Header().Get("Content-Length"))
		if e != nil {
			l = 0
		}
		Metric(reading, 1, float64(l))
	})
}
