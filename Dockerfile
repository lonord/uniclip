FROM --platform=$BUILDPLATFORM golang:alpine AS builder

ARG TARGETPLATFORM
ARG BUILDPLATFORM
ARG TARGETOS
ARG TARGETARCH

COPY . /app

ENV GOPRIVATE=git.lonord.name
ENV GOPROXY=https://goproxy.cn,direct
ENV GOOS=$TARGETOS
ENV GOARCH=$TARGETARCH

RUN cd /app/server \
    && go build -ldflags "-s -w" -o uniclip

FROM alpine

COPY --from=builder /app/server/uniclip /app/uniclip

RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories \
    && apk add ca-certificates \
    && rm -rf /var/cache/apk/* \
    && update-ca-certificates

WORKDIR /app

EXPOSE 8080

CMD [ "/app/uniclip" ]