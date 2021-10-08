# syntax=docker/dockerfile:1
FROM alpine
#ADD https://storage.googleapis.com/kubernetes-release/release/v1.15.2/bin/linux/amd64/kubectl /usr/local/bin/kubectl
#RUN chmod +x /usr/local/bin/kubectl

#COPY kubectl-mayastor /
COPY test_conductor_steady_state /

EXPOSE 8080/tcp
CMD ["./test_conductor_steady_state"]
