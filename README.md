# geocloud

## Developing

### Prerequisites

* golang is *required* - version 1.11.x or above is required for go mod to work
* containerd is *required* - version 1.5.x is tested; earlier versions may also work
* runc is *required* - version 1.0.0-rc9x is tested; earlier versions may also work
* docker is recommended - version 20.10.x is tested; earlier versions may also work
* go mod is used for dependency management of golang packages

> _Note: containerd and runc are dependencies used by and installed alongside docker as of version 1.11_

### Running

#### Worker

> _Note: when running the worker inside of a container, be sure not to override the default value for the_ `--containerd-root` _flag, as `/var/lib/geocloud/containerd` is the only non-overlayfs volume in the container_

```sh
# [recommended] run in container 
docker build -t geocloud .
docker run --rm --privileged --tmpfs /run geocloud worker
```

> _Note: when running the worker outside of a container, if your machine already has containerd running on it, the flags specified above may need to be altered depending on file locations on your machine and are recommended so as not to impact your containerd installation. See_ `geocloud worker --help` _for more info_

```sh
# run on host machine
go build -o bin/geocloud ./cmd/ worker --containerd-no-run --containerd-address /run/containerd/containerd.sock
bin/geocloud
```

### CI

#### Set Pipeline

```sh
# login and set pipeline
fly -t geocloud login -c https://ci.logsquaredn.io --team-name geocloud
fly -t geocloud set-pipeline -p geocloud -c ci/pipeline.yml
# set pipeline from a template
# ytt -f ci/pipeline | fly -t geocloud set-pipeline -p geocloud -c -
```

> _Note: The_ `k14s/image` _image can be used in place of installing ytt on your machine to set a pipeline from a template. e.g.:_

```sh
# set pipeline from a template using k14s/image
docker run --rm -v `pwd`:/src:ro k14s/image ytt -f /src/ci/pipeline | fly -t geocloud set-pipeline -p geocloud -c -
```
