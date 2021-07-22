```
description: "Scheduling WASM functions as tasks"
``` 

# Scheduled Tasks

In addition to writing WASM functions to handle HTTP traffic, Tarmac developers can also create scheduled tasks which execute WASM functions.

Scheduled Tasks is a standard and powerful method of using serverless functions. As scheduled tasks in distributed systems often perform simple jobs but require quite a bit of boilerplate around ensuring they run. With Tarmac, developers can create scheduled tasks via a simple JSON configuration.

Tarmac supports multiple configuration options such as Environment Variables, JSON Config files, and Consul-based configuration. Administrators can create a JSON file similar to the below example within a `./conf` directory for Tarmac to autoload on boot.

```json
{
	"scheduled_tasks": {
		"example-task": {
			"interval": 30,
			"wasm_function": "/functions/tarmac.wasm",
			"headers": {
				"task-name": "example-task"
			}
		}
	}
}
```

The above example shows configuring a set of simple Scheduled Tasks. For additional options, consult the Configuration Guide.

## Route definition

Within WASM functions Tarmac requires users to register methods underneath a specific route. For Scheduled Tasks the route will always be defined as `scheduler:RUN`.

```go
func main() {
	// Tarmac uses waPC to facilitate WASM module execution. Modules must register their custom handlers under the
	// appropriate method as shown below.
	wapc.RegisterFunctions(wapc.Functions{
		// Register a GET request handler
		"http:GET": Count,
		// Register a POST request handler
		"http:POST": IncCount,
		// Register a PUT request handler
		"http:PUT": IncCount,
		// Register a DELETE request handler
		"http:DELETE": NoHandler,

		// Register a handler for scheduled tasks
		"scheduler:RUN": IncCount,
	})
}
```
