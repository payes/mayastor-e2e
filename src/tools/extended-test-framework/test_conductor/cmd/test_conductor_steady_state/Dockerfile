# syntax=docker/dockerfile:1
FROM alpine
COPY test_conductor_steady_state /

COPY csi-daemonset.yaml /
COPY mayastor-daemonset.yaml /
COPY moac-deployment.yaml /
COPY moac-rbac.yaml /
COPY nats-deployment.yaml /

COPY etcd/* /etcd/

EXPOSE 8080/tcp
CMD ["./test_conductor_steady_state"]
