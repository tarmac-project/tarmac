# Tasks

[![Build Status](https://travis-ci.org/madflojo/tasks.svg?branch=master)](https://travis-ci.org/madflojo/tasks) 
[![Coverage Status](https://coveralls.io/repos/github/madflojo/tasks/badge.svg?branch=master)](https://coveralls.io/github/madflojo/tasks?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/madflojo/tasks)](https://goreportcard.com/report/github.com/madflojo/tasks) 
[![Documentation](https://godoc.org/github.com/madflojo/tasks?status.svg)](http://godoc.org/github.com/madflojo/tasks)

Package tasks is an easy to use in-process scheduler for recurring tasks in Go. Tasks is focused on high frequency
tasks that run quick, and often. The goal of Tasks is to support concurrent running tasks at scale without scheduler
induced jitter.

Tasks is focused on accuracy of task execution. To do this each task is called within it's own goroutine. This ensures 
that long execution of a single invocation does not throw the schedule as a whole off track.

As usage of this scheduler scales, it is expected to have a larger number of sleeping goroutines. As it is designed to 
leverage Go's ability to optimize goroutine CPU scheduling.

For simplicity this task scheduler uses the time.Duration type to specify intervals. This allows for a simple interface 
and flexible control over when tasks are executed.

Below is an example of starting the scheduler and registering a new task that runs every 30 seconds.

```go
// Start the Scheduler
scheduler := tasks.New()
defer scheduler.Stop()

// Add a task
id, err := scheduler.Add(&tasks.Task{
  Interval: time.Duration(30 * time.Second),
  TaskFunc: func() error {
    // Put your logic here
  }(),
})
if err != nil {
  // Do Stuff
}
```

Sometimes schedules need to started at a later time. This package provides the ability to start a task only after a 
certain time. The below example shows this in practice.

```go
// Add a recurring task for every 30 days, starting 30 days from now
id, err := scheduler.Add(&tasks.Task{
  Interval: time.Duration(30 * (24 * time.Hour)),
  StartAfter: time.Now().Add(30 * (24 * time.Hour)),
  TaskFunc: func() error {
    // Put your logic here
  }(),
})
if err != nil {
  // Do Stuff
}
```

It is also common for applications to run a task only once. The below example shows scheduling a task to run only once 
after waiting for 60 seconds.

```go
// Add a one time only task for 60 seconds from now
id, err := scheduler.Add(&tasks.Task{
  Interval: time.Duration(60 * time.Second)
  RunOnce:  true,
  TaskFunc: func() error {
    // Put your logic here
  }(),
})
if err != nil {
  // Do Stuff
}
```

One powerful feature of Tasks is that it allows users to specify custom error handling. This is done by allowing users 
to define a function that is called when a task returns an error. The below example shows scheduling a task that logs 
when an error occurs.

```go
// Add a task with custom error handling
id, err := scheduler.Add(&tasks.Task{
  Interval: time.Duration(30 * time.Second),
  TaskFunc: func() error {
    // Put your logic here
  }(),
  ErrFunc: func(e error) {
    log.Printf("An error occured when executing task %s - %s", id, e)
  }(),
})
if err != nil {
  // Do Stuff
}
```

