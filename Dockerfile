FROM golang:1.23-alpine AS build

WORKDIR /src
COPY go.mod main.go ./

ARG VERSION=dev
RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -ldflags="-s -w -X main.version=${VERSION}" \
    -o /out/nkp-dummy .

FROM scratch

COPY --from=build /out/nkp-dummy /nkp-dummy

EXPOSE 8080
USER 65532:65532
ENTRYPOINT ["/nkp-dummy"]
