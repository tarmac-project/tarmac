---
description: Configuring Tarmac with BoltDB
---

# BoltDB Configuration

This page contains BoltDB specific configuration to utilize BoltDB with Tarmac.

Tarmac supports multiple configuration sources such as Environment Variables, JSON files, or using HashiCorp Consul. All of these configuration options can also exist together to provide both static and dynamic configurations.

When using Environment Variables, all configurations are prefixed with `APP_`. The list below will show both Environment and Consul/JSON format for configuration.

| Environment Variable | Consul/JSON | Type | Description |
| :--- | :--- | :--- | :--- |
| `APP_BOLTDB_FILENAME` | `boltdb_filename` | `string` | The full path and filename of the BoltDB file. If the file does not exist, it will be created. |
| `APP_BOLTDB_BUCKET` | `boltdb_bucket` | `string` | The name of the BoltDB bucket to use. If the bucket does not exist, it will be created. |
| `APP_BOLTDB_PERMISSIONS` | `boltdb_permissions` | `int` | The permissions to use when creating the BoltDB file. This is an octal value. |
| `APP_BOLTDB_TIMEOUT` | `boltdb_timeout` | `int` | The timeout in seconds to wait for BoltDB to open. |

## Consul Format

When using Consul the `consul_keys_prefix` should be the path to a key with a JSON `string` as the value. For example, a key of `tarmac/config` will have a value of `{"from_consul":true}`.

![](../.gitbook/assets/consul-example.png)

