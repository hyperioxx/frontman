# Frontman
Frontman is an open-source API gateway written in Go that allows you to manage your microservices and expose them as a single API endpoint. It acts as a reverse proxy and handles requests from clients, routing them to the appropriate backend service.

Frontman provides a simple, flexible, and scalable solution for managing microservices. It includes features such as and automatic health checks, making it easy to manage and maintain your API gateway and backend services.

With Frontman, you can easily add new backend services, update existing ones, and remove them, without affecting the clients. Frontman also provides an HTTP API for managing backend services, making it easy to automate the management of your API gateway.

Frontman is designed to be highly available and fault-tolerant. It uses Redis as a data store for managing backend services, ensuring that your services are always available and up-to-date. It also supports distributed deployments, allowing you to scale horizontally as your traffic and service requirements grow.

Overall, Frontman is a powerful and flexible API gateway that simplifies the management of microservices, making it easier for developers to build and maintain scalable and reliable API endpoints.

<p>&nbsp;</p>

[![Go Report Card](https://goreportcard.com/badge/github.com/hyperioxx/frontman)](https://goreportcard.com/report/github.com/hyperioxx/frontman) [![GitHub license](https://img.shields.io/github/license/hyperioxx/frontman)](https://github.com/hyperioxx/frontman/blob/main/LICENCE) ![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/Hyperioxx/frontman)
<br />

## Roadmap

#### Phase 1: MVP (Minimum Viable Product)

- Complete basic functionality to act as an API gateway.
- Implement support for popular protocols like HTTP, HTTPS, WebSocket, and TCP.
- Implement support for commonly used load balancing algorithms like round-robin, least connection, and IP hash.
- Basic health check support.
- Basic logging support.
- Basic rate limiting support.

#### Phase 2: Security and Authentication

- Implement support for secure communication with clients using SSL/TLS.
- Implement support for client authentication using certificates.
- Implement support for basic authentication using username and password.
- Implement support for OAuth 2.0 authentication.
- Implement support for JSON Web Tokens (JWT) authentication.
- Implement support for API keys.
  
#### Phase 3: Monitoring and Analytics

- Implement support for distributed tracing with OpenTelemetry.
- Implement support for custom metrics collection and analysis.
- Implement support for log aggregation and analysis with popular tools like ELK stack or Splunk.
  
#### Phase 4: Advanced Features

- Implement support for service discovery using popular tools like Consul, Zookeeper or etcd.
- Implement support for advanced routing rules based on request headers, query parameters, or path patterns.
- Implement support for WebSocket message routing.
- Implement support for Webhooks and callbacks.
- Implement support for caching using popular tools like Redis or Memcached.
- Implement support for API versioning.

## Features
- Reverse proxy requests to backend services
- Dynamic backend service configuration using Redis as a backend database
- Automatic refresh of connections to upstream targets with configurable timeouts and maximum idle connections
- Option to strip the service path from requests before forwarding to upstream targets
- Written in Go for efficient performance and concurrency support
  
## Usage
### Configuration

#### Env Variables
Frontman is configured using environment variables. The following variables are supported:
|Environment Variable| Description| Default|
|:--------------------:|:--------------:|:--------------:|
|FRONTMAN_SERVICE_TYPE|The service type to use|`yaml`|
|FRONTMAN_SERVICES_FILE|The path to the services file.|`services.yaml`|
|FRONTMAN_LOG_LEVEL|The log level to use|`info`|
|FRONTMAN_API_ADDR | The address and port on which the /api/services endpoint should listen for incoming requests | ```0.0.0.0:8080```|
|FRONTMAN_GATEWAY_ADDR | The address and port on which the gateway should listen for incoming requests| ```0.0.0.0:8000```|
|FRONTMAN_REDIS_URI | The URI of the Redis instance used for storing backend service configuration |```redis://localhost:6379```

#### Frontman Configuration File

This describes the structure and options of the Frontman configuration file. This file is used to configure the Frontman API gateway and services.

The configuration file is written in YAML format and is structured as follows:
```yaml
global:
  service_type: [SERVICE TYPE]
  services_file: [SERVICES FILE]
  redis_namespace: [REDIS NAMESPACE]
  redis_uri: [REDIS URI]

api:
  addr: [API ADDRESS]

gateway:
  addr: [GATEWAY ADDRESS]

logging:
  level: [LOG LEVEL]

```

#### Global Section
The global section contains global configuration options that apply to both the API and Gateway.

|Key| Description|	Default Value|
|:--:|:---:|:---:|
|service_type|	The type of service registry used to store backend services. Valid options are yaml and redis.|	yaml|
|services_file|	The path to the YAML file used to store backend services when using the yaml service registry.|	services.yaml|
|redis_namespace|	The namespace used to prefix all Redis keys when using the redis service registry.|	frontman|
|redis_uri|is a string representing the URI of the Redis server that the application will use to store and retrieve backend services data. ||

#### API Section
The api section contains configuration options for the Frontman API.

|Key| Description|Default Value|
|:--:|:---:|:---:|
|addr|	The address on which the Frontman API will listen.|	0.0.0.0:8080|

#### Gateway Section
The gateway section contains configuration options for the Frontman Gateway.

|Key| Description|Default Value|
|:--:|:---:|:---:|
|addr|	The address on which the Frontman Gateway will listen.	|0.0.0.0:8000|

#### Logging Section
The logging section contains configuration options for the Frontman logging.

|Key| Description|Default Value|
|:--:|:---:|:---:|
|level|	The log level of the Frontman logging. Valid options are debug, info, warn, error, and fatal.|	info

#### Configuration Options
`service_type`
The service_type option specifies the type of service registry used to store backend services. The two valid options are yaml and redis. When using the yaml service registry, the services_file option is also required.

`services_file`
The services_file option specifies the path to the YAML file used to store backend services when using the yaml service registry. This option is required when using the yaml service registry.

`redis_namespace`
The redis_namespace option specifies the namespace used to prefix all Redis keys when using the redis service registry.

`api.addr`
The api.addr option specifies the address on which the Frontman API will listen.

`gateway.addr`
The gateway.addr option specifies the address on which the Frontman Gateway will listen.

`logging.level`
The logging.level option specifies the log level of the Frontman logging. Valid options are debug, info, warn, error, and fatal. The default log level is info.

### Starting Frontman
To start Frontman, you can download the latest release binary for your platform from the releases page or build it from source.

##### Building from Source
To build the application from source, make sure you have Go installed on your system.

Clone the repository:
```bash
$ git clone https://github.com/hyperioxx/frontman.git
$ cd frontman
```
Build the binaries:
```
$ make all
```
This will build all programs defined in the cmd directory and place the resulting binaries in the bin directory.

(Optional) Clean up the binaries:
```bash
$ make clean
```
This will remove all binaries from the bin directory.

#### Running Frontman Locally

Once you have the binary, you can start Frontman by running:

```bash
$ ./frontman
```
This will start Frontman with the default configuration, using the Redis instance running on localhost:6379.

#### Running Frontman in Docker
Frontman can also be run as a Docker container. Here's an example command to start Frontman in Docker, assuming your Redis instance is running on localhost:

```bash
$ docker run -p 8080:8080 -e FRONTMAN_REDIS_URL=redis://host.docker.internal:6379 hyperioxx/frontman:latest
```
This command starts a new container, maps port 8080 on the host to port 8080 in the container, and sets the FRONTMAN_REDIS_URL environment variable to the URL of the Redis instance. Note that in this example, we're using host.docker.internal to reference the Redis instance running on the host machine, but you can replace this with the actual IP or hostname of your Redis instance.


## Managing Backend Services
Frontman uses Redis to store the configuration for backend services. Backend services are represented as JSON objects and stored in a Redis list. Here's an example of a backend service configuration:

```json
{
	"name": "Example Service",
	"scheme": "http",
	"upstreamTargets":["service1:8000", "service2:8000"],
	"path": "/test",
	"domain": "",
	"maxIdleConns": 100,
	"maxIdleTime": 60
}
```
You can add, update, and remove backend services using the following REST endpoints:

- GET /services - Retrieves a list of all backend services
- GET /health - Performs a health check on all backend services and returns the status of each service
- POST /services - Adds a new backend service
- PUT /services/{name} - Updates an existing backend service
- DELETE /services/{name} - Removes a backend service
## Contributing
If you'd like to contribute to Frontman, please fork the repository and submit a pull request. We welcome bug reports, feature requests, and code contributions.

## License
Frontman is released under the GNU General Public License. See LICENSE for details.