FROM golang:1.21

WORKDIR /redis-cluster-acl-tool
COPY ./ /redis-cluster-acl-tool
RUN go env -w GOPROXY=https://goproxy.cn,direct && make all

FROM redis:7.0.14
RUN sed -i 's/deb.debian.org/mirrors.cloud.tencent.com/g' /etc/apt/sources.list.d/debian.sources  \
    && apt update  \
    && apt install -y ca-certificates  \
    && update-ca-certificates \
    && apt install -y pwgen

COPY --from=0 /redis-cluster-acl-tool/acltool /usr/bin

CMD ["sleep","1"]