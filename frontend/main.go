package main

import (
	"context"
	"encoding/json"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

const (
	defaultPort       = "8080"
	defaultBackendURL = "http://nkp-dummy:8080"
	headerTimeout     = 5 * time.Second
	backendTimeout    = 3 * time.Second
)

var version = "dev"

type pageData struct {
	Version      string
	FetchedAt    string
	BackendError string
	App          string
	BackendVer   string
	Hostname     string
	Platform     string
	Environment  string
	Message      string
	SecretLoaded bool
	Fingerprint  string
	HasBackend   bool
}

func main() {
	server := &http.Server{
		Addr:              listenAddr(),
		Handler:           routes(),
		ReadHeaderTimeout: headerTimeout,
	}
	log.Printf("frontend listening on %s backend=%s", server.Addr, backendURL())
	log.Fatal(server.ListenAndServe())
}

func routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", handlePage)
	mux.HandleFunc("GET /healthz", handleLive)
	return mux
}

func listenAddr() string {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}
	return ":" + port
}

func backendURL() string {
	if value := os.Getenv("BACKEND_URL"); value != "" {
		return value
	}
	return defaultBackendURL
}

func handleLive(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = io.WriteString(w, "{\"status\":\"ok\"}\n")
}

func handlePage(w http.ResponseWriter, r *http.Request) {
	data := pageData{
		Version:   version,
		FetchedAt: time.Now().UTC().Format("15:04:05"),
	}
	info, infoErr := getJSON(r, "/")
	config, configErr := getJSON(r, "/config")
	if infoErr != nil || configErr != nil {
		data.BackendError = "Backend tidak menjawab. Coba muat ulang."
	} else {
		data.HasBackend = true
		data.App = stringField(info, "app")
		data.BackendVer = stringField(info, "version")
		data.Hostname = stringField(info, "hostname")
		data.Platform = stringField(config, "platform")
		data.Environment = stringField(config, "environment")
		data.Message = stringField(config, "message")
		data.SecretLoaded, _ = config["secretLoaded"].(bool)
		data.Fingerprint = stringField(config, "secretFingerprint")
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := page.Execute(w, data); err != nil {
		log.Printf("render page: %v", err)
	}
}

func getJSON(r *http.Request, path string) (map[string]any, error) {
	ctx, cancel := context.WithTimeout(r.Context(), backendTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, backendURL()+path, nil)
	if err != nil {
		return nil, err
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, errStatus
	}
	var body map[string]any
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		return nil, err
	}
	return body, nil
}

var errStatus = errString("backend status")

type errString string

func (e errString) Error() string { return string(e) }

func stringField(body map[string]any, key string) string {
	value, _ := body[key].(string)
	return value
}

var page = template.Must(template.New("page").Parse(`<!DOCTYPE html>
<html lang="id">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>nkp-dummy</title>
  <style>
    :root { color-scheme: light; }
    body {
      margin: 0;
      background: #f3f0e8;
      color: #1a2332;
      font: 18px/1.45 Georgia, "Iowan Old Style", "Palatino Linotype", serif;
    }
    header, main { max-width: 38rem; margin: 0 auto; padding: 1.5rem 1.25rem; }
    header { border-bottom: 4px solid #0c88d9; }
    h1 { font-size: 2rem; font-weight: 500; margin: 0; }
    p { margin: 0.4rem 0 0; }
    .muted { color: #5c6b7a; font-family: "Segoe UI", Helvetica, Arial, sans-serif; font-size: 0.9rem; }
    dl { margin: 1.5rem 0 0; }
    div { display: grid; grid-template-columns: 9rem 1fr; gap: 0.35rem 1rem; padding: 0.7rem 0; border-bottom: 1px solid #d9d3c7; }
    dt { font-family: "Segoe UI", Helvetica, Arial, sans-serif; font-size: 0.82rem; letter-spacing: 0.04em; text-transform: uppercase; color: #5c6b7a; }
    dd { margin: 0; }
    button {
      margin-top: 1.4rem;
      background: #0c88d9;
      color: #fff;
      border: 0;
      border-radius: 2px;
      padding: 0.7rem 1.1rem;
      font: 1rem/1 "Segoe UI", Helvetica, Arial, sans-serif;
      cursor: pointer;
    }
    button:focus-visible { outline: 3px solid #1a2332; outline-offset: 3px; }
    .warn { color: #8a3b12; }
  </style>
</head>
<body>
  <header>
    <h1>nkp-dummy</h1>
    <p class="muted">Frontend lab. Data di bawah diambil dari service backend di namespace yang sama.</p>
  </header>
  <main>
    {{if .BackendError}}
    <p class="warn">{{.BackendError}}</p>
    {{else}}
    <dl>
      <div><dt>Platform</dt><dd>{{.Platform}}</dd></div>
      <div><dt>Environment</dt><dd>{{.Environment}}</dd></div>
      <div><dt>Pesan</dt><dd>{{.Message}}</dd></div>
      <div><dt>Aplikasi</dt><dd>{{.App}}</dd></div>
      <div><dt>Versi backend</dt><dd>{{.BackendVer}}</dd></div>
      <div><dt>Pod backend</dt><dd>{{.Hostname}}</dd></div>
      <div><dt>Secret</dt><dd>{{if .SecretLoaded}}terbaca{{else}}kosong{{end}}</dd></div>
      <div><dt>Sidik jari</dt><dd>{{.Fingerprint}}</dd></div>
    </dl>
    {{end}}
    <p class="muted">Frontend {{.Version}} · diambil {{.FetchedAt}} UTC</p>
    <form method="get" action="/"><button type="submit">Muat ulang</button></form>
  </main>
</body>
</html>
`))
