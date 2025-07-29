FROM golang:1.23.4-alpine AS builder

WORKDIR /usr/local/src

RUN apk --no-cache add bash gcc musl-dev

COPY ["app/go.mod", "app/go.sum", "./"]
RUN go mod download

COPY app ./
RUN go build -o ./bin/app cmd/main.go

FROM alpine AS runner

COPY --from=builder /usr/local/src/bin/app /
COPY config/config.yaml config/config.yaml

CMD ["/app"]