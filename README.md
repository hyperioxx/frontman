![Logo](/assets/logo.png)
# Frontman
Frontman is an open-source API gateway written in Go that allows you to manage your microservices and expose them as a single API endpoint. It acts as a reverse proxy and handles requests from clients, routing them to the appropriate backend service.

Frontman provides a simple, flexible, and scalable solution for managing microservices. It includes features such as and automatic health checks, making it easy to manage and maintain your API gateway and backend services.

With Frontman, you can easily add new backend services, update existing ones, and remove them, without affecting the clients. Frontman also provides an HTTP API for managing backend services, making it easy to automate the management of your API gateway.

Frontman is designed to be highly available and fault-tolerant. It uses Redis as a data store for managing backend services, ensuring that your services are always available and up-to-date. It also supports distributed deployments, allowing you to scale horizontally as your traffic and service requirements grow.

Overall, Frontman is a powerful and flexible API gateway that simplifies the management of microservices, making it easier for developers to build and maintain scalable and reliable API endpoints.

<p>&nbsp;</p>

[![Go Report Card](https://goreportcard.com/badge/github.com/Frontman-Labs/frontman)](https://goreportcard.com/report/github.com/Frontman-Labs/frontman) [![GitHub license](https://img.shields.io/github/license/Frontman-Labs/frontman)](https://github.com/Frontman-Labs/frontman/blob/main/LICENCE) ![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/Frontman-Labs/frontman)
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
  - [Frontman Plugins](#frontman-plugins)
  - [Contributing](#contributing)
  - [License](#license)

## Features
- Reverse proxy requests to backend services
- Dynamic backend service configuration using Yaml, Redis or MongoDB as a backend database
- Automatic refresh of connections to upstream targets with configurable timeouts and maximum idle connections
- TLS encryption for secure communication with clients
- Option to strip the service path from requests before forwarding to upstream targets
- Written in Go for efficient performance and concurrency support
- Plugin system to modify requests and responses before they are sent to or received from the backend service
  
## Usage
### Configuration

#### Env Variables
Frontman is configured using environment variables. The following variables are supported:
|Environment Variable| Description| Default|
|:--------------------:|:--------------:|:--------------:|
|FRONTMAN_SERVICE_TYPE	|The service type to use|`yaml`|
|FRONTMAN_SERVICES_FILE	|The path to the services file|`services.yaml`|
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
|FRONTMAN_GATEWAY_SSL_KEY|	The path to the gateway SSL/TLS key file||
|FRONTMAN_LOG_LEVEL|	The log level to use|`info`|

#### Frontman Configuration File

This describes the structure and options of the Frontman configuration file. This file is used to configure the Frontman API gateway and services.

The configuration file is written in YAML format and is structured as follows:
```yaml
global:
  service_type: "my_service"
  services_file: "services.yaml"
  redis_uri: "redis://localhost:6379"
  redis_namespace: "my_service"
  mongo_uri: "mongodb://localhost:27017"
  mongo_db_name: "my_db"
  mongo_collection_name: "my_collection"
api:
  addr: ":8080"
  ssl:
    enabled: true
    cert: "/path/to/cert.pem"
    key: "/path/to/key.pem"
gateway:
  addr: ":8081"
  ssl:
    enabled: false
logging:
  level: "debug"
plugins:
  enabled: true
  order:
    - "/path/to/plugin1.so"
    - "/path/to/plugin2.so"

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
|mongo_db_name|	is a string representing the name of the MongoDB database where the backend services will be stored.|`frontman`|
|mongo_collection_name|	is a string representing the name of the MongoDB collection where the backend services will be stored.|	`services`|

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
$ git clone https://github.com/Frontman-Labs/frontman.git
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

To add a backend service, send a POST request to /api/services with the backend service configuration in the request body. To update a backend service, send a PUT request to /api/services/{name} with the updated backend service configuration in the request body. To remove a backend service, send a DELETE request to /api/services/{name}.

Once backend services have been defined and stored in the service registry, Frontman can be configured to use that service registry by specifying the appropriate service type and configuration parameters in the global config file.

#### Running Frontman in Docker
Frontman can also be run as a Docker container. Here's an example command to start Frontman in Docker, assuming your Redis instance is running on localhost:

```bash
$ docker run -p 8080:8080 hyperioxx/frontman:latest
```
This command starts a new container, maps port 8080 on the host to port 8080 in the container

## Managing Backend Services
Frontman uses various storage systems to manage backend service configurations. Backend services are represented as JSON objects and stored in a data storage system. Here's an example of a backend service configuration:


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
      "weights": []
    }
  }
}
```
Supported Load Balancer types:
- Round-robin
  ```json
    "loadBalancerPolicy": { 
    "type": "round_robin"
    }
  ```
- Weighted Round-robin
  ```json
  "loadBalancerPolicy": { 
    "type": "weighted_round_robin",
    "options": {
      "weights": [1, 2, 3]
    }
  ```

You can add, update, and remove backend services using the following REST endpoints:

- GET /services - Retrieves a list of all backend services
- GET /health - Performs a health check on all backend services and returns the status of each service
- POST /services - Adds a new backend service
- PUT /services/{name} - Updates an existing backend service
- DELETE /services/{name} - Removes a backend service

## Frontman Plugins

Frontman allows you to create custom plugins that can be used to extend its functionality. Plugins are implemented using the FrontmanPlugin interface, which consists of three methods:

- Name(): returns the name of the plugin.
- PreRequest(): is called before sending the request to the target service. It takes in the original request, a ServiceRegistry, and a Config. An error is returned if the plugin encounters any issues.
- PostResponse(): is called after receiving the response from the target service. It takes in the response, a ServiceRegistry, and a Config. An error is returned if the plugin encounters any issues.
- Close(): is called when the plugin is being shut down. An error is returned if the plugin encounters any issues.

To create a plugin, implement the FrontmanPlugin interface in a Go package. Then, add the plugin to the Frontman configuration file.

Here is an example of a simple Frontman plugin:

```go
type FrontmanPlugin struct {}

func (p *FrontmanPlugin) Name() string {
    return "Example Plugin"
}

func (p *FrontmanPlugin) PreRequest(req *http.Request, sr service.ServiceRegistry, cfg config.Config) error {
    // Modify the request before sending it to the target service
    return nil
}

func (p *FrontmanPlugin) PostResponse(resp *http.Response, sr service.ServiceRegistry, cfg config.Config) error {
    // Modify the response before sending it back to the client
    return nil
}

func (p *FrontmanPlugin) Close() error {
    // Cleanup resources used by the plugin
    return nil
}
```

To compile a Go plugin, you first need to implement the FrontmanPlugin interface in a Go package. Once your plugin is implemented, you can compile it using the following command:

```bash
go build -buildmode=plugin -o example.so example.go
```

This will produce a shared library named example.so that can be loaded dynamically at runtime.

After compiling the plugin, you need to update the Frontman configuration file to load the plugin. In the PluginConfig section of the configuration file, set enabled to true and add the path to the plugin library to the order array. For example:

```
plugin:
  enabled: true
  order:
    - "/path/to/example.so"
```

In this example, the plugin library is located at /path/to/example.so. If you have multiple plugins, you can specify the order in which they should be loaded by adding their paths to the order array.

Once you have updated the configuration file, restart Frontman to load the new plugins. Your plugin should now be loaded and its PreRequest and PostResponse methods will be called for each request.

## Contributing
If you'd like to contribute to Frontman, please fork the repository and submit a pull request. We welcome bug reports, feature requests, and code contributions.

## License
Frontman is released under the GNU General Public License. See LICENSE for details.