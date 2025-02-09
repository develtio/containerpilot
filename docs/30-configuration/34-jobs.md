# Jobs

A ContainerPilot job is a user-defined process and rules for when to execute it, how to health check it, and how to advertise it to Consul. The rules are intended to allow for flexibility to cover nearly any type of process one might want to run. Some possible job configurations include:

- A long running application like a web server, which needs to be restarted if it crashes.
- A one-time setup task that runs at the start of a container's lifetime but never again afterwards.
- A periodic process that runs every few minutes or hours, such as a backup.
- A task that runs when some other event occurs, such as running only when another job has become healthy.


## Lifecycle Events

Every job will emit events associated with the lifecycle of its process. Any job can react to the events emitted by any other job (or even its own events) via the [`when`](#when) configuration.

- `healthy`: emitted when the job's [health check](#health-check) succeeds.
- `unhealthy`: emitted when the job's [health check](#health-check) fails.
- `exitSuccess`: emitted when the process associated with the job exits with an exit code 0.
- `exitFailed`: emitted when the process associated with the job exits with a non-0 exit code.
- `stopping`: emitted when the job is asked to stop but before it does so. Useful when the job has a [stop timeout](#stop-timeout).
- `stopped`: emitted when the job is stopped. Note that this is not the same as the process exiting because a job might have many executions of its process.

Note that although `stopping` and `stopped` events are emitted for each running job when ContainerPilot is shutting down, the receiving job will have a limited window in which to execute. This window is 5 seconds, in order to provide enough time for ContainerPilot to halt all jobs, gracefully shut down its own listeners, and exit within the default Docker shutdown timeout of 10 seconds. After this point all processes receive a `SIGKILL` and are forced to exit immediately.

Additionally, jobs may react to these events:

- `startup`: published to all jobs when ContainerPilot is ready to start.
- `shutdown`: published to all jobs when ContainerPilot is shutting down.
- `changed`: published when a [`watch`](./30-configuration/35-watches.md) sees a change in a dependency.
- `enterMaintenance`: published when the [control plane](./30-configuration/37-control-plane.md) is told to enter maintenance mode for the container. All jobs will be automatically deregistered from Consul when this happens, so you only want to react to this event if there is some other task to perform.
- `exitMaintenance`: published when the [control plane](./30-configuration/37-control-plane.md) is told to exit maintenance mode for the container.

Finally, there are two special `source` values that can be used to trigger a job when ContainerPilot receives a UNIX signal.

- `SIGHUP`: published when a ContainerPilot process receives the UNIX signal `SIGHUP`.
- `SIGUSR2`: published when a ContainerPilot process receives the UNIX signal `SIGUSR2`.

Signal events come in handy when you need to kick off some type of special process (reloading configs, publishing debug info) inside a container. This type of external communication is only supported when running ContainerPilot within a Docker container running on a Docker host, or under the supervision of a scheduler like Nomad.

Note: Either two signals can be sent to ContainerPilot acting as a PID 1 supervisor or its standalone worker process.

## Configuration

Job configurations include the following fields:

```json5
jobs: [
  {
    name: "app",
    exec: "/bin/app",
    logging: {
      raw: false
    },

    // 'when' defines the events that cause the job to run
    when: {
      source: "setup",
      once: "exitSuccess",
      timeout: "60s"
      // interval: "10s",     // can't be set at the same time as 'source'/'once'
      // each: "exitSuccess", // can't be set at the same time as 'once'
    },

    // these fields interact with 'when' behaviors (see below)
    timeout: "300s",
    stopTimeout: "10s",
    restarts: "unlimited",

    // 'health' defines how the job is health checked
    health: {
      exec: "/usr/bin/curl --fail -s -o /dev/null http://localhost/app",
      interval: 5,
      ttl: 10,
      timeout: "5s",
    },

    // 'port', 'tags', 'interfaces', and 'consul' define options for
    // service discovery with Consul
    port: 80,
    initial_status: "warning", // optional status to immediately register service with
    tags: [
      "app",
      "prod"
    ],
    meta: {
      keyA: "A",
      keyB: "B",
    },
    interfaces: [
      "eth0",
      "eth1[1]",
      "192.168.0.0/16",
      "2001:db8::/64",
      "eth2:inet",
      "eth2:inet6",
      "inet",
      "inet6",
      "static:192.168.1.100", // a trailing comma isn't an error!
    ],
    consul: {
      enableTagOverride: true,
      deregisterCriticalServiceAfter: "10m"
    }
  }
]
```

##### `name`

The `name` field is the name of the job as it will appear in logs and events. It will also be the name of the service as it will appear in Consul (if the job is registered to Consul). Each instance of the service in Consul will have a unique ID made up from the `name`+hostname of the container. Names must match requirements for Consul; they start with a lower-case letter and contain upper- or lower-case letters, numerals, or `-` but no other characters. (Or in other words they must match the regex `^[a-z][a-zA-Z0-9\-]+$`)

##### `exec`

The `exec` field is the executable (and its arguments) that is called when the job runs. This field can contain a string or an array of strings ([see below](#exec-arguments) for details on the format). The command to be run will have a process group set and this entire process group will be reaped by ContainerPilot when the process exits. The process will be run concurrently to all other work, so the process won't block the processing of other ContainerPilot events.

##### `uid`

The `uid` field is the ID of the user that runs the command.

##### `gid`

The `gid` field is the ID of the group that runs the command.

##### `logging`

Jobs and health checks have a `logging` configuration block with a single option: `raw`. When the `raw`field is set to `false` (the default), ContainerPilot will wrap each line of output from an `exec` process's stdout/stderr in a log line. If set to `true`, ContainerPilot will attach the stdout/stderr of the process to the container's stdout/stderr and these streams will be unmodified by ContainerPilot. The latter option can be useful if the process emits structured logs in its own format.

#### Running and timing fields

The following fields define when a job starts, stops, restarts, and times out.

##### `when`

The `when` field defines a hook for an event that starts the job's `exec`. By default, a job's `exec` process starts as soon as ContainerPilot has finished startup. Many jobs will want to have a configuration that determines some specific event to wait for, using the `when` field.

- `source` is the source of the event that triggers the job.
- `once` names an event that triggers the start of the job one time only.
- `each` names an event that triggers the start of the job every time it happens.
- `interval` is the time between executions of the job. Supports milliseconds, seconds, minutes. The frequency must be a positive non-zero duration with a time unit suffix. (Example: `60s`. See the golang [`ParseDuration`](https://golang.org/pkg/time/#ParseDuration) docs for this format.) Valid time units are `ns`, `us` (or `µs`), `ms`, `s`, `m`, `h`. The minimum interval is `1ms` but in practice it takes 20-50ms for a process to be forked and executed so the interval should be considerably longer.
- `timeout` under `when` is optional and is the amount of time to wait for the `when` event to be received before giving up. The format for this field is the same as that of `interval`.

If the `interval` field is set it is the only field permitted under `when`. Otherwise, the `once` and `each` fields are mutually exclusive -- you can set one or the other but not both.

##### `timeout`

The `timeout` field is optional and is the amount of time to wait after the job starts before it is killed. Processes killed this way are terminated immediately (`SIGKILL`) without an opportunity to clean up their state and a heartbeat will not be sent.

For long-running jobs like servers, you will generally want to omit this field. If this field is omitted and the job does not have a [`when.frequency` field](#when), then the job will never timeout. If the field is omitted and the job does have a `when.frequency` field, then the timeout will default to the frequency.

If set and not left as the default, the minimum timeout is `1ms` (see the golang [`ParseDuration`](https://golang.org/pkg/time/#ParseDuration) docs for this format) but in practice it takes 20-50ms for a process to be forked and executed so the timeout should be considerably longer.

##### `stopTimeout`

`stopTimeout` is the maximum amount of time a `stopping` job will wait for another job that might be watching for the `stopping` event.

This can be useful for jobs which need to perform a task when they begin shutting down, but before they've done so. For example, a Consul agent might need to be removed from the list of available nodes via `consul leave`, which requires that the agent process is still running.

In the example below, the `consul-agent` job will need to leave time between the `stopping` and `stopped` events in order for `consul leave` to perform. The `stopTimeout` field is the time that the job will wait before exiting and killing its process. `consul-agent` job waits 5 seconds after it's `stopping` event has fired in order to allow the `leave-consul` job to execute.

```json5
jobs: [
  {
    name: "consul-agent",
    exec: "consul agent...",
    stopTimeout: "5s"
  },
  {
    name: "leave-consul",
    exec: "consul leave",
    when: {
      source: "consul-agent",
      once: "stopping"
    }
  }
]
```

The job that's watching for the `stopping` event can take however long it wants to do it's work. If you want to make sure the watching job is also going to finish, you need to add the `timeout` field to that job as well.

##### `restarts`

The `restarts` field is the number of times the process will be restarted if it exits. This field supports any non-negative numeric value (ex. `0` or `1`) or the strings `"unlimited"` or `"never"`. This value is optional and usually defaults to `"never"` (see the note below about the `interval` field for the exception).

It's important to understand how this field compares to the `when` field. A restart is run only when the job receives its own `exitSuccess` or `exitFailed` event. The `when` field is for triggering on other events. In the example below the `app` job is first started when the `db` job is `healthy`, but it will restart whenever it exits. Using `restarts` with the `each` option of `when` is not recommended because each time the `each` event triggers, it will spawn an `exec` that can restart after exit. In the case of unlimited restarts this would eventually use up all the resources in your container, so trying to use `restarts: "unlimited"` and `each` will return an error.

```json5
jobs: [
  {
    name: "app",
    restarts: "unlimited",
    when: {
      source: "db",
      once: "healthy"
    }
  }
]
```

The behavior of `restarts` is somewhat different if the `when` field is using the `interval` option. In this case, the `restarts` field indicates how many times the `exec` will be run on that interval. In the example configuration below, the `app` job will be run every 5 seconds for a maximum of 4 times (3 restarts). When the `interval` is set, the `restarts` field defaults to `"unlimited"`, which means the job will run every `interval` period without stopping.

```json5
jobs: [
  {
    name: "app",
    restarts: 3,
    when: {
      interval: "5s"
    }
  }
]
```

#### Health checks

The `health` field defines how ContainerPilot determines if a job is healthy. This field is optional. Jobs without a `health` field set will not emit `healthy` and `changed` events.

- `exec` field is the executable (and its arguments) to run to health check the job.
- `interval` is the time in seconds between health checks.
- `ttl` is the time-to-live in seconds of a successful health check. This should be longer than the `interval` polling rate so that the check and the TTL aren't racing; otherwise the job will be marked unhealthy in Consul.
- `timeout` is a value to wait before forcibly killing the health check `exec`. Health checks killed this way are terminated immediately (`SIGKILL`) without an opportunity to clean up their state and a heartbeat will not be sent. The minimum timeout is `1ms` (see the golang [`ParseDuration`](https://golang.org/pkg/time/#ParseDuration) docs for this format) but in practice it takes 20-50ms for a process to be forked and executed so the timeout should be considerably longer.


#### Service discovery

The following fields define how a job is registered with Consul.

##### `port`

The `port` field is the port the service will advertise to Consul. Note that this assumes the job is listening on that port but does not change anything in the process. If you want to dynamically assign this port you might use an environment variable and [template rendering](./32-configuration-file.md#template-rendering) for both the `exec` and `port` fields. For example:

```json5
jobs: [
  {
    name: "myjob",
    exec: "node /bin/server.js --port {{ .PORT }}",
    port: {{ .PORT }}
  }
]
```

##### `initial_status`

The `initial_status` field is optional and specifies which status to immediately register the service with. If not specified, the service will not be registered in consul until after the first successful health check. Valid values are `passing`, `warning` or `critical`.

##### `tags`

The `tags` field is an optional array of tags to be used when the job is registered as a service in Consul. Other containers can use these tags in `watches` to filter a service by tag.

##### `meta`

The `meta` field is an optional map key/value to be used when the job is registered as a service in Consul. Key names must be valid JSON5/Ecmascript identifierNames or be quoted and follow consul limitation , practical this means only [a-zA-Z0-9_-] can be used in key names and key names with '-' must be quoted

##### `interfaces`

The `interfaces` field is an optional single or array of interface specifications. If given, the IP of the service will be obtained from the first interface specification that matches. (Default value is `["eth0:inet"]`). The value that ContainerPilot uses for the IP address of the interface will be set as an environment variable with the name `CONTAINERPILOT_{JOB}_IP`. See the [environment variables](./32-configuration-file.md#environment-variables) section.

##### `consul`

The `consul` field is an optional block of job-specific Consul configuration.

- `enableTagOverride` if set to true, then external agents can update this service in the catalog and modify the tags.
- `deregisterCriticalServiceAfter` is a timeout in Go time format. If a check is in the critical state for more than this configured value, then its associated service (and all of its associated checks) will automatically be deregistered.


#### Exec arguments

All `exec` fields that configure a child process (`jobs/exec` and `jobs/health/exec`) accept both a string or an array. If a string is given, the command and its arguments are separated by spaces; otherwise, the first element of the array is the command path, and the rest are its arguments. This is sometimes useful for breaking up long command lines.

**String command**

```json5
health: {
  exec: "/usr/bin/curl --fail -s http://localhost/app"
}
```

**Array command**

```json5
health: {
  exec: [
    "/usr/bin/curl",
    "--fail",
    "-s",
    "http://localhost/app"
  ]
}
```
