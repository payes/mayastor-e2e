# VERSION 0.1
FROM ubuntu

MAINTAINER mayastor mayastor@mayadata.io

RUN apt-get update

RUN apt-get install -y fio gcc musl-dev

COPY DiskTest /usr/local/bin

ADD e2e_fio.c /

RUN gcc -Wall -o /e2e_fio e2e_fio.c

ENTRYPOINT ["/e2e_fio"]
