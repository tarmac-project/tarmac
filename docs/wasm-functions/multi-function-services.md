---
description: Configure multi-function services as well as multi-service instances.
---

# Multi-Function Services

Configuring Multi-Function Services is a core capability of Tarmac. To do so, we will create a `tarmac.json` file that defines Functions and Routes to expose those functions.

By default, Tarmac looks within the `/functions/` directory for a `tarmac.json`; however this can be overridden using the `WASM_FUNCTION_CONFIG` configuration parameter.

The `tarmac.json` file has a simple structure that consists of a single object with a `services` key. The `services` key defines each service and its corresponding functions and routes.

```json
{
  "services": {
    "my-service": {
      "name": "my-service",
      "functions": {
        "function1": {
          "filepath": "/path/to/function1.wasm",
          "pool_size": 10
        },
        "function2": {
          "filepath": "/path/to/function2.wasm"
        }
      },
      "routes": [
        {
          "type": "http",
          "path": "/function1",
          "methods": ["GET"],
          "function": "function1"
        },
        {
          "type": "http",
          "path": "/function2",
          "methods": ["POST"],
          "function": "function2"
        }
      ]
    }
  }
}
```

## Configuration Options

### Services

The `services` object contains one or more key-value pairs, with each key representing the name of a service.

Each service object should include the following properties:

- `name`: The name of the service (required).
- `functions`: An object containing the functions for the service (required).
- `routes`: An array of objects defining the routes for the service (required).

#### Functions

The functions object contains one or more key-value pairs, with each key representing the name of a function.

Each function object should include the following properties:

- `filepath`: The file path to the .wasm file containing the function code (required).
- `pool_size`: The number of instances of the function to create (optional). Defaults to 100.

#### Routes

The "routes" property in the `tarmac.json` configuration file defines the endpoints (HTTP or scheduled task) of the service and maps them to their respective functions.

##### HTTP Routes

The routes array in the `tarmac.json` configuration file defines the HTTP endpoints for your service.

Each route object contains the following properties:

- `type` (required): For HTTP routes, set this to `http`.
- `path` (required): The URL path for the endpoint.
- `methods` (required): An array of HTTP methods that the endpoint supports (i.e. `GET`, `POST`, `PUT`, `DELETE`).
- `function` (required): The function to call when the endpoint receives requests.

Here is an example of a route object that defines an HTTP endpoint that responds to GET requests on the root path and calls the default function:

```json
{
  "type": "http",
  "path": "/",
  "methods": ["GET"],
  "function": "default"
}

```

You can define multiple HTTP routes in the routes array.

##### Scheduled Tasks

In addition to HTTP endpoints, Tarmac also supports scheduled tasks.

You can define a scheduled task route by adding a route object with the following properties to the routes array:

- `type` (required): For Schedule Tasks, set to `scheduled_task`.
- `function` (required): The function to call when the task is executed.
- `frequency` (required): The frequency in seconds to execute the function.

Here is an example of a route object that defines a scheduled task that executes the default function every 15 seconds:

```json
{
  "type": "scheduled_task",
  "function": "default",
  "frequency": 15
}
```

You can define multiple scheduled tasks in the routes array.

##### Init Functions

In addition to HTTP and scheduled task routes, Tarmac also supports init functions.

You can define an init function route by adding a route object with the following properties to the routes array:

- `type` (required): For Init Functions, set to `init`.
- `function` (required): The function to call when the service is initialized.
- `retries` (optional): The number of times to retry the function if it fails. Defaults to 0.
- `frequency` (optional): The frequency in seconds to retry the function if it fails. Exponential backoff is used. Defaults to 1.

Here is an example of a route object that defines an init function that executes the default function when the service is initialized:

```json
{
  "type": "init",
  "function": "default"
}
```

You can define multiple init functions in the routes array. Functions will be executed before the server is fully started but after the WASM modules are loaded and callbacks are registered.

##### Functions

Tarmac supports the ability for Functions to call other Functions using the Function to Function route. 

You can define a function route by adding a route object with the following properties to the route array.

- `type` (required): For Function to Function routes, set to `function`.
- `function` (required): The function to call when executed.

Here is an example of a route object that defines the "function1" function.

```json
{
  "type": "function",
  "function": "function1"
}
```
