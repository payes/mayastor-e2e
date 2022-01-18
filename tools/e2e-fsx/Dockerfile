FROM alpine:3.14 AS builder

RUN apk --update add git automake autoconf build-base pkgconf

RUN git clone https://github.com/linux-test-project/ltp.git && \
    cd ltp && \
    make autotools && \
    ./configure && \
    cd testcases/kernel/fs/fsx-linux && \
    make

FROM alpine:3.14

RUN apk --update add e2fsprogs util-linux

COPY --from=builder /ltp/testcases/kernel/fs/fsx-linux/fsx-linux /usr/bin
ADD fsxtest.sh /
RUN chmod +x /fsxtest.sh

ENTRYPOINT ["/bin/sh", "/fsxtest.sh"]