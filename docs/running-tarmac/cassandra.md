---
description: Configuring Tarmac with Cassandra
---

# Cassandra Configuration

This page contains Cassandra specific configuration to utilize Cassandra with Tarmac.

Tarmac supports multiple configuration sources such as Environment Variables, JSON files, or using HashiCorp Consul. All of these configuration options can also exist together to provide both static and dynamic configurations.

When using Environment Variables, all configurations are prefixed with `APP_`. The list below will show both Environment and Consul/JSON format for configuration.

| Environment Variable | Consul/JSON | Type | Description |
| :--- | :--- | :--- | :--- |
| `APP_CASSANDRA_HOSTS` | `cassandra_hosts` | `[]string` | Cassandra node addresses |
| `APP_CASSANDRA_PORT` | `cassandra_port` | `int` | Cassandra node port |
| `APP_CASSANDRA_KEYSPACE` | `cassandra_keyspace` | `string` | Cassandra Keyspace name |
| `APP_CASSANDRA_CONSISTENCY` | `cassandra_consistency` | `string` | Desired Consistency (Default: `Quorum`)|
| `APP_CASSANDRA_REPL_STRATEGY` | `cassandra_repl_strategy` | `string` | Replication Strategy for Cluster (Default: `SimpleStrategy`)|
| `APP_CASSANDRA_REPLICAS` | `cassandra_replicas` | `int` | Default number of replicas for data (Default: `1`) |
| `APP_CASSANDRA_USER` | `cassandra_user` | `string` | Username to authenticate with |
| `APP_CASSANDRA_PASSWORD` | `cassandra_password` | `string` | Password to authenticate with |
| `APP_CASSANDRA_HOSTNAME_VERIFY` | `cassandra_hostname_verify` | `bool` | Enable/Disable hostname verification for TLS |

## Consul Format

When using Consul the `consul_keys_prefix` should be the path to a key with a JSON `string` as the value. For example, a key of `tarmac/config` will have a value of `{"from_consul":true}`.

![](../.gitbook/assets/consul-example.png)

