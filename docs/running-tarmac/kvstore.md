---
description: Selecting a KV Store
---

# Key:Value Datastore Setup

Tarmac has support for multiple Key:Value datastore storage systems. These datastores can change with basic 
configuration options within Tarmac. As a WASM Function developer, you do not need to know the underlying datastore 
when writing the function. Callbacks for accessing the Key:Value datastore are generic across all supported datastores.

To start using a Key:Value datastore, set the `enable_kvstore` configuration to `true` and specify which supported 
platform to use with the `kvstore_type` variable.

The below table outlines the different available options.

| Datastore | Type option | Description | Useful for |
| --------  | ----------- | ----------- | ---------- |
| In-Memory | `in-memory` | In-Memory key/value store | Testing, Development, Non-Persistent Caching |
| BoltDB | `boltdb` | BoltDB Embedded key/value store | Strong Consistency, Persistent Storage |
| Redis | `redis` | Redis including Sentinel and Enterprise capabilities | Strong Consistency, Fast Reads and Writes, Non-Persistent storage  |
| Cassandra | `cassandra` | Cassandra including TLS connectivity | Eventual Consistency, Persistent Storage, Large sets of data |

For more detailed configuration options, check out the [Configuration](configuration.md) documentation.
