# Tarmac - Microservice Framework for WASM

[![PkgGoDev](https://pkg.go.dev/badge/github.com/madflojo/tarmac)](https://pkg.go.dev/github.com/madflojo/tarmac)
[![Go Report Card](https://goreportcard.com/badge/github.com/madflojo/tarmac)](https://goreportcard.com/report/github.com/madflojo/tarmac)

Tarmac is a framework for building distributed services for any language. Like many other distributed service frameworks or microservice toolkits, Tarmac abstracts the complexities of building distributed systems. Except unlike other toolkits, Tarmac is language agnostic.

Using Web Assembly (WASM), Tarmac users can write their logic in many different languages like Rust, Go, Javascript, or even C and run it using the same framework.

## Tarmac vs. Functions as a Service

Like other FaaS services, it's easy to use Tarmac to create either Functions or full Microservices. However, unlike FaaS platforms, Tarmac provides users with much more than an easy way to run a function inside a Docker container.

By leveraging Web Assembly System Interface (WASI), Tarmac creates an isolated environment for running functions. But unlike FaaS platforms, Tarmac users will be able to import Tarmac functions.

Like any other microservice framework, Tarmac will handle the complexities of Database Connections, Caching, Metrics, and Dynamic Configuration. Users can focus purely on the function logic and writing it in their favorite programming language.

Tarmac aims to enable users to create robust and performant distributed services with the ease of writing serverless functions with the convenience of a standard microservices framework.

## Not ready for Production

At the moment, Tarmac is Experimental, and interfaces will change; new features will come. But if you are interested in WASM and want to write a simple Microservice, Tarmac will work.

Of course, Contributions are also welcome.

## Getting Started with Tarmac

At the moment, Tarmac can run WASI-generated WASM code. However, the serverless function code must follow a pre-defined function signature with Tarmac executing an `HTTPHandler()` function.

To understand this better, take a look at our simple example written in Go (found in [example/go](example/go)).

```golang
package main

import (
        "encoding/base64"
        "fmt"
        "os"
)

//export HTTPHandler
func HTTPHandler() int {
        d, err := base64.StdEncoding.DecodeString(os.Getenv("HTTP_PAYLOAD"))
        if err != nil {
                fmt.Fprintf(os.Stderr, "Invalid Payload")
                return 400
        }
        if os.Getenv("HTTP_METHOD") == "POST" || os.Getenv("HTTP_METHOD") == "PUT" {
                fmt.Fprintf(os.Stdout, "%s", d)
        }
        return 200
}

func main() {}
```

As we can see from the above code Tarmac passes the HTTP Context and Payload through Environment Variables (at the moment). The Payload is Base64 encoded but otherwise untouched.

To compile the example above simply run:

```shell
$ tinygo build -o tarmac_module.wasm -target wasi ./main.go
```

Once compiled users can run Tarmac using the following command:

```shell
$ docker run -v ./path/to/wasm-module:/wasm-module -e "APP_WASM_MODULE=/wasm-module/target_module.wasm" madflojo/tarmac
```

Do pay attention to the volume mount and the `APP_WASM_MODULE` environment variable, as these are key to specifying what WASM module to execute.

### Configuring Tarmac


Tarmac supports multiple configuration sources from Environment Variables, a JSON file, or using HashiCorp Consul. All of these configuration options can also exist together to provide both static and dynamic configuration.

When using Environment Variables, all configurations are prefixed with `APP_`. The list below will show both Environment and Consul/JSON format for configuration.

| Environment Variable | Consul/JSON | Type | Description |
|----------------------|-------------|------|-------------|
| `APP_ENABLE_TLS` | `enable_tls` | bool | Enable the HTTPS Listener (default: `True`) |
| `APP_LISTEN_ADDR` | `listen_addr` | string | Define the HTTP/HTTPS Listener address (default: `0.0.0.0:8443`) |
| `APP_CONFIG_WATCH_INTERVAL` | `config_watch_interval` | integer | Frequency in seconds which Consul configuration will be refreshed (default: 15) |
| `APP_USE_CONSUL` | `use_consul` | bool | Enable Consul based configuration (default: `False`) |
| `APP_CONSUL_ADDR` | `consul_addr` | string | Consul address (i.e. `consul.example.com:8500`) |
| `APP_CONSUL_KEYS_PREFIX` | `consul_keys_prefix` | string | Key path for app specific consul configuration |
|| `from_consul` | bool | Indicator to reflect whether Consul config was loaded |
| `APP_DEBUG` | `debug` | bool | Enable debug logging |
| `APP_TRACE` | `trace` | bool | Enable trace logging | 
| `APP_DISABLE_LOGGING` | `disable_logging` | bool | Disable all logging |
| `APP_DB_SERVER` | `db_server` | string | Database server address |
| `APP_DB_PASSWORD` | `db_password` | string | Database password | 
| `APP_CERT_FILE` | `cert_file` | string | Certificate File Path (i.e. `/some/path/cert.crt`) |
| `APP_KEY_FILE` | `key_file` | string | Key File Path (i.e. `/some/path/cert.key`) 

#### Consul Format

When using Consul the `consul_keys_prefix` should be the path to a key with a JSON string as the value. For example, a key of `tarmac/config` will have a value of `{"from_consul":true}`.

![](static/img/consul-example.png)
