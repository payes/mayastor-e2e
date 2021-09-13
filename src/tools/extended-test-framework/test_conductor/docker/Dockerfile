# syntax=docker/dockerfile:1
FROM ubuntu:18.04
COPY test_conductor ./
COPY config.yaml ./
EXPOSE 8080/tcp
CMD ["./test_conductor"]
