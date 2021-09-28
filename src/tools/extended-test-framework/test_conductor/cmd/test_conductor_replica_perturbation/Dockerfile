# syntax=docker/dockerfile:1
FROM ubuntu:18.04
COPY test_conductor_replica_perturbation /
#COPY config.yaml /

COPY csi-daemonset.yaml /
COPY mayastor-daemonset.yaml /
COPY moac-deployment.yaml /
COPY moac-rbac.yaml /
COPY nats-deployment.yaml /

COPY etcd/* /etcd/

EXPOSE 8080/tcp
CMD ["./test_conductor_replica_perturbation"]
