![Logo](/assets/logo.png)
# Frontman
Frontman is an open-source API gateway written in Go that allows you to manage your microservices and expose them as a single API endpoint. It acts as a reverse proxy and handles requests from clients, routing them to the appropriate backend service.

Frontman provides a simple, flexible, and scalable solution for managing microservices. It includes features such as and automatic health checks, making it easy to manage and maintain your API gateway and backend services.

With Frontman, you can easily add new backend services, update existing ones, and remove them, without affecting the clients. Frontman also provides an HTTP API for managing backend services, making it easy to automate the management of your API gateway.

Frontman is designed to be highly available and fault-tolerant. It uses Redis as a data store for managing backend services, ensuring that your services are always available and up-to-date. It also supports distributed deployments, allowing you to scale horizontally as your traffic and service requirements grow.

Overall, Frontman is a powerful and flexible API gateway that simplifies the management of microservices, making it easier for developers to build and maintain scalable and reliable API endpoints.

<p>&nbsp;</p>

[![Go Report Card](https://goreportcard.com/badge/github.com/hyperioxx/frontman)](https://goreportcard.com/report/github.com/hyperioxx/frontman) [![GitHub license](https://img.shields.io/github/license/hyperioxx/frontman)](https://github.com/hyperioxx/frontman/blob/main/LICENCE) ![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/Hyperioxx/frontman)
<br />

## Glossary

- [Frontman](#frontman)
	- [Glossary](#glossary)
	- [Features](#features)
	- [Usage](#usage)
		- [Configuration](#configuration)
			- [Env Variables](#env-variables)
			- [Frontman Configuration File](#frontman-configuration-file)
			- [Global Section](#global-section)
			- [API Section](#api-section)
			- [Gateway Section](#gateway-section)
			- [Logging Section](#logging-section)
		- [Starting Frontman](#starting-frontman)
				- [Building from Source](#building-from-source)
			- [Running Frontman Locally](#running-frontman-locally)
			- [Running Frontman in Docker](#running-frontman-in-docker)
	- [Managing Backend Services](#managing-backend-services)
	- [Contributing](#contributing)
	- [License](#license)

## Features
- Reverse proxy requests to backend services
- Dynamic backend service configuration using Yaml, Redis or MongoDB as a backend database
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
|FRONTMAN_SERVICE_TYPE	|The service type to use|	`yaml`|
|FRONTMAN_SERVICES_FILE	|The path to the services file	|`services.yaml`|
|FRONTMAN_REDIS_NAMESPACE|	The namespace used to prefix all Redis keys	|`frontman`|
|FRONTMAN_REDIS_URL	|The URL of the Redis instance used for storing service registry |`redis://localhost:6379`|
|FRONTMAN_MONGO_URL	|The URL of the Mongo instance used for storing service registry |`mongodb://localhost:27017`|
|FRONTMAN_MONGO_DB_NAME	| The name of the MongoDB database where the service registry is stored. |`frontman`|
|FRONTMAN_MONGO_COLLECTION_NAME	| The name of the MongoDB collection where the service registry is stored. |`services`|
|FRONTMAN_API_ADDR	|The address and port on which the API should listen for incoming requests|	`0.0.0.0:8080`|
|FRONTMAN_API_SSL_ENABLED|	Whether or not the API should use SSL/TLS encryption|	`false`|
|FRONTMAN_API_SSL_CERT	|The path to the API SSL/TLS certificate file||	
|FRONTMAN_API_SSL_KEY|	The path to the API SSL/TLS key file	||
|FRONTMAN_GATEWAY_ADDR	|The address and port on which the gateway should listen for incoming requests|	`0.0.0.0:8000`|
|FRONTMAN_GATEWAY_SSL_ENABLED|	Whether or not the gateway should use SSL/TLS encryption|	`false`|
|FRONTMAN_GATEWAY_SSL_CERT|	The path to the gateway SSL/TLS certificate file||	
|FRONTMAN_GATEWAY_SSL_KEY|	The path to the gateway SSL/TLS key file	||
|FRONTMAN_LOG_LEVEL|	The log level to use	|`info`|

#### Frontman Configuration File

This describes the structure and options of the Frontman configuration file. This file is used to configure the Frontman API gateway and services.

The configuration file is written in YAML format and is structured as follows:
```yaml
global:
  service_type: SERVICE_TYPE
  services_file: SERVICES_FILE
  redis_uri: REDIS_URI
  redis_namespace: REDIS_NAMESPACE
  mongo_uri: MONGO_URI
  mongo_db_name: MONGO_DB_NAME
  mongo_collection_name: MONGO_COLLECTION_NAME
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

|Key| Description|Default Value|
|:--:|:---:|:---:|
|service_type|	The type of service registry used to store backend services. Valid options are yaml and redis.|	`yaml`|
|services_file|	The path to the YAML file used to store backend services when using the yaml service registry.|	`services.yaml`|
|redis_namespace|	The namespace used to prefix all Redis keys when using the redis service registry.|	frontman|
|redis_uri|is a string representing the URI of the Redis server that the application will use to store and retrieve backend services data. |`redis://localhost:6379`|
|mongo_uri|	is a string representing the URI of the MongoDB server that the application will use to store and retrieve backend services data.|`mongodb://localhost:27017`|
mongo_db_name|	is a string representing the name of the MongoDB database where the backend services will be stored.|`frontman`|
mongo_collection_name|	is a string representing the name of the MongoDB collection where the backend services will be stored.|	`services`|

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

After downloading or building Frontman, you can run it locally by providing a configuration file with the -config flag. The configuration file specifies the configuration options. Before running Frontman, make sure to review the configuration file to understand the options and backend service configurations.

To start Frontman, navigate to the directory containing the binary and run the following command:

```bash
$ ./frontman -config /path/to/config.yml
```

Replace /path/to/config.yml with the actual path to your configuration file. This command will start Frontman with the specified configuration options. You can then make requests to Frontman's API or Gateway at the specified endpoints to have them forwarded to the configured backend services.

Frontman does not provide a way to configure backend services via the configuration file. Instead, you can use Frontman's API to add, update, or remove backend services dynamically. The API allows you to manage backend services in real-time without having to restart Frontman.

To add a backend service, send a POST request to /api/backend-services with the backend service configuration in the request body. To update a backend service, send a PUT request to /api/backend-services/{name} with the updated backend service configuration in the request body. To remove a backend service, send a DELETE request to /api/backend-services/{name}.

Once backend services have been defined and stored in the service registry, Frontman can be configured to use that service registry by specifying the appropriate service type and configuration parameters in the global config file.

#### Running Frontman in Docker
Frontman can also be run as a Docker container. Here's an example command to start Frontman in Docker, assuming your Redis instance is running on localhost:

```bash
$ docker run -p 8080:8080 hyperioxx/frontman:latest
```
This command starts a new container, maps port 8080 on the host to port 8080 in the container

## Managing Backend Services
Frontman uses various storage systems to manage backend service configurations. Backend services are represented as JSON objects and stored in a data storage system. Here's an example of a backend service configuration:

**Note:** *The loadBalancerPolicy configuration option is currently not active, but the option exists for future implementation of load balancing policies.*

```json
{
  "name": "test_service",
  "scheme": "http",
  "upstreamTargets": [
    "http://localhost:8080"
  ],
  "path": "/api/test",
  "domain": "localhost",
  "healthCheck": "/health",
  "retryAttempts": 3,
  "timeout": 10,
  "maxIdleConns": 100,
  "maxIdleTime": 30,
  "stripPath": true,
  "loadBalancerPolicy": { 
    "type": "",
    "options": {
      "weighted": {}
    }
  }
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