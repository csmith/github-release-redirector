FROM golang:1.24.5 AS build
WORKDIR /go/src/app
COPY . .

RUN --mount=type=cache,target=/go/pkg/mod \
    set -eux; \
    CGO_ENABLED=0 GO111MODULE=on go install .; \
    go run github.com/google/go-licenses@latest save ./... --save_path=/notices;

FROM ghcr.io/greboid/dockerbase/nonroot:1.20250716.0
COPY --from=build /go/bin/github-release-redirector /github-release-redirector
COPY --from=build /notices /notices
ENTRYPOINT ["/github-release-redirector"]