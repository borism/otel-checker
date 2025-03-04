package main

import (
	"embed"
	_ "embed"
	"html/template"
	"log"
	"net/http"

	"otel-checker/checks"
	"otel-checker/checks/utils"
)

//go:embed static/*
var static embed.FS

//go:embed tmpl/*
var tmpls embed.FS

var messages map[string][]string

func main() {
	commands := utils.GetArguments()
	messages = checks.RunAllChecks(commands)

	if !commands.WebServer {
		return
	}

	mux := http.NewServeMux()

	t, err := template.ParseFS(tmpls, "tmpl/*.tmpl")
	if err != nil {
		panic(err)
	}

	mux.Handle("/static/", http.FileServer(http.FS(static)))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		err = t.ExecuteTemplate(w, "index.html.tmpl", struct {
			Messages map[string][]string
		}{
			Messages: messages,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	log.Println("Application available on http://localhost:8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Println("server failed:", err)
	}
}
