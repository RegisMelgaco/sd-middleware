package sdmiddleware

import (
	"embed"
	"fmt"
	"html/template"
	"net/http"
	"sync"
)

type GE struct {
	Addr   string
	Broker BrokerClient
}

func (g GE) Run() {
	mux := http.NewServeMux()

	mux.HandleFunc("/", g.handleHome)
	mux.HandleFunc("/messages", g.handleMsgs)
	mux.HandleFunc("DELETE /messages/{id}", g.handleDeleteMsg)

	http.ListenAndServe(g.Addr, mux)
}

//go:embed templates
var tmplFS embed.FS

func (g GE) handleHome(w http.ResponseWriter, _ *http.Request) {
	err := indexTmpl().Execute(w, nil)
	if err != nil {
		err = fmt.Errorf("executing index.html template: %w", err)
		panic(err)
	}
}

var indexTmpl = sync.OnceValue(func() *template.Template {
	tmpl, err := template.ParseFS(tmplFS, "templates/index.html")
	if err != nil {
		panic(fmt.Errorf("loading index.html: %w", err))
	}

	return tmpl
})

func (g GE) handleMsgs(w http.ResponseWriter, _ *http.Request) {
	topics := g.Broker.List()

	err := msgsTmpl().Execute(w, topics)
	if err != nil {
		err = fmt.Errorf("executing index.html template: %w", err)
		panic(err)
	}
}

var msgsTmpl = sync.OnceValue(func() *template.Template {
	tmpl, err := template.ParseFS(tmplFS, "templates/msgs.html")
	if err != nil {
		panic(fmt.Errorf("loading msgs.html: %w", err))
	}

	return tmpl
})

func (g GE) handleDeleteMsg(w http.ResponseWriter, r *http.Request) {
	g.Broker.Delete(r.PathValue("id"))
}
