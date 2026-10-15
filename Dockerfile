# One image with two binaries: /api (default) and /e2e (run by Argo Rollouts before switching traffic).
FROM --platform=$BUILDPLATFORM golang:1.24-alpine AS build
ARG TARGETOS TARGETARCH
WORKDIR /src
COPY . .
ARG VERSION=dev
ARG COMMIT=unknown
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -trimpath \
      -ldflags "-s -w -X main.version=${VERSION} -X main.commit=${COMMIT}" -o /out/api ./cmd/api && \
    CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -trimpath -ldflags "-s -w" -o /out/e2e ./cmd/e2e

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/api /out/e2e /
EXPOSE 8080
ENTRYPOINT ["/api"]
