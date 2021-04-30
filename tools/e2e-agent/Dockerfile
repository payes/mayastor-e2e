From ubuntu
RUN apt-get update
RUN apt-get install net-tools iptables wget -y
ADD  ./ /
RUN wget https://golang.org/dl/go1.16.3.linux-amd64.tar.gz
RUN tar -C /usr/local/ -xzf go1.16.3.linux-amd64.tar.gz; rm -rf go1.16.3.linux-amd64.tar.gz
RUN export PATH=$PATH:/usr/local/go/bin; go build; rm -rf /usr/local/go
RUN cp e2e-agent /usr/local/bin
CMD ["e2e-agent"]
