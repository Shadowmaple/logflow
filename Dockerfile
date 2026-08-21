FROM docker.1ms.run/library/golang:1.26.6-alpine3.24 AS build

RUN apk update && apk add make

WORKDIR /logflow
COPY . .

ENV GOPROXY=https://goproxy.cn,direct
RUN make

FROM docker.1ms.run/library/alpine:3.24

ARG TZ="Asia/Shanghai"
ENV TZ=${TZ}

COPY --from=build /logflow/logflow /usr/local/bin/
