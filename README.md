# geocloud

## Development

### Prerequisites

* golang is *required* - version 1.11.x or above is required for go mod to work
* containerd is *required* - version 1.5.x is tested; earlier versions may also work
* runc is *required* - version 1.0.0-rc9x is tested; earlier versions may also work
* docker is recommended - version 20.10.x is tested; earlier versions may also work
* go mod is used for dependency management of golang packages

> _Note: containerd and runc are dependencies used by and installed alongside docker as of version 1.11_

### Running

```sh
# run inside of a container (recommended)
docker build -t geocloud .
docker run --rm --privileged --tmpfs /run geocloud

# run on host machine
go build -o bin/geocloud ./cmd/
bin/geocloud
```

> _Note: when running the worker outside of a container, some set up and additional flags may be required. see_ `geocloud worker --help` _for more info_
