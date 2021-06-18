# VERSION 0.1
FROM ubuntu:20.04 AS builder

RUN apt-get update && apt-get install -qy fio gcc && rm -fr /var/cache/apt/*

ADD e2e_fio.c /

RUN gcc -Wall -o /e2e_fio e2e_fio.c

FROM ubuntu:20.04

MAINTAINER mayastor mayastor@mayadata.io

COPY --from=builder /e2e_fio /e2e_fio

COPY DiskTest /usr/local/bin

ENTRYPOINT ["/e2e_fio"]