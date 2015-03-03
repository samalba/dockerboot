# dockerboot

**Disclaimer**: This is an experimental re-implementation of a subset of [Docker Compose](https://github.com/docker/compose).

Boot your machine with a simple fig.yml file and docker.

## How to use?

Install [Golang](https://golang.org/doc/install) and configure your [GOPATH](https://github.com/golang/go/wiki/GOPATH).

Then:

```
$ go get github.com/samalba/dockerboot
$ export PATH=$PATH:${GOPATH}/bin
$ dockerboot
Usage: dockerboot [OPTIONS] COMMAND

Options:
  -H="unix:///var/run/docker.sock": URL of the docker daemon
  -f="./fig.yml": Fig yaml file to read the services config from

Commands:
  start: Start all services and process configuration updates
  reload: Process configuration updates (alias to `start')
  stop: Stop all services
  restart: Restart all services and process configuration updates
```

## Give me an example

**Step 1:** Write a `fig.yml' file that contains the services you want to boot your system with.

*Note: this example will do nothing but start containers...*

```yaml
web:
  image: ubuntu:14.04
  command: cat
  ports:
   - "8000:8000"
db:
  image: busybox
  command: cat
```

**Step 2:** Start the services using dockerboot.

```
$ ./dockerboot start
2015/02/28 17:57:08 main.Config{FigFile:"./fig.yml", DockerUrl:"unix:///var/run/docker.sock"}
2015/02/28 17:57:09 Discovered possible existing services: web, db, foo, bar, test2, test1, test, furious_galileo
2015/02/28 17:57:09 Service `web' does not change
2015/02/28 17:57:09 Starting service `web'...
2015/02/28 17:57:09 Service `db' does not change
2015/02/28 17:57:09 Starting service `db'...
```

**Step 3:** You can still use `docker' directly for everything else.

```
$ docker ps
CONTAINER ID        IMAGE               COMMAND             CREATED             STATUS              PORTS               NAMES
1e733ffc9162        ubuntu:14.04        "cat"               12 days ago         Up 24 minutes       8000/tcp            web
f5e1f4817064        busybox:latest      "cat"               12 days ago         Up 24 minutes                           db
```
