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
	mux.Handle("/json/", http.HandlerFunc(handleJson))
	mux.Handle("/json/notes", http.HandlerFunc(jsonNotes))
}

// index handles the template rendering
func index(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" || r.URL.Path == "/index" {
		tmpl.ExecuteTemplate(w, "kubismus", pg)
	} else {
		http.NotFound(w, r)
	}
}

func jsonNotes(w http.ResponseWriter, r *http.Request) {
	notes := getNotes()
	b, err := json.Marshal(notes)
	if err != nil {
		http.Error(w, err.Error(), 500)
	} else {
		w.Write(b)
	}
}

// handleJson handles producing the metrics as JSON needed by Cubism
func handleJson(w http.ResponseWriter, r *http.Request) {
	http.NotFound(w, r)
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
