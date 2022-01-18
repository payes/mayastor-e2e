From ubuntu:18.04
ARG GO_VERSION
COPY  ./ /
# Parted is used to create and manipulate disk partition.
# Udev provides a dynamic way of setting up device.
# It ensures that devices are configured as soon as they are plugged in and discovered.
# It propagates information about a processed device.
RUN apt-get update; apt-get install net-tools iptables wget parted udev nvme-cli -y;
RUN wget https://golang.org/dl/go${GO_VERSION}.linux-amd64.tar.gz; \
	tar -C /usr/local/ -xzf go${GO_VERSION}.linux-amd64.tar.gz; \
	rm -rf go${GO_VERSION}.linux-amd64.tar.gz; \
	export PATH=$PATH:/usr/local/go/bin; go build; rm -rf /usr/local/go; \
	mv e2e-agent /usr/local/bin
CMD ["e2e-agent"]
