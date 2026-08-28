# THis is a muilti-stage build: compile to Go app in a builder image, then copy only the binary and migrations into a small runtime image.
# Shorter build time

FROM golang:1.26 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /server ./cmd/server

FROM alpine:latest
WORKDIR /
COPY --from=builder /server /server
COPY migrations /migrations
EXPOSE 8080
ENTRYPOINT ["/server"]