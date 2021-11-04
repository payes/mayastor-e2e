# syntax=docker/dockerfile:1
FROM ubuntu:18.04
RUN apt-get update && apt-get install ca-certificates -y
COPY test_director /
# the folling file is provided by a config map
#COPY config-local.yaml /config/config-local.yaml
EXPOSE 8080/tcp
CMD ["/test_director"]
