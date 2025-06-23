# ---- Build Stage ----
FROM golang:1.24.4 AS build

ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o cloudrift main.go

# ---- Run Stage ----
FROM alpine:latest

# Create non-root user named cloudrift
RUN adduser -D cloudrift
USER cloudrift

COPY --from=build /app/cloudrift /cloudrift

ENTRYPOINT ["/cloudrift"]
