FROM golang:1.19-alpine AS builder

WORKDIR /app

# Copy module files first to leverage Docker cache
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o itsplanned .

# Use a smaller image for the final container
FROM alpine:latest

WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/itsplanned /app/

# Add runtime dependencies
RUN apk --no-cache add ca-certificates tzdata

# Expose port
EXPOSE 8080

# Run the application
CMD ["./itsplanned"] 