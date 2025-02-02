ARG GO_VERSION=latest

FROM golang:${GO_VERSION} as builder

WORKDIR /src
COPY . .

WORKDIR /src/cmd/service

RUN go vet ./...
RUN go test -cover ./...
RUN CGO_ENABLED=1 go build -o /bin/runner

FROM gcr.io/distroless/base as runtime

COPY --from=builder /bin/runner /

CMD ["/runner"]
