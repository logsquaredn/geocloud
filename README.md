# geocloud

## Developing

> _many commands in this README communicate with AWS, and so rely on access key ID and secret access key credentials_

### Prerequisites

* golang is *required* - version 1.11.x or above is required for go mod to work
* docker is *required* - version 20.10.x is tested; earlier versions may also work
* docker-compose is *required* - version 1.29.x is tested; earlier versions may also work
* go mod is *required* for dependency management of golang packages
* make is *required* - version 3.81 is tested; earlier versions may also work
* awscli is recommended - version 1.18.69 is tested; earlier versions may also work

### Running

> `docker-compose` _requires credentials to be supplied through the shell via environment variables_ `AWS_ACCESS_KEY_ID` _and_ `AWS_SECRET_ACCESS_KEY` _or an environment file_ `.env` _in the root of the repository_

```sh
# setup services
make infra
# run geocloud
make up
# restart geocloud
make restart
```

### Migrations

#### Create Migration

```sh
# generate a migration version
version=`date -u +%Y%m%d%T | tr -cd [0-9]`
touch datastore/psql/migrations/core/${version}_my_title.up.sql
```

see [Postgres migration tutorial](https://github.com/golang-migrate/migrate/blob/master/database/postgres/TUTORIAL.md)

#### Migrate

```sh
# run migrations
geocloud migrate
```
