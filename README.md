# ContainerPilot

*An init system for cloud-native distributed applications that automates the process of service discovery, configuration, and lifecycle management inside the container, so you can focus on your apps.*

[![Build Status](https://drone.greenbaum.cloud/api/badges/greenbaum.cloud/containerpilot/status.svg)](https://drone.greenbaum.cloud/greenbaum.cloud/containerpilot)
[![MPL licensed](https://img.shields.io/badge/license-MPL_2.0-blue.svg)](https://github.com/tritondatacenter/containerpilot/blob/master/LICENSE)
[![GoDoc](https://godoc.org/github.com/tritondatacenter/containerpilot?status.svg)](https://godoc.org/github.com/tritondatacenter/containerpilot)

## What is ContainerPilot?

Orchestration is the automation of the operations of an application. Most application require operational tasks like connecting them to related components ([WordPress needs to know where it's MySQL and Memcached servers are, for example](https://www.tritondatacenter.com/blog/wordpress-on-autopilot)), and some applications require special attention as they start up or shut down to be sure they bootstrap correctly or persist their data. We can do all that by hand, but modern applications automate those tasks in code. That's called "orchestration."

To make this work, every application needs to do the following (at a minimum):

- Register itself in a service catalog (like Consul or Etcd) for use by other apps
- Look to the service catalog to find the apps it depends on
- Configure itself when the container starts, and reconfigure itself over time

We can write our new applications to do that, but existing apps will need some help. We can wrap each application in a shell script that registers itself with the discovery service easily enough, but watching for changes to that service and ensuring that health checks are being made is more complicated. We can put a second process in the container, but as soon as we do that we need an init system running inside the container as well.

### ContainerPilot to the rescue!

ContainerPilot is an init system designed to live inside the container. It acts as a process supervisor, reaps zombies, run health checks, registers the app in the service catalog, watches the service catalog for changes, and runs your user-specified code at events in the lifecycle of the container to make it all work right. ContainerPilot uses Consul to coordinate global state among the application containers.

## Quick Start Guide

Check out our ["Hello, World" application](https://github.com/autopilotpattern/hello-world) on GitHub. Assuming you have Docker and Docker Compose available, it's as easy as:

```
git clone git@github.com:autopilotpattern/hello-world.git
cd hello-world
docker-compose up -d
open http://localhost
```

This application blueprint demonstrates using ContainerPilot to update Nginx upstream configuration at runtime. Try scaling up via `docker-compose scale hello=2 world=3` to see the Nginx configuration updated.

You can also [download](https://github.com/tritondatacenter/containerpilot/releases) the latest release of ContainerPilot from GitHub.

## Documentation

Documentation for ContainerPilot and where it fits with the rest of the Triton ecosystem can be found at [www.tritondatacenter.com/containerpilot](https://www.tritondatacenter.com/containerpilot). The index below links to the documentation in this repo for convenience.

[Lifecycle](./docs/10-lifecycle.md)
- [What is a job?](./docs/10-lifecycle.md#what-is-a-job)
- [What is an event?](./docs/10-lifecycle.md#what-is-an-event)
- [What is a watch?](./docs/10-lifecycle.md#what-is-a-watch)
- [How do events trigger jobs?](./docs/10-lifecycle.md#how-do-events-trigger-jobs)
- [How can jobs be ordered?](./docs/10-lifecycle.md#how-can-jobs-be-ordered)

[Design: the Why of ContainerPilot](./docs/20-design.md)
- [Why active service discovery?](./docs/20-design.md#why-active-service-discovery)
- [Why isn't there a "post-start" or "started" event?](./docs/20-design.md#why-isnt-there-a-post-start-or-started-event)
- [Why Consul and not etcd or Zookeeper?](./docs/20-design.md#why-consul-and-not-etcd-or-zookeeper)
- [Why are jobs not the same as services?](./docs/20-design.md#why-are-jobs-not-the-same-as-services)
- [Why don't watches or metrics have an exec field?](./docs/20-design.md#why-dont-watches-or-metrics-have-an-exec-field)
- [Why use something other than ContainerPilot?](./docs/20-design.md#why-use-something-other-than-containerpilot)


Configuration
- [Installation](./docs/30-configuration/31-installation.md)
- [Configuration file](./docs/30-configuration/32-configuration-file.md)
  - [Schema](./docs/30-configuration/32-configuration-file.md#schema)
    - [Consul](./docs/30-configuration/32-configuration-file.md#consul)
    - [Logging](./docs/30-configuration/32-configuration-file.md#logging)
    - [Jobs](./docs/30-configuration/32-configuration-file.md#jobs)
    - [Watches](./docs/30-configuration/32-configuration-file.md#watches)
    - [Control](./docs/30-configuration/32-configuration-file.md#control)
    - [Telemetry](./docs/30-configuration/32-configuration-file.md#telemetry)
  - [Extras](./docs/30-configuration/32-configuration-file.md#configuration-extras)
    - [Interfaces](./docs/30-configuration/32-configuration-file.md#interfaces)
    - [Environment variables](./docs/30-configuration/32-configuration-file.md#environment-variables)
    - [Template rendering](./docs/30-configuration/32-configuration-file.md#template-rendering)
- [Consul](./docs/30-configuration/33-consul.md)
  - [Client configuration](./docs/30-configuration/33-consul.md#client-configuration)
  - [Consul agent configuration](./docs/30-configuration/33-consul.md#consul-agent-configuration)
- [Jobs](./docs/30-configuration/34-jobs.md)
  - [Lifecycle Events](./docs/30-configuration/34-jobs.md#lifecycle-events)
  - [Configuration](./docs/30-configuration/34-jobs.md#configuration)
    - [name](./docs/30-configuration/34-jobs.md#name)
    - [exec](./docs/30-configuration/34-jobs.md#exec)
    - [when](./docs/30-configuration/34-jobs.md#when)
    - [timeout](./docs/30-configuration/34-jobs.md#timeout)
    - [stopTimeout](./docs/30-configuration/34-jobs.md#stopTimeout)
    - [restarts](./docs/30-configuration/34-jobs.md#restarts)
    - [health checks](./docs/30-configuration/34-jobs.md#health-checks)
    - [service discovery](./docs/30-configuration/34-jobs.md#service-discovery)
  - [Exec arguments](./docs/30-configuration/34-jobs.md#exec-arguments)
- [Watches](./docs/30-configuration/35-watches.md)
- [Telemetry](./docs/30-configuration/36-telemetry.md)
  - [Sensor configuration](./docs/30-configuration/36-telemetry.md#sensor-configuration)
- [Control plane](./docs/30-configuration/37-control-plane.md)
  - [ContainerPilot subcommands](./docs/30-configuration/37-control-plane.md#containerpilot-subcommands)
- [Logging](./docs/30-configuration/38-logging.md)
- [Example configurations](./docs/30-configuration/39-config-examples.md)


[Support](./docs/40-support.md)
- [Where to file issues](./docs/40-support.md#where-to-file-issues)
- [Contributing](./docs/40-support.md#contributing)
- [Backwards compatibility](./docs/40-support.md#backwards-compatibility)

You might also read [our guide building self-operating applications with ContainerPilot](https://www.tritondatacenter.com/blog/applications-on-autopilot) and look at the examples below.

## Examples

We've published a number of example applications demonstrating how ContainerPilot works.

- [Applications on autopilot: a guide to how to build self-operating applications with ContainerPilot](https://www.tritondatacenter.com/blog/applications-on-autopilot)
- [MySQL (Percona Server) with auto scaling and fail-over](https://www.tritondatacenter.com/blog/dbaas-simplicity-no-lock-in)
- [Autopilot Pattern WordPress](https://www.tritondatacenter.com/blog/wordpress-on-autopilot)
- [ELK stack](https://www.tritondatacenter.com/blog/docker-log-drivers)
- [Node.js + Nginx + Couchbase](https://www.tritondatacenter.com/blog/docker-nodejs-nginx-nosql-autopilot)
- [CloudFlare DNS and CDN with dynamic origins](https://github.com/autopilotpattern/cloudflare)
- [Consul, running as an HA raft](https://github.com/autopilotpattern/consul)
- [Couchbase](https://github.com/autopilotpattern/couchbase)
- [Mesos on Joyent Triton](https://www.tritondatacenter.com/blog/mesos-by-the-pound)
- [Nginx with dynamic upstreams](https://www.tritondatacenter.com/blog/dynamic-nginx-upstreams-with-containerbuddy)
