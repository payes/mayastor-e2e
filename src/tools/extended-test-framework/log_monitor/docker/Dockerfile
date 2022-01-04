# syntax=docker/dockerfile:1
FROM ubuntu:18.04
RUN apt-get update && apt-get install ca-certificates -y
COPY log_monitor /
CMD ["/log_monitor"]
