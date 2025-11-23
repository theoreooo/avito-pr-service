FROM golang:1.25-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o app ./cmd/app

FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/app .
COPY --from=builder /app/config ./config
COPY --from=builder /app/openapi.yaml .

EXPOSE 8080

USER nobody

CMD ["./app"]