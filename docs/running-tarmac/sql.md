---
description: Selecting a SQL Datastore
---

# SQL Datastore Setup

Tarmac has support for multiple SQL datastore storage systems. These datastores can change with basic 
configuration options within Tarmac. As a WASM Function developer, you do not need to know the underlying datastore 
when writing the function. Callbacks for accessing the SQL datastore are generic across all supported datastores.

To start using a SQL datastore, set the `enable_sql` configuration to `true` and specify which supported 
platform to use with the `sqlstore_type` variable.

The below table outlines the different available options.

| Datastore | Type option | Description | Useful for |
| --------  | ----------- | ----------- | ---------- |
| MySQL | `mysql` | MySQL a widely used, open-source RDBMS | Strong Consistency, Well Known, Persistent Storage, Scales Well |
| PostgreSQL | `postgres` | PostgreSQL a widely used, open-source RDBMS | Strong Consistency, Well Known, Persistent Storage, Scales Well |

For more detailed configuration options, check out the [Configuration](configuration.md) documentation.
