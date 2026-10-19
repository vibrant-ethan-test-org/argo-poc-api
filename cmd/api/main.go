// Command api is a small JSON service, the second service in the Argo CD + Argo Rollouts PoC.
package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

// Set at build time with -ldflags "-X main.version=... -X main.commit=...".
var (
	version        = "dev"
	commit         = "unknown"
	unhealthyAfter = "" // UNHEALTHY_AFTER build flag, staging only (controller acceptance UC5).
)

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

func main() {
	env := os.Getenv("APP_ENV")
	pod, _ := os.Hostname()

	started := time.Now()
	delay, err := time.ParseDuration(unhealthyAfter)
	if unhealthyAfter != "" && (err != nil || delay <= 0 || delay > time.Hour) {
		log.Fatal("invalid UNHEALTHY_AFTER")
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", healthHandler(env, delay, started))
	mux.HandleFunc("GET /version", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, map[string]string{"version": version, "commit": commit, "env": env})
	})
	mux.HandleFunc("GET /api/info", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, map[string]string{"service": "argo-poc-api", "version": version, "env": env, "pod": pod, "message": "hello from argo-poc-api"})
	})

	log.Printf("argo-poc-api %s (%s) listening on :8080", version, commit)
	log.Fatal(http.ListenAndServe(":8080", mux))
}

// Staging loses readiness during its bake; test and prod remain healthy with the same image.
func healthHandler(env string, delay time.Duration, started time.Time) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if delay > 0 && strings.HasSuffix(env, "-staging") && time.Since(started) >= delay {
			http.Error(w, "fixture unhealthy", http.StatusServiceUnavailable)
			return
		}
		w.Write([]byte("ok\n"))
	}
}
