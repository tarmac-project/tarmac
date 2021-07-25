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
| `APP_WASM_FUNCTION` | `wasm_function` | `string` | Path and Filename of the WASM Function to execute \(Default: `/functions/tarmac.wasm`\) |
| `APP_ENABLE_PPROF` | `enable_pprof` | `bool` | Enable PProf Collection HTTP end-points |
| `APP_ENABLE_KVSTORE` | `enable_kvstore` | `bool` | Enable the KV Store |
| `APP_KVSTORE_TYPE` | `kvstore_type` | `string` | Select KV Store to use (Options: `redis`, `cassandra`)|
| `APP_REDIS_SERVER` | `redis_server` | `string` | Redis server address |
| `APP_REDIS_DATABASE` | `redis_database` | `int` | Redis Database (default: `0`) |
| `APP_REDIS_PASSWORD` | `redis_password` | `string` | Redis password |
| `APP_REDIS_SENTINEL_SERVERS` | `redis_sentinel_servers` | `[]string` | Redis Sentinel Server Addresses |
| `APP_REDIS_SENTINEL_MASTER` | `redis_sentinel_master` | `string` | Redis Sentinel Master Instance Name |
| `APP_REDIS_CONNECT_TIMEOUT` | `redis_connect_timeout` | `int` | Redis Connection Timeout in seconds |
| `APP_REDIS_HOSTNAME_VERIFY` | `redis_hostname_verify` | `bool` | Skip hostname verification for TLS |
| `APP_REDIS_KEEPALIVE` | `redis_keepalive` | `int` | TCP Keepalive Interval in seconds (Default: `300`) |
| `APP_REDIS_MAX_ACTIVE` | `redis_max_active` | `int` | Max Active Connections |
| `APP_REDIS_READ_TIMEOUT` | `redis_read_timeout` | `int` | Read timeout in seconds |
| `APP_REDIS_WRITE_TIMEOUT` | `redis_write_timeout` | `int` | Write timeout in seconds |
| `APP_CASSANDRA_HOSTS` | `cassandra_hosts` | `[]string` | Cassandra node addresses |
| `APP_CASSANDRA_PORT` | `cassandra_port` | `int` | Cassandra node port |
| `APP_CASSANDRA_KEYSPACE` | `cassandra_keyspace` | `string` | Cassandra Keyspace name |
| `APP_CASSANDRA_CONSISTENCY` | `cassandra_consistency` | `string` | Desired Consistency (Default: `Quorum`)|
| `APP_CASSANDRA_REPL_STRATEGY` | `cassandra_repl_strategy` | `string` | Replication Strategy for Cluster (Default: `SimpleStrategy`)|
| `APP_CASSANDRA_REPLICAS` | `cassandra_replicas` | `int` | Default number of replicas for data (Default: `1`) |
| `APP_CASSANDRA_USER` | `cassandra_user` | `string` | Username to authenticate with |
| `APP_CASSANDRA_PASSWORD` | `cassandra_password` | `string` | Password to authenticate with |
| `APP_CASSANDRA_HOSTNAME_VERIFY` | `cassandra_hostname_verify` | `bool` | Enable/Disable hostname verification for TLS |
| | `scheduled_tasks` | `map[string]ScheduledTask` | Configured Scheduled WASM Function executions |

### Scheduled Task Definition

The below options are used to configure scheduled tasks.

| JSON | Type | Description |
| :--- | :--- | :--- |
| `interval` | `int` | Interval (in seconds) task execution should run (recurring) |
| `wasm_function` | `string` | Path and Filename of the WASM Function to execute |
| `headers` | `map[string]string` | Custom headers applied to the ServerRequest provided to WASM functions during execution |

## Consul Format

When using Consul the `consul_keys_prefix` should be the path to a key with a JSON `string` as the value. For example, a key of `tarmac/config` will have a value of `{"from_consul":true}`.

![](../.gitbook/assets/consul-example.png)

