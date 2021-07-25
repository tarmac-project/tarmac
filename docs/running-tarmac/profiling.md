---
description: Profiling a running Tarmac instance
---

# Troubleshooting Performance Issues

Troubleshooting Performance Issues or Memory leaks within running services can be a complex task. Luckily, Tarmac uses the native Go tooling to create profiling snapshots of a running instance.

PProf is a Go tool for capturing and visualizing profiling data. Tarmac uses the `net/http/pprof` package to make PProf available via HTTP end-points.

By default, all PProf end-points are disabled, preventing unauthorized use of PProf (which itself can affect performance). To enable PProf, set the Configuration value of `enable-pprof` to `true`. Using a distributed configuration service such as Consul, users can change this value live without restarting the application instance.

Follow the [Configuration guide](configuration.md) for more details on configuring Tarmac.

Once enabled, users can use the following end-points to capture profiling data.

| URI | Description |
|---- | ----------- |
| `/debug/pprof` | PProf Index linking to individual profiling pages |
| `/debug/pprof/allocs` | A sampling of all past memory allocations |
| `/debug/pprof/block` | Stack traces that led to blocking on synchronization primitives |
| `/debug/pprof/cmdline` | The command line invocation of the current program |
| `/debug/pprof/goroutine` | Stack traces of all current goroutines |
| `/debug/pprof/heap` | A sampling of memory allocations of live objects |
| `/debug/pprof/mutex` | Stack traces of holders of contended mutexes |
| `/debug/pprof/profile` | CPU Profile |
| `/debug/pprof/threadcreate` | Stack traces that led to the creation of new OS threads |
| `/debug/pprof/trace` | A trace of execution of the current program |

More information about PProf can be found via the [official documentation](https://pkg.go.dev/net/http/pprof).
