FROM alpine:3.6
LABEL maintainer="Andreas Peters <support@aventer.biz>"

ENV BIND_ADDRESS=:4050 DATABASE_TYPE=sqlite3 DATABASE_URL=/go-avbot/data/go-neb.db?_busy_timeout=5000 

ARG BRANCH=v0.0.3

RUN apk add --update git go gcc g++  && \
    go get github.com/sirupsen/logrus && \
    go get github.com/matrix-org/util && \
    go get github.com/mattn/go-sqlite3 && \
    go get github.com/prometheus/client_golang/prometheus && \
    go get github.com/matrix-org/dugong && \
    go get github.com/AVENTER-UG/gomatrix && \
    go get github.com/mattn/go-shellwords && \
    go get gopkg.in/yaml.v2 && \
    go get golang.org/x/oauth2 && \
    go get github.com/google/go-github/github && \
    go get gopkg.in/alecthomas/kingpin.v2 && \
    go get github.com/russross/blackfriday && \
    go get github.com/aws/aws-sdk-go && \
    go get github.com/golang/lint/golint && \
    go get github.com/fzipp/gocyclo && \
    go get github.com/client9/misspell/... && \
    go get github.com/gordonklaus/ineffassign && \     
    mkdir -p /go-avbot/log

VOLUME /go-avbot/data

COPY . /go-avbot/
COPY run.sh /run.sh

EXPOSE 4050

#ENTRYPOINT ["/run.sh"]
CMD ["/bin/sh"]
