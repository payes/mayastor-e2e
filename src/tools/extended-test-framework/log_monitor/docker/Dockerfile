# syntax=docker/dockerfile:1
FROM ubuntu:18.04
RUN apt-get update && apt-get install ca-certificates -y
COPY log_monitor /
# the folling file is provided by a config map
#COPY config-local.yaml /config/config-local.yaml
CMD ["/log_monitor"]
