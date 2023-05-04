# go-avbot - the aventer bot

## What is go-avbot

AVBOT is our digital working partner. He is helping us with our daily business. From creating and sending out invoices to install server applications. AVBOT is based on [go-neb](https://github.com/matrix-org/go-neb), a matrix BOT developed in golang.


## How to use it?

First we have to create a config.yaml inside of data directory that we have to mount into the container. A sample of these config can be found in our Github repository.

```bash
docker run -v ./data:/go-avbot/data:rw avhost/go-avbot:latest 
```

## License

go-neb is under the Apache License. To make it more complicated, our code are under GPL. These are:

- aws (services/aws)
- invoice (services/invoice)
- pentest (services/pentest)

## Features

### AWS

- Start/Stop of AWS instances
- Show list of all instances in all regions
- Create Instances
- Search AMI's

### Ispconfig

- Create Invoice and send them out
- Show invoices of a user

### Pentest

- Penetrate a server target
- Create a report about the penetrations test result and upload it into the chat room

There are still a lot of work. Currently our main focus is the AWS support.

### Github

- Receive Webhooks from your github repositories.
- Create Issues

### Travis-CI

- Receive Webhooks from your travis account

### Wekan

- Receive Webhooks from your wekan boards

### Gitea

- Receive Webhooks from your gitea repo

### NLP (Natural Language Processing) 

- Gateway to the IKY Framework

## Software Requirements

```bash
go get github.com/sirupsen/logrus
go get github.com/matrix-org/util
go get github.com/mattn/go-sqlite3
go get github.com/prometheus/client_golang/prometheus
go get github.com/matrix-org/dugong
go get git.aventer.biz/AVENTER/gomatrix
go get github.com/mattn/go-shellwords
go get gopkg.in/yaml.v2
go get golang.org/x/oauth2
go get github.com/google/go-github/github
go get gopkg.in/alecthomas/kingpin.v2
go get github.com/russross/blackfriday
go get github.com/aws/aws-sdk-go
```

## API Documentation

- [Matrix API](https://www.matrix.org/docs/spec/r0.0.0/client_server.html)
- [AWS API](https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/setting-up.html)
- [OpenVAS](https://docs.greenbone.net/API/GMP/gmp-20.08.html)
