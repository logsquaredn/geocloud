# geocloud

## Developing

> _many commands in this README communicate with AWS, and so rely on access key ID and secret access key credentials_

### Prerequisites

* golang is *required* - version 1.11.x or above is required for go mod to work
* containerd is *required* - version 1.5.x is tested; earlier versions may also work
* runc is *required* - version 1.0.0-rc9x is tested; earlier versions may also work
* docker is recommended - version 20.10.x is tested; earlier versions may also work
* docker-compose is recommended - version 1.29.x is tested; earlier versions may also work
* terraform is recommended - version 1.0.2 is tested; earlier versions may also work; docker can be used instead
* awscli is recommended - version 1.18.69 is tested; earlier versions may also work
* go mod is used for dependency management of golang packages

> _containerd and runc are dependencies used by and installed alongside docker as of version 1.11_

### Running

> `docker-compose` _requires credentials to be supplied through the shell via environment variables_ `AWS_ACCESS_KEY_ID` _and_ `AWS_SECRET_ACCESS_KEY` _or an environment file_ `.env` _in the root of the repository_

> _if you run into problems connecting to postgres, try running_ `docker-compose up -d postgres` _before executing the following commands_

```sh
# run geocloud
docker-compose up --build
```

#### API

```sh
# run the api
docker-compose up --build api
```

#### Worker

> _when running the worker inside of a container, the_ `--containerd-root` _flag always falls back to_ `/var/lib/containerd` _as it is the only non-overlayfs volume in the container, making it the only volume in the container suitable to be containerd's root directory_

> _when running the worker inside of a container,_ `*-ip` _flags always fall back to_ `0.0.0.0`

```sh
# run the worker
docker-compose up --build worker
```

### Infrastructure

> `terraform` _requires credentials to be supplied through the shell via environment variables_ `AWS_ACCESS_KEY_ID` _and_ `AWS_SECRET_ACCESS_KEY` _or a credentials file configured in_ `~/.aws/`

#### Create Infrastructure

```sh
# create queue and bucket
terraform -chdir=infrastructure/tf/ init
terraform -chdir=infrastructure/tf/ apply
```

> `hashicorp/terraform` _can be used in place of installing terraform on your machine to create infrastructure_

```sh
# create queue and bucket using hashicorp/terraform 
docker run --rm -v `pwd`:/src/:ro -e AWS_ACCESS_KEY_ID=$AWS_ACCESS_KEY_ID -e AWS_SECRET_ACCESS_KEY=$AWS_SECRET_ACCESS_KEY hashicorp/terraform -chdir=/src/infrastructure/ apply
```

### Migrations

#### Create Migration

```sh
# generate a migration version
version=`date -u +%Y%m%d%T | tr -cd [0-9]`
touch datastore/psql/migrations/${version}_my-title.up.sql
touch migrate/migratecmd/migrations/${version}_my-title.down.sql
```

see [Postgres migration tutorial](https://github.com/golang-migrate/migrate/blob/master/database/postgres/TUTORIAL.md)

#### Migrate

```sh
# run the migrations
geocloud migrate
```

### Tasks

#### Build Tasks

```sh
cd tasks/
docker-compose build
docker-compose push
```
