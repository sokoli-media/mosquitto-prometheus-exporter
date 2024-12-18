# mosquitto-prometheus-exporter

Simple project to export Mosquitto metrics to Prometheus.

Docker image: `sokolimedia/mosquitto-prometheus-exporter:latest`

Project exports http api on `:9000` with metrics at `/metrics` url.

Required environmental variables:
* `MOSQUITTO_BROKER`: host of your Mosquitto instance
* `MOSQUITTO_CLIENT_ID`: customer's client id (set to something unique, e.g. name of this project)
* `MOSQUITTO_USERNAME`: Mosquitto username
* `MOSQUITTO_PASSWORD`: Mosquitto password

Prerequisites:
* mosquitto installed
