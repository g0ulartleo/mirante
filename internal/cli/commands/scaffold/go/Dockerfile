FROM golang:1.24.2-alpine AS deps
WORKDIR /app
COPY go.mod ./
RUN go mod download

FROM deps AS dev
WORKDIR /app
COPY . .
EXPOSE 50051
CMD ["go", "run", "./cmd/runtime"]

FROM deps AS build
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/runtime ./cmd/runtime

FROM alpine:3.22 AS prod
WORKDIR /app
RUN apk add --no-cache ca-certificates
COPY --from=build /out/runtime ./runtime
EXPOSE 50051
CMD ["./runtime"]
