
We share all our software under OpenSource License. If you like what we are doing, it would be nice if you donate a small fee. 

[![Donate](https://liberapay.com/assets/widgets/donate.svg)](https://liberapay.com/AVENTER/donate)

Our Kanban Board is [here](https://wekan.aventer.biz/b/XePZjKD4mK3eFY8MS/go-avbot)

# License
go-neb is under the Apache License. To make it more complicated, our code are under GPL. These are:
- aws (services/aws)
- invoice (services/invoice)


# What is go-avbot

AVBOT is our digital working partner. He is helping us with our daily business. From creating and sending out invoices to install server applications. AVBOT is based on [go-neb](https://github.com/matrix-org/go-neb), a matrix BOT developed in golang. 

# Features

### AWS
- Userspecified Access Key configuration
- Start/Stop of AWS instances
- Show AWS List and infos of all AWS instances

### Github
- Monitor repositories and create webhooks of them
- Create issues to repos

### Ispconfig
- Create Invoice and send them out
- Show invoices of a user

There are still a lot of work. Currently our main focus is the AWS support.

# Software Requirements

```
go get github.com/Sirupsen/logrus 
go get github.com/matrix-org/util
go get github.com/mattn/go-sqlite3
go get github.com/prometheus/client_golang/prometheus
go get github.com/matrix-org/dugong
go get github.com/matrix-org/gomatrix
go get github.com/mattn/go-shellwords
go get gopkg.in/yaml.v2
go get golang.org/x/oauth2
go get github.com/google/go-github/github
go get github.com/sirupsen/logrus
go get gopkg.in/alecthomas/kingpin.v2
go get github.com/russross/blackfriday
go get github.com/aws/aws-sdk-go
```

# API Documentation

- Matrix API: https://www.matrix.org/docs/spec/r0.0.0/client_server.html
- AWS API: https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/setting-up.html