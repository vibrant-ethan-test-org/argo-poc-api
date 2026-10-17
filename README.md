# argo-poc-api

Second service in the Argo CD + Argo Rollouts proof of concept: a small JSON API.
Deployment config and the walkthrough live in [argo-poc-gitops](https://github.com/vibrant-ethan-test-org/argo-poc-gitops).

| Path | What |
|---|---|
| `cmd/api/` | Service: `/api/info`, `/healthz`, `/version` |
| `cmd/e2e/` | e2e checks, shipped as `/e2e` in the same image and run by Argo Rollouts against the preview Service |
| `.github/workflows/ci.yaml` | Calls the shared `release.yaml` in the gitops repo: build, push to GHCR, deploy to test |

Where each commit is deployed shows in this repo's **Environments** panel (test, prod).

## Promotion controller fixture

`deploy/pipeline.yaml` registers this project with the promotion controller. Its `test`, `staging` and `prod`
stages use `deploy/overlays/<stage>`, backed by the blue-green Rollout and e2e AnalysisTemplate in `deploy/base`.
The older `deploy/manifests` and `deploy/octopus` layouts remain available for the other PoCs.

The controller pins image digests and supplies release, commit and deployment annotations through Argo CD's
Kustomize overrides. The included transformer configurations propagate these overrides into the Rollout pod
template and the AnalysisTemplate Job. Git does not select image tags. Each local stage runs one replica.

CI validates PR builds without publishing. Main pushes and explicit workflow dispatches retain the `v0.2.<run>`
release tags and publish revision and creation annotations on the image index and platform manifests.
