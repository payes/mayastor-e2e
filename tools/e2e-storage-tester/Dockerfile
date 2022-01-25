FROM alpine
MAINTAINER mayastor mayastor@mayadata.io
RUN apk --no-cache add gcc musl-dev
ADD e2e-storage-tester.c /
RUN gcc -Wall -static -o /e2e-storage-tester e2e-storage-tester.c
CMD ["sleep","100000"]
