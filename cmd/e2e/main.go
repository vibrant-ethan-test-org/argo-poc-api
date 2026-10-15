// Command e2e checks a deployed argo-poc-api. Argo Rollouts runs it (as /e2e in the same image)
// against the preview Service before switching traffic; a non-zero exit aborts the rollout.
package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

var client = &http.Client{Timeout: 5 * time.Second}

func getJSON(url string, v any) error {
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status %d", resp.StatusCode)
	}
	if v == nil {
		return nil
	}
	return json.NewDecoder(resp.Body).Decode(v)
}

func main() {
	target := strings.TrimRight(os.Getenv("TARGET_URL"), "/")
	expectedSHA := os.Getenv("EXPECTED_SHA")
	if target == "" {
		fmt.Println("FAIL: TARGET_URL is not set")
		os.Exit(2)
	}
	fmt.Printf("e2e against %s (expected commit %q)\n", target, expectedSHA)

	checks := []struct {
		name string
		run  func() error
	}{
		{"healthz returns 200", func() error { return getJSON(target+"/healthz", nil) }},
		{"version matches the commit being deployed", func() error {
			var v struct{ Commit string }
			if err := getJSON(target+"/version", &v); err != nil {
				return err
			}
			if expectedSHA != "" && v.Commit != expectedSHA {
				return fmt.Errorf("commit %q, want %q", v.Commit, expectedSHA)
			}
			return nil
		}},
		{"api/info identifies the service", func() error {
			var v struct{ Service string }
			if err := getJSON(target+"/api/info", &v); err != nil {
				return err
			}
			if v.Service != "argo-poc-api" {
				return fmt.Errorf("service %q, want argo-poc-api", v.Service)
			}
			return nil
		}},
	}

	failed := 0
	for _, c := range checks {
		var err error
		// Retry briefly: the preview Service can take a moment to route to new pods.
		for attempt := 1; attempt <= 5; attempt++ {
			if err = c.run(); err == nil {
				break
			}
			time.Sleep(2 * time.Second)
		}
		if err != nil {
			failed++
			fmt.Printf("FAIL  %s: %v\n", c.name, err)
		} else {
			fmt.Printf("PASS  %s\n", c.name)
		}
	}
	if failed > 0 {
		fmt.Printf("%d check(s) failed\n", failed)
		os.Exit(1)
	}
	fmt.Println("all checks passed")
}
