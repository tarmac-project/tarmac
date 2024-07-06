---
description: Configuring Tarmac
---

# Configuration

Tarmac supports multiple configuration sources such as Environment Variables, JSON files, or using HashiCorp Consul. All of these configuration options can also exist together to provide both static and dynamic configurations.

When using Environment Variables, all configurations are prefixed with `APP_`. The list below will show both Environment and Consul/JSON format for configuration.

| Environment Variable | Consul/JSON | Type | Description |
| :--- | :--- | :--- | :--- |
| `APP_ENABLE_TLS` | `enable_tls` | `bool` | Enable the HTTPS Listener \(default: `True`\) |
| `APP_LISTEN_ADDR` | `listen_addr` | `string` | Define the HTTP/HTTPS Listener address \(default: `0.0.0.0:8443`\) |
| `APP_CONFIG_WATCH_INTERVAL` | `config_watch_interval` | `int` | Frequency in seconds which Consul configuration will be refreshed \(default: `15`\) |
| `APP_USE_CONSUL` | `use_consul` | `bool` | Enable Consul based configuration \(default: `False`\) |
| `APP_CONSUL_ADDR` | `consul_addr` | `string` | Consul address \(i.e. `consul.example.com:8500`\) |
| `APP_CONSUL_KEYS_PREFIX` | `consul_keys_prefix` | `string` | Key path for app specific consul configuration |
|  | `from_consul` | `bool` | Indicator to reflect whether Consul config was loaded |
| `APP_DEBUG` | `debug` | `bool` | Enable debug logging |
| `APP_TRACE` | `trace` | `bool` | Enable trace logging |
| `APP_DISABLE_LOGGING` | `disable_logging` | `bool` | Disable all logging |
| `APP_CERT_FILE` | `cert_file` | `string` | Certificate File Path \(i.e. `/some/path/cert.crt`\) |
| `APP_KEY_FILE` | `key_file` | `string` | Key File Path \(i.e. `/some/path/cert.key`\) |
| `APP_CA_FILE` | `ca_file` | `string` | Certificate Authority Bundle File Path \(i.e `/some/path/ca.pem`\). When defined, enables mutual-TLS authentication |
| `APP_IGNORE_CLIENT_CERT` | `ignore_client_cert` | `string` | When defined will disable Client Cert validation for m-TLS authentication |
| `APP_WASM_FUNCTION` | `wasm_function` | `string` | Path and Filename of the WASM Function to execute \(Default: `/functions/tarmac.wasm`\) |
| `APP_WASM_FUNCTION_CONFIG` | `wasm_function_config` | `string` | Path to Service configuration for multi-function services \(Default: `/functions/tarmac.json`\) |
| `APP_WASM_POOL_SIZE` | `wasm_pool_size` | `int` | Number of WASM function instances to create \(Default: `100`\). Only applicable when `wasm_function` is used. |
| `APP_ENABLE_PPROF` | `enable_pprof` | `bool` | Enable PProf Collection HTTP end-points |
| `APP_ENABLE_KVSTORE` | `enable_kvstore` | `bool` | Enable the KV Store |
| `APP_KVSTORE_TYPE` | `kvstore_type` | `string` | Select KV Store to use (Options: `redis`, `cassandra`, `boltdb`, `in-memory`, `internal`)|
| `APP_ENABLE_SQL` | `enable_sql` | `bool` | Enable the SQL Store |
| `APP_SQL_TYPE` | `sql_type` | `string` | Select SQL Store to use (Options: `postgres`, `mysql`)|
| `APP_RUN_MODE` | `run_mode` | `string` | Select the run mode for Tarmac (Options: `daemon`, `job`). Default: `daemon`. The `job` option will cause Tarmac to exit after init functions are executed. |
| `APP_ENABLE_MAINTENANCE_MODE` | `enable_maintenance_mode` | `bool` | Enable Maintenance Mode. When enabled, Tarmac will return a 503 for requests to `/ready` allowing the service to go into "maintenance mode". |

## Consul Format

When using Consul the `consul_keys_prefix` should be the path to a key with a JSON `string` as the value. For example, a key of `tarmac/config` will have a value of `{"from_consul":true}`.

![](../.gitbook/assets/consul-example.png)

