# Frontman

Frontman Gateway is an API gateway that provides a reverse proxy and load balancing functionality to route incoming HTTP requests to backend services. By default, the Frontman Gateway runs on port 8080.

In addition, Frontman Gateway provides a set of API endpoints to manage the registered backend services. These API endpoints run on a separate port, By default, the services API endpoints run on port 8000.

When a new backend service is registered via the /api/services endpoint, the Frontman Gateway adds it to its registry and starts forwarding incoming requests to that service. The service can then be updated or removed at any time via the /api/services/{name} endpoints.

It's recommended to keep the services API endpoints on a separate port to minimize the risk of conflicts or collisions with the incoming HTTP traffic.

<p>&nbsp;</p>

[![Go Report Card](https://goreportcard.com/badge/github.com/hyperioxx/frontman)](https://goreportcard.com/report/github.com/hyperioxx/frontman) [![GitHub license](https://img.shields.io/github/license/Naereen/StrapDown.js.svg)](https://github.com/hyperioxx/frontman/blob/main/LICENCE) ![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/Hyperioxx/frontman)
<br />
## Features

- Reverse proxy and load balancing of requests to backend services
- Dynamic service registration and removal via a REST API
- Support for multiple backend services

## Requirements

- Go 1.18 or later
- Redis 5.0 or later

## Getting Started

1. Clone this repository:
git clone https://github.com/hyperioxx/frontman.git
cd frontman

2. Set the REDIS_URL environment variable to the URL of your Redis instance
  ```export REDIS_URL=redis://localhost:6379```
1. Start the Frontman Gateway:
 ```go run main.go```
  You should see the following output:
  ```Starting Frontman Gateway...```
  The Frontman Gateway is now running on port `8080`.

4. Register a backend service:
 ```
 curl -X POST -H "Content-Type: application/json" -d '{"name": "example", "scheme": "http", "url": "http://localhost:8000", "path": "/", "healthCheck": "/healthcheck"}' http://localhost:8080/api/services
 ```
 This registers a backend service named `example` with the URL `http://localhost:8000`.

1. Test the Frontman Gateway:
 ```curl http://localhost:8080/example/path/to/resource```



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

### DELETE /api/services/{name}

Removes a backend service.

Parameters:

- `name`: the name of the backend service to remove

Example request:

DELETE /api/services/example HTTP/1.1

Example response:

Removed service example
