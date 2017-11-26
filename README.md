# Top Stations
[![Build Status](https://drone.element-43.com/api/badges/EVE-Tools/top-stations/status.svg)](https://drone.element-43.com/EVE-Tools/top-stations) [![Docker Image](https://images.microbadger.com/badges/image/evetools/top-stations.svg)](https://microbadger.com/images/evetools/top-stations)

This service for [Element43](https://element-43.com) serves a list of top stations by market volume, based on data gathered from E43's API. As the dataset is very large (multiple hundred megabytes of JSON), a streaming parser is used. Data is collected by a Python script on startup and every 60 minutes from then on. All aggregate data is put into a file which is served by a Caddy server. TLS is not enabled as it is expected to run this service behind a TLS-terminating reverse proxy. It can be enabled by editing the `Caddyfile`.

## Installation
Either use the prebuilt Docker images, or:

* Clone this repo
* Install the Caddy webserver on your system
* Run `pip install -r requirements.txt`
* Run `python main.py`

Now a server will listen on port `8000` unless configured otherwise, serving stats data once they have been generated.

## Deployment Info
Builds and releases are handled by Drone.

Environment Variable | Default | Description
--- | --- | ---
PORT | 8000 | Port the integrated webserver will listen on

## Endpoints
URL Pattern | Description
--- | ---
`/api/top-stations/v1/list` | Get stats for all stations in New Eden.

```json
[
  {
      "ask_volume": 149456931151.31012,
      "station_id": 60012550,
      "bid_volume": 4816469548.169999,
      "total_volume": 154273400699.4801,
      "total_orders": 1636
  },
  {
      "ask_volume": 409047916414.52997,
      "station_id": 60013012,
      "bid_volume": 6911425319661.421,
      "total_volume": 7320473236075.952,
      "total_orders": 1608
  },
  ...
]
```
