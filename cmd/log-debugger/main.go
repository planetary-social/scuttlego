package main

import (
	"bytes"
	"embed"
	_ "embed"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"strings"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/cmd/log-debugger/debugger"
	"github.com/planetary-social/scuttlego/cmd/log-debugger/debugger/log"
)

//go:embed output.tmpl
var outputTemplate string

//go:embed assets/*
var assets embed.FS

func main() {
	if err := run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func run() error {
	logFilename := os.Args[1]

	log, err := log.LoadLog(logFilename)
	if err != nil {
		return errors.Wrap(err, "failed to load the log")
	}

	g := debugger.NewPeers()
	for _, entry := range log {
		if err := g.Add(entry); err != nil {
			return errors.Wrapf(err, "error adding an entry '%+v'", entry)
		}
	}

	b, err := createReport(logFilename, g)
	if err != nil {
		return errors.Wrap(err, "error creating the report")
	}

	port := 8080

	fmt.Printf("http://localhost:%d\n", port)

	return http.ListenAndServe(fmt.Sprintf(":%d", port), http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if strings.HasPrefix(request.URL.Path, "/assets") {
			http.FileServer(http.FS(assets)).ServeHTTP(writer, request)
			return
		}

		_, err := writer.Write(b)
		if err != nil {
			fmt.Println("writer error", err)
		}
	}))
}

func createReport(logFilename string, peers debugger.Peers) ([]byte, error) {
	var funcMap = template.FuncMap{
		//"InitiatedByRemote": func() debugger.StreamInitiatedBy { return debugger.InitiatedByRemoteNode },
		"MessageTypeSent": func() debugger.MessageType { return debugger.MessageTypeSent },
	}

	tmpl, err := template.New("output").Funcs(funcMap).Parse(outputTemplate)
	if err != nil {
		return nil, errors.Wrap(err, "error creating a template")
	}

	buf := &bytes.Buffer{}

	if err = tmpl.Execute(buf, struct {
		LogFilename string
		Peers       debugger.Peers
	}{
		LogFilename: logFilename,
		Peers:       peers,
	}); err != nil {
		return nil, errors.Wrap(err, "error executing the template")
	}

	return buf.Bytes(), nil
}
