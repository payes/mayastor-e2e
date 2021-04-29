From ubuntu
RUN apt-get update
RUN apt-get install net-tools iptables -y
COPY e2e-agent /usr/local/bin
CMD ["e2e-agent"]
