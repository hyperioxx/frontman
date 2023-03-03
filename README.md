# Frontman
Frontman is an open-source API gateway written in Go that allows you to manage your microservices and expose them as a single API endpoint. It acts as a reverse proxy and handles requests from clients, routing them to the appropriate backend service.

Frontman provides a simple, flexible, and scalable solution for managing microservices. It includes features such as and automatic health checks, making it easy to manage and maintain your API gateway and backend services.

With Frontman, you can easily add new backend services, update existing ones, and remove them, without affecting the clients. Frontman also provides an HTTP API for managing backend services, making it easy to automate the management of your API gateway.

Frontman is designed to be highly available and fault-tolerant. It uses Redis as a data store for managing backend services, ensuring that your services are always available and up-to-date. It also supports distributed deployments, allowing you to scale horizontally as your traffic and service requirements grow.

Overall, Frontman is a powerful and flexible API gateway that simplifies the management of microservices, making it easier for developers to build and maintain scalable and reliable API endpoints.

<p>&nbsp;</p>

[![Go Report Card](https://goreportcard.com/badge/github.com/hyperioxx/frontman)](https://goreportcard.com/report/github.com/hyperioxx/frontman) [![GitHub license](https://img.shields.io/github/license/Naereen/StrapDown.js.svg)](https://github.com/hyperioxx/frontman/blob/main/LICENCE) ![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/Hyperioxx/frontman)
<br />

##Features
- Reverse proxy requests to backend services
- Health checks for backend services
- Dynamic backend service configuration using Redis
  
##Usage
####Configuration

Frontman is configured using environment variables. The following variables are supported:
|Environment Variable| Description| Default|
|:--------------------:|:--------------:|:--------------:|
|FRONTMAN_LISTEN_ADDR | The address and port on which Frontman should listen for incoming requests| ```0.0.0.0:8080```|
|FRONTMAN_REDIS_URL | The URL of the Redis instance used for storing backend service configuration |```redis://localhost:6379```

####Starting Frontman
To start Frontman, you can download the latest release binary for your platform from the releases page or build it from source.

Once you have the binary, you can start Frontman by running:

```bash
$ ./frontman
```
This will start Frontman with the default configuration, using the Redis instance running on localhost:6379.

####Running Frontman in Docker
Frontman can also be run as a Docker container. Here's an example command to start Frontman in Docker, assuming your Redis instance is running on localhost:

```bash
$ docker run -p 8080:8080 -e FRONTMAN_REDIS_URL=redis://host.docker.internal:6379 hyperioxx/frontman:latest
```
This command starts a new container, maps port 8080 on the host to port 8080 in the container, and sets the FRONTMAN_REDIS_URL environment variable to the URL of the Redis instance. Note that in this example, we're using host.docker.internal to reference the Redis instance running on the host machine, but you can replace this with the actual IP or hostname of your Redis instance.


##Managing Backend Services
Frontman uses Redis to store the configuration for backend services. Backend services are represented as JSON objects and stored in a Redis list. Here's an example of a backend service configuration:

```json
{
	"name": "Example Service",
	"scheme": "http",
	"url": "example.com",
	"path": "/api",
	"domain": "",
	"healthCheck": "http://example.com/health",
	"retryAttempts": 3,
	"timeout": "10s",
	"maxIdleConns": 100,
	"maxIdleTime": "60s"
}
```
You can add, update, and remove backend services using the following REST endpoints:

- GET /services - Retrieves a list of all backend services
- GET /health - Performs a health check on all backend services and returns the status of each service
- POST /services - Adds a new backend service
- PUT /services/{name} - Updates an existing backend service
- DELETE /services/{name} - Removes a backend service
##Contributing
If you'd like to contribute to Frontman, please fork the repository and submit a pull request. We welcome bug reports, feature requests, and code contributions.

##License
Frontman is released under the GNU General Public License. See LICENSE for details.