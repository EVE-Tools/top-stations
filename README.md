# Top Stations
[![Build Status](https://drone.element-43.com/api/badges/EVE-Tools/top-stations/status.svg)](https://drone.element-43.com/EVE-Tools/top-stations) [![Docker Image](https://images.microbadger.com/badges/image/evetools/top-stations.svg)](https://microbadger.com/images/evetools/top-stations)

This service for [Element43](https://element-43.com) serves a list of top stations by market volume, based on data gathered from [esi-markets](https://github.com/EVE-Tools/esi-markets). The statistics are updated hourly and persisted to disk using BoltDB.

Issues can be filed [here](https://github.com/EVE-Tools/element43). Pull requests can be made in this repo.

## Interface
The service's gRPC description can be found [here](https://github.com/EVE-Tools/element43/blob/master/services/topStations/topStations.proto).

## Installation
Either use the prebuilt Docker images and pass the appropriate env vars (see below), or:

* Install Go, clone this repo into your gopath
* Get and run [esi-markets](https://github.com/EVE-Tools/esi-markets) 
* Set the ESI_MARKETS_HOST environment variable (see below)
* Run `go get ./...` to fetch the service's dependencies
* Run `bash generateProto.sh` to generate the necessary gRPC-related code
* Run `go build` to build the service
* Run `./top-stations` to start the service

Now a gRPC server will listen on port `43000` unless configured otherwise, serving stats data once it has been generated.

## Deployment Info
Builds and releases are handled by Drone.

Environment Variable | Default | Description
--- | --- | ---
CRON | @hourly | Stats refresh interval
DB_PATH | top-stations.db | DB persistence path
LOG_LEVEL | info | The service's log level
ESI_MARKETS_HOST | esi-markets.element43.svc.cluster.local:43000 | Host/port of the esi-markets instance to be used
PORT | 43000 | Port the gRPC server will listen on

