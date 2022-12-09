# rototiller

## Developing

### Prerequisites

* golang is *required* - version 1.18 or above is required for generics
* docker is *required* - version 20.10.x is tested; earlier versions may also work
* docker-compose is *required* - version 1.29.x is tested; earlier versions may also work
* make is *required* - version 3.81 is tested; earlier versions may also work

### Running

```sh
# setup services
make infra
# run rototiller
make up
# restart rototiller
make restart
```

### Release

```sh
# push a tag--automation handles the rest
# please use semantic versioning
make release VERSION=1.0.0 PRERELEASE=
```

### Examples

#### Example API calls

```sh
# get an API key
curl -d '{ "email": "" }' -X POST "https://rototiller.logsquaredn.io/api/v1/api-key"
# create buffer job
curl -X POST -H "Content-Type: application/zip" -H "Authorization: <token>" --data-binary '@/path/to/a.zip' "https://rototiller.logsquaredn.io/api/v1/jobs/buffer?buffer-distance=5&quadrant-segment-count=50"
# get job result
curl -X GET -H "Content-Type: application/zip" -H "Authorization: <token>" -o "/path/to/a.zip" "https://rototiller.logsquaredn.io/api/v1/jobs/9b45f141-a137-4f52-a36f-2640129d92e8/output/content"
# create storage
curl -X POST -H "Content-Type: application/zip" -H "Authorization: <token>" --data-binary '@/path/to/a.zip' "https://rototiller.logsquaredn.io/api/v1/storages?name=<name>"
# create vector lookup job
curl -X POST -H "Content-Type: application/zip" -H "Authorization: <token>" --data-binary '@/path/to/a.zip' "https://rototiller.logsquaredn.io/api/v1/jobs/vectorlookup?attributes=RADII,ADVNUM&latitude=20.33&longitude=-64.23"
```
