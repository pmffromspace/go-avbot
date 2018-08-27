FROM golang:1.10.0-alpine3.7
LABEL maintainer="Andreas Peters <support@aventer.biz>"

ENV BIND_ADDRESS=:4050 DATABASE_TYPE=sqlite3 DATABASE_URL=/go-avbot/data/go-neb.db?_busy_timeout=5000 

COPY . /go-avbot/
COPY run.sh /run.sh

RUN apk update && \
    apk add git gcc libc-dev && \
    cd /go-avbot/ && \
    go get -d && \
    mkdir -p /go-avbot/log

VOLUME /go-avbot/data

RUN cd /go-avbot && \
    go build -ldflags "-X main.MinVersion=`date -u +%Y%m%d%.H%M%S`" app.go init.go

EXPOSE 4050

ENTRYPOINT ["/run.sh"]
#CMD ["/bin/sh"]
