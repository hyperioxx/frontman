# Frontman Gateway

Frontman Gateway is a reverse proxy and load balancer that routes requests to backend services based on the service name in the URL path. It supports dynamic service registration and removal, and can be used to build a scalable microservice architecture.

## Features

- Reverse proxy and load balancing of requests to backend services
- Dynamic service registration and removal via a REST API
- Support for multiple backend services
- HTTP and HTTPS support
- Automatic creation of database table if it does not exist

## Requirements

- Go 1.16 or later
- PostgreSQL 9.6 or later

## Getting Started

1. Clone this repository:
git clone https://github.com/hyperioxx/frontman.git
cd frontman

2. Set the `DATABASE_URI` environment variable to the URI of your PostgreSQL database:

export DATABASE_URI=postgres://user:password@host:port/database

3. Start the Frontman Gateway:

go run main.go


You should see the following output:

Starting Frontman Gateway...

The Frontman Gateway is now running on port `8080`.

4. Register a backend service:

curl -X POST -d "name=example&url=http://localhost:8000" http://localhost:8080/api/services


This registers a backend service named `example` with the URL `http://localhost:8000`.

1. Test the Frontman Gateway:

curl http://localhost:8080/example/path/to/resource



This should route the request to the `example` backend service and return the response.

## API

The Frontman Gateway provides a REST API for registering and removing backend services:

### GET /api/services

Returns a list of registered backend services.

Example response:

example: http://localhost:8000



### POST /api/services

Registers a new backend service.

Parameters:

- `name`: the name of the backend service
- `url`: the URL of the backend service

Example request:

POST /api/services HTTP/1.1
Content-Type: application/x-www-form-urlencoded

name=example&url=http://localhost:8000




Example response:

Added service example: http://localhost:8000

markdown


### DELETE /api/services/{name}

Removes a backend service.

Parameters:

- `name`: the name of the backend service to remove

Example request:

DELETE /api/services/example HTTP/1.1




Example response:

Removed service example