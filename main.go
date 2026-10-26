package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"
)

const (
	appName           = "nkp-dummy"
	defaultPort       = "8080"
	headerTimeout     = 5 * time.Second
	fingerprintLength = 12
	secretEnvKey      = "DB_PASSWORD"
)

// version diisi saat build lewat -ldflags.
var version = "dev"

type infoResponse struct {
	App      string `json:"app"`
	Version  string `json:"version"`
	Hostname string `json:"hostname"`
	Time     string `json:"time"`
}

// configResponse tidak pernah memuat nilai secret. Fingerprint dipakai untuk
// membandingkan secret di OCP dan NKP tanpa membuka isinya.
type configResponse struct {
	Platform          string `json:"platform"`
	Environment       string `json:"environment"`
	Message           string `json:"message"`
	SecretLoaded      bool   `json:"secretLoaded"`
	SecretFingerprint string `json:"secretFingerprint,omitempty"`
}

func main() {
	server := &http.Server{
		Addr:              listenAddr(),
		Handler:           routes(),
		ReadHeaderTimeout: headerTimeout,
	}
	log.Printf("listening on %s version=%s", server.Addr, version)
	log.Fatal(server.ListenAndServe())
}

func routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", handleInfo)
	mux.HandleFunc("GET /config", handleConfig)
	mux.HandleFunc("GET /healthz", handleLive)
	mux.HandleFunc("GET /readyz", handleLive)
	return mux
}

func listenAddr() string {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}
	return ":" + port
}

func handleInfo(w http.ResponseWriter, _ *http.Request) {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}
	writeJSON(w, infoResponse{
		App:      appName,
		Version:  version,
		Hostname: hostname,
		Time:     time.Now().UTC().Format(time.RFC3339),
	})
}

func handleConfig(w http.ResponseWriter, _ *http.Request) {
	secret := os.Getenv(secretEnvKey)
	writeJSON(w, configResponse{
		Platform:          os.Getenv("PLATFORM"),
		Environment:       os.Getenv("APP_ENV"),
		Message:           os.Getenv("APP_MESSAGE"),
		SecretLoaded:      secret != "",
		SecretFingerprint: fingerprint(secret),
	})
}

func handleLive(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, map[string]string{"status": "ok"})
}

func fingerprint(value string) string {
	if value == "" {
		return ""
	}
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])[:fingerprintLength]
}

func writeJSON(w http.ResponseWriter, body any) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(body); err != nil {
		log.Printf("write response: %v", err)
	}
}
