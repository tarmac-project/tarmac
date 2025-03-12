---
description: Logging options of Tarmac
---

# Logging

Tarmac uses structured logging to provide detailed information about system operations. By default, logs are output in JSON format, which is ideal for automated log processing and analysis tools.

## Log Format

Administrators can control the format of Tarmac's logs:

- **JSON Format (Default)**: Structured logging that's machine-readable and easily parsed by log aggregation tools
- **Text Format**: Human-readable plain text format that's easier to read in the console or log files

To use text format instead of JSON, set the `text_log_format` configuration option to `true`.

## Controlling Log Levels

Administrators can control Tarmac's logging levels dynamically while using a distributed configuration service such as Consul. 
As a WASM Function developer, you can select which logging level you wish your log callbacks to use. 

As an Administrator, you can choose to enable or disable dynamically certain log levels such as Debug or Trace. To do 
this, modify the `debug` and `trace` configuration options within Consul. 

It is also possible to disable all logging by changing the `disable_logging` configuration option to `true`.

For more information on connecting Tarmac with Consul, consult our [Configuration](configuration.md) documentation.
