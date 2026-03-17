# Build stage
FROM golang:1.25-alpine AS builder

# Set working directory
WORKDIR /app

ARG GOPROXY
ENV GOPROXY=$GOPROXY

# Install git and build dependencies for CGO (needed for SQLite)
RUN apk add --no-cache git gcc musl-dev sqlite-dev

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application with CGO enabled for SQLite support
RUN CGO_ENABLED=1 GOOS=linux go build -a -ldflags '-linkmode external -extldflags "-static"' -o lazy-ai-coder .

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/lazy-ai-coder .

# Copy config files
COPY --from=builder /app/config ./config

# Copy web assets
COPY --from=builder /app/web ./web

# Create directory for images
RUN mkdir -p web/images

# Expose port
EXPOSE 8888

# Command to run
CMD ["./lazy-ai-coder", "web", "-p", "8888"] 
