// Command api is a small JSON service, the second service in the Argo CD + Argo Rollouts PoC.
package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
)

// Set at build time with -ldflags "-X main.version=... -X main.commit=...".
var (
	version = "dev"
	commit  = "unknown"
)

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

func main() {
	env := os.Getenv("APP_ENV")
	pod, _ := os.Hostname()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok\n"))
	})
	mux.HandleFunc("GET /version", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, map[string]string{"version": version, "commit": commit, "env": env})
	})
	mux.HandleFunc("GET /api/info", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, map[string]string{"service": "argo-poc-api", "version": version, "env": env, "pod": pod, "message": "hello from argo-poc-api"})
	})

	log.Printf("argo-poc-api %s (%s) listening on :8080", version, commit)
	log.Fatal(http.ListenAndServe(":8080", mux))
}
