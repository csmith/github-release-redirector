FROM golang:1.12 AS build

WORKDIR /go/src/app

COPY . .
RUN CGO_ENABLED=0 GO111MODULE=on go install .

FROM scratch
COPY --from=build /go/bin/github-release-redirector /github-release-redirector
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

ENTRYPOINT ["/github-release-redirector"]