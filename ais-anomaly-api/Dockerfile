FROM golang:1.25.1-alpine

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN go build -o /ais-anomaly-api .

# Expose port
EXPOSE 3000

# Run the application
CMD ["/ais-anomaly-api"]
