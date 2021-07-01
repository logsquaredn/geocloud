# geocloud

## Developing

### Prerequisites

* golang is *required* - version 1.11.x or above is required for go mod to work
* containerd is *required* - version 1.5.x is tested; earlier versions may also work
* runc is *required* - version 1.0.0-rc9x is tested; earlier versions may also work
* docker is recommended - version 20.10.x is tested; earlier versions may also work
* ytt is recommended - version 0.34.x is tested; earlier versions may also work; docker can be used instead
* go mod is used for dependency management of golang packages

> _Note: containerd and runc are dependencies used by and installed alongside docker as of version 1.11_

### Running

#### Worker

> _Note: when running the worker inside of a container, the_ `--containerd-root` _flag always falls back to_ `/var/lib/geocloud/containerd` _as it is the only non-overlayfs volume in the container, making it the only volume in the container suitable to be containerd's root directory_

```sh
# [recommended] run in container 
docker build -t geocloud .
docker run --rm --privileged --tmpfs /run geocloud worker
```

> _Note: when running the worker outside of a container, it is recommended that you target an already-running containerd process by specifying the_ `--containerd-no-run` _and_ `--containerd-address` _flags. See_ `geocloud worker --help` _for more info_

```sh
# run on host machine
go run ./cmd/ worker --containerd-no-run --containerd-address /run/containerd/containerd.sock
# or
go build -o bin/geocloud ./cmd/
bin/geocloud worker --containerd-no-run --containerd-address /run/containerd/containerd.sock
```

### CI

#### Set Pipeline

```sh
# login and set pipeline
fly -t geocloud login -c https://ci.logsquaredn.io --team-name geocloud
fly -t geocloud set-pipeline -p geocloud -c ci/pipeline.yml -v branch=my-branch
# or set pipeline from template
ytt -f ci/pipeline | fly -t geocloud set-pipeline -p geocloud -c - -v branch=my-branch
```

> _Note: The_ `k14s/image` _image can be used in place of installing ytt on your machine to set a pipeline from a template_

```sh
# set pipeline from template using k14s/image
docker run --rm -v `pwd`:/src:ro k14s/image ytt -f /src/ci/pipeline | fly -t geocloud set-pipeline -p geocloud -c - -v branch=my-branch
```
