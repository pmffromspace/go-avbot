FROM golang:alpine as builder

WORKDIR /build

COPY . /build/

RUN apk add git gcc musl-dev && \
    go get -d

RUN GOOS=linux go build -ldflags "-X main.MinVersion=`date -u +%Y%m%d%.H%M%S` -extldflags \"-static\"" -o main app.go init.go


FROM alpine
LABEL maintainer="Andreas Peters <support@aventer.biz>"

ENV BIND_ADDRESS=:4050 DATABASE_TYPE=sqlite3 DATABASE_URL=/go-avbot/data/go-neb.db?_busy_timeout=5000 


RUN adduser -S -D -H -h /go-avbot appuser && \
    mkdir /go-avbot && \
    chown appuser: /go-avbot && \
    chmod 755 /go-avbot

USER appuser

WORKDIR "/go-avbot"

COPY --from=builder /build/main /go-avbot/
COPY run.sh /run.sh

RUN mkdir -p /go-avbot/log

VOLUME /go-avbot/data

EXPOSE 4050

ENTRYPOINT ["/run.sh"]
