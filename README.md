# geocloud

## Developing

> _many commands in this README communicate with AWS, and so rely on access key ID and secret access key credentials_

### Prerequisites

* golang is *required* - version 1.11.x or above is required for go mod to work
* docker is *required* - version 20.10.x is tested; earlier versions may also work
* docker-compose is *required* - version 1.29.x is tested; earlier versions may also work
* go mod is *required* for dependency management of golang packages
* make is *required* - version 3.81 is tested; earlier versions may also work
* containerd is recommended - version 1.5.x is tested; earlier versions may also work
* runc is recommended - version 1.0.0-rc9x is tested; earlier versions may also work
* awscli is recommended - version 1.18.69 is tested; earlier versions may also work

> _containerd and runc are dependencies used by and installed alongside docker as of version 1.11_

### Running

> `docker-compose` _requires credentials to be supplied through the shell via environment variables_ `AWS_ACCESS_KEY_ID` _and_ `AWS_SECRET_ACCESS_KEY` _or an environment file_ `.env` _in the root of the repository_

```sh
# setup services, build tasks
make infra
# run geocloud
make up
# restart geocloud
make restart
```

### Tasks

#### Build Tasks

```sh
# build tasks to bin/
make build-tasks-c
# build task images
make build-tasks
# build and save tasks tarball to runtime/tasks.tar
make save-tasks
# build and push tasks
make push-tasks
```

### Migrations

#### Create Migration

```sh
# generate a migration version
version=`date -u +%Y%m%d%T | tr -cd [0-9]`
touch datastore/psql/migrations/${version}_my-title.up.sql
```

see [Postgres migration tutorial](https://github.com/golang-migrate/migrate/blob/master/database/postgres/TUTORIAL.md)

#### Migrate

```sh
# run migrations
geocloud migrate
```
