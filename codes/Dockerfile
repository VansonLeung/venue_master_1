# syntax=docker/dockerfile:1
ARG GO_VERSION=1.23
FROM golang:${GO_VERSION} AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download || true
COPY . .

ARG CMD_PATH
ARG BIN_NAME=service
RUN test -n "$CMD_PATH" || (echo "CMD_PATH build arg is required" && exit 1)
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/${BIN_NAME} ./${CMD_PATH}

FROM gcr.io/distroless/base-debian12
WORKDIR /
COPY --from=builder /out /bin
ENV PORT=8080
EXPOSE 8080
ENTRYPOINT ["/bin/service"]
