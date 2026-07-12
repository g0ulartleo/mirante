FROM golang:1.24-alpine AS builder

WORKDIR /app

RUN apk add --no-cache build-base sqlite-dev make git

ENV CGO_ENABLED=1

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go install github.com/a-h/templ/cmd/templ@v0.3.906
RUN templ generate
RUN make build

FROM golang:1.24-alpine AS dev

WORKDIR /app

RUN apk add --no-cache build-base sqlite-dev git

ENV CGO_ENABLED=1

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go install github.com/a-h/templ/cmd/templ@v0.3.906 && \
    go install github.com/air-verse/air@v1.61.7

EXPOSE 40169

FROM alpine:3.20 AS prod

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata sqlite-libs && update-ca-certificates

COPY --from=builder /app/bin/ /app/bin/
COPY --from=builder /app/static /app/static

ENV HTTP_ADDR=0.0.0.0

EXPOSE 40169

CMD ["./bin/http-server"]
