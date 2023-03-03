# Use an official Golang runtime as a parent image
FROM golang:1.17.5-alpine AS build

# Set the working directory
WORKDIR /go/src/app

# Copy the source code to the container
COPY . .

# Build the binary
RUN CGO_ENABLED=0 go build -o /go/bin/frontman .

# Use an official lightweight Alpine image as a parent image
FROM alpine:latest

# Set the working directory
WORKDIR /app

# Copy the binary from the previous stage
COPY --from=build /go/bin/frontman .

# Expose the ports
EXPOSE 8080 8000

# Start the service
CMD ["./frontman"]