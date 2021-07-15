---
description: Logging options of Tarmac
---

# Controlling Log Levels

Administrators can control Tarmac's logging dynamically while using a distributed configuration service such as Consul. 
As a WASM Function developer, you can select which logging level you wish your log callbacks to use. 

As an Administrator, you can choose to enable or disable dynamically certain log levels such as Debug or Trace. To do 
this, modify the `debug` and `trace` configuration options within Consul. 

It is also possible to disable all logging by changing the `disable_logging` configuration option to `true`.

For more information on connecting Tarmac with Consul, consult our [Configuration](configuration.md) documentation.
