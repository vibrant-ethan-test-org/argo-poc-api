# argo-poc-api

Second service in the Argo CD + Argo Rollouts proof of concept: a small JSON API.
Deployment config and the walkthrough live in [argo-poc-gitops](https://github.com/vibrant-ethan-test-org/argo-poc-gitops).

| Path | What |
|---|---|
| `cmd/api/` | Service: `/api/info`, `/healthz`, `/version` |
| `cmd/e2e/` | e2e checks, shipped as `/e2e` in the same image and run by Argo Rollouts against the preview Service |
| `.github/workflows/ci.yaml` | Calls the shared `release.yaml` in the gitops repo: build, push to GHCR, deploy to test |

Where each commit is deployed shows in this repo's **Environments** panel (test, prod).
