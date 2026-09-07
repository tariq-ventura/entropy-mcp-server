FROM golang:1.27 AS bin-stage

SHELL ["/bin/bash", "-c"]

RUN mkdir -p /github.com/tariq-ventura/entropy-mcp-server

WORKDIR /github.com/tariq-ventura/entropy-mcp-server

COPY . .

RUN go mod download && GODEBUG=http2client=0 go mod tidy

RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/api ./cmd/services/main.go

FROM gcr.io/distroless/static-debian12 AS release-stage

WORKDIR /

COPY --from=bin-stage /bin/api /bin/api
VOLUME /var/log

EXPOSE 3002

CMD ["/bin/api"]