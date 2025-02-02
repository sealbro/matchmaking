ARG GO_VERSION=latest

FROM golang:${GO_VERSION} as builder

WORKDIR /src
COPY . .

ARG CMD_PATH=cmd/service
WORKDIR /src/${CMD_PATH}

RUN CGO_ENABLED=0 go build -o /bin/runner

FROM gcr.io/distroless/base as runtime

COPY --from=builder /bin/runner /

CMD ["/runner"]
