// Command e2e checks a deployed argo-poc-api. The Harness pipeline runs it (as /e2e in the same image)
// against the stage Service before swapping traffic; a non-zero exit fails the stage.
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

// failStage is set at build time from the E2E_FAIL build argument (-X main.failStage=<stage>). The promotion
// controller's e2e tests use it to build a release whose e2e fails on purpose in that stage.
var failStage string

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
	expectedVersion := os.Getenv("EXPECTED_VERSION")
	if target == "" {
		fmt.Println("FAIL: TARGET_URL is not set")
		os.Exit(2)
	}
	fmt.Printf("e2e against %s (expected version %q, commit %q)\n", target, expectedVersion, expectedSHA)

	checks := []struct {
		name string
		run  func() error
	}{
		{"healthz returns 200", func() error { return getJSON(target+"/healthz", nil) }},
		{"version matches the release being deployed", func() error {
			var v struct{ Version, Commit string }
			if err := getJSON(target+"/version", &v); err != nil {
				return err
			}
			if expectedVersion != "" && v.Version != expectedVersion {
				return fmt.Errorf("version %q, want %q", v.Version, expectedVersion)
			}
			if expectedSHA != "" && v.Commit != expectedSHA {
				return fmt.Errorf("commit %q, want %q", v.Commit, expectedSHA)
			}
			return nil
		}},
		{"build is not marked to fail in this stage", func() error {
			if failStage == "" {
				return nil
			}
			var v struct{ Env string }
			if err := getJSON(target+"/version", &v); err != nil {
				return err
			}
			// Env is the stage namespace, <project>-<stage>.
			if v.Env == failStage || strings.HasSuffix(v.Env, "-"+failStage) {
				return fmt.Errorf("built with E2E_FAIL=%s, running in %s", failStage, v.Env)
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
