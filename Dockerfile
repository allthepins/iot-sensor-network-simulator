# Stage 1: build application

FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -o /app/simulator ./cmd/simulator

# Stage 2: final image

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/simulator /app/simulator

ENTRYPOINT ["/app/simulator"]
