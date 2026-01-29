# STAGE 1: Build
FROM golang:1.24-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy dependency files first to leverage Docker cache optimization
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application source code
COPY . .

# Compile the static binary
# CGO_ENABLED=0 ensures the binary is statically linked and portable
RUN CGO_ENABLED=0 GOOS=linux go build -o jota-server ./cmd/server/main.go

# STAGE 2: Runtime (Clean final image)
FROM alpine:latest

WORKDIR /app

# Install CA certificates (required for secure HTTPS calls to Gemini API)
RUN apk --no-cache add ca-certificates

# Copy only the compiled binary from the builder stage
COPY --from=builder /app/jota-server .

# Copy the knowledge base (RAG source)
COPY --from=builder /app/kb ./kb

# Expose the application port
EXPOSE 8080

# Run the application
CMD ["./jota-server"]