# Frontman
Frontman is an open-source API gateway written in Go that allows you to manage your microservices and expose them as a single API endpoint. It acts as a reverse proxy and handles requests from clients, routing them to the appropriate backend service.

Frontman provides a simple, flexible, and scalable solution for managing microservices. It includes features such as and automatic health checks, making it easy to manage and maintain your API gateway and backend services.

With Frontman, you can easily add new backend services, update existing ones, and remove them, without affecting the clients. Frontman also provides an HTTP API for managing backend services, making it easy to automate the management of your API gateway.

Frontman is designed to be highly available and fault-tolerant. It uses Redis as a data store for managing backend services, ensuring that your services are always available and up-to-date. It also supports distributed deployments, allowing you to scale horizontally as your traffic and service requirements grow.

Overall, Frontman is a powerful and flexible API gateway that simplifies the management of microservices, making it easier for developers to build and maintain scalable and reliable API endpoints.

<p>&nbsp;</p>

[![Go Report Card](https://goreportcard.com/badge/github.com/hyperioxx/frontman)](https://goreportcard.com/report/github.com/hyperioxx/frontman) [![GitHub license](https://img.shields.io/github/license/hyperioxx/frontman)](https://github.com/hyperioxx/frontman/blob/main/LICENCE) ![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/Hyperioxx/frontman)
<br />

## Features
- Reverse proxy requests to backend services
- Dynamic backend service configuration using Redis as a backend database
- Automatic refresh of connections to upstream targets with configurable timeouts and maximum idle connections
- TLS encryption for secure communication with clients
- Option to strip the service path from requests before forwarding to upstream targets
- Written in Go for efficient performance and concurrency support
  
## Usage
### Configuration

#### Env Variables
Frontman is configured using environment variables. The following variables are supported:
|Environment Variable| Description| Default|
|:--------------------:|:--------------:|:--------------:|
|FRONTMAN_SERVICE_TYPE	|The service type to use|	yaml|
|FRONTMAN_SERVICES_FILE	|The path to the services file	|services.yaml|
|FRONTMAN_REDIS_NAMESPACE|	The namespace used to prefix all Redis keys	|frontman|
|FRONTMAN_REDIS_URI	|The URI of the Redis instance used for storing backend service configuration	|redis://localhost:6379|
|FRONTMAN_API_ADDR	|The address and port on which the API should listen for incoming requests|	0.0.0.0:8080|
|FRONTMAN_API_SSL_ENABLED|	Whether or not the API should use SSL/TLS encryption|	false|
|FRONTMAN_API_SSL_CERT	|The path to the API SSL/TLS certificate file||	
|FRONTMAN_API_SSL_KEY|	The path to the API SSL/TLS key file	||
|FRONTMAN_GATEWAY_ADDR	|The address and port on which the gateway should listen for incoming requests|	0.0.0.0:8000|
|FRONTMAN_GATEWAY_SSL_ENABLED|	Whether or not the gateway should use SSL/TLS encryption|	false|
|FRONTMAN_GATEWAY_SSL_CERT|	The path to the gateway SSL/TLS certificate file||	
|FRONTMAN_GATEWAY_SSL_KEY|	The path to the gateway SSL/TLS key file	||
|FRONTMAN_LOG_LEVEL|	The log level to use	|info|

#### Frontman Configuration File

This describes the structure and options of the Frontman configuration file. This file is used to configure the Frontman API gateway and services.

The configuration file is written in YAML format and is structured as follows:
```yaml
global:
  service_type: SERVICE_TYPE
  services_file: SERVICES_FILE
  redis_namespace: REDIS_NAMESPACE
  redis_uri: REDIS_URI

api:
  addr: API_ADDR
  ssl:
    enabled: API_SSL_ENABLED
    cert: API_SSL_CERT
    key: API_SSL_KEY

gateway:
  addr: GATEWAY_ADDR
  ssl:
    enabled: GATEWAY_SSL_ENABLED
    cert: GATEWAY_SSL_CERT
    key: GATEWAY_SSL_KEY

logging:
  level: LOG_LEVEL

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
|ssl.enabled|	Whether or not the API should use SSL/TLS encryption.|	false|
|ssl.cert|	The path to the API SSL/TLS certificate file.||	

#### Gateway Section
The gateway section contains configuration options for the Frontman Gateway.

|Key| Description|Default Value|
|:--:|:---:|:---:|
|addr|	The address on which the Frontman Gateway will listen.	|0.0.0.0:8000|
|ssl.enabled|	Whether or not the Gateway should use SSL/TLS encryption.|	false|
|ssl.cert|	The path to the Gateway SSL/TLS certificate file.||	

#### Logging Section
The logging section contains configuration options for the Frontman logging.

|Key| Description|Default Value|
|:--:|:---:|:---:|
|level|	The log level of the Frontman logging. Valid options are debug, info, warn, error, and fatal.|	info

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
$ docker run -p 8080:8080 hyperioxx/frontman:latest
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