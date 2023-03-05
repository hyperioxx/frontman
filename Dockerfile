# Use an official Golang runtime as a parent image
FROM golang:1.18-alpine AS build

# Install GNU Make
RUN apk add --no-cache make

# Set the working directory
WORKDIR /go/src/app

# Copy the source code to the container
COPY . .

# Build the binary using GNU Make
RUN CGO_ENABLED=0 make all

# Use an official lightweight Alpine image as a parent image
FROM alpine:latest

# Set the working directory
WORKDIR /app

# Copy the binary from the previous stage
COPY --from=build /go/src/app/bin/frontman .

WORKDIR /app
# Make the binary executable
RUN chmod +x /app/frontman

COPY frontman.yaml /app/frontman.yaml
COPY entrypoint.sh /app/entrypoint.sh


# Expose the ports
EXPOSE 8080 8000

# Start the service
CMD ["./entrypoint.sh"]

