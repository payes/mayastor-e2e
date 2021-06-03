From ubuntu
RUN apt-get update
# Parted is used to create and manipulate disk partition.
# Udev provides a dynamic way of setting up device.
# It ensures that devices are configured as soon as they are plugged in and discovered.
# It propagates information about a processed device.
RUN apt-get install net-tools iptables wget parted udev nvme-cli -y
ADD  ./ /
RUN wget https://golang.org/dl/go1.16.3.linux-amd64.tar.gz
RUN tar -C /usr/local/ -xzf go1.16.3.linux-amd64.tar.gz; rm -rf go1.16.3.linux-amd64.tar.gz
RUN export PATH=$PATH:/usr/local/go/bin; go build; rm -rf /usr/local/go
RUN cp e2e-agent /usr/local/bin
CMD ["e2e-agent"]
