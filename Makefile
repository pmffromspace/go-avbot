#Dockerfile vars

#vars
IMAGENAME=go-avbot
TAG=v0.0.7
BUILDDATE=${shell date -u +%Y-%m-%dT%H:%M:%SZ}
IMAGEFULLNAME=avhost/${IMAGENAME}
BRANCH=${shell git symbolic-ref --short HEAD}
LASTCOMMIT=$(shell git log -1 --pretty=short | tail -n 1 | tr -d " " | tr -d "UPDATE:")

.PHONY: help build bootstrap all docs publish push version

help:
	    @echo "Makefile arguments:"
	    @echo ""
	    @echo "Makefile commands:"
			@echo "push"
	    @echo "build"
			@echo "build-bin"
	    @echo "all"
			@echo "docs"
			@echo "publish"
			@echo "version"
			@echo ${TAG}

.DEFAULT_GOAL := all

ifeq (${BRANCH}, master) 
        BRANCH=latest
endif

ifneq ($(shell echo $(LASTCOMMIT) | grep -E '^v([0-9]+\.){0,2}(\*|[0-9]+)'),)
        BRANCH=${LASTCOMMIT}
else
        BRANCH=latest
endif

build:
	@echo ">>>> Build docker image"
	@docker build --build-arg TAG=${TAG} --build-arg BUILDDATE=${BUILDDATE} -t ${IMAGEFULLNAME}:${BRANCH} .

build-bin:
	@echo ">>>> Build binary"
	@CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags "-X main.BuildVersion=${BUILDDATE} -X main.GitVersion=${TAG} -extldflags \"-static\"" .

push:
	@echo ">>>> Publish docker image: " ${BRANCH}
	@docker buildx create --use --name buildkit
	@docker buildx build --platform linux/arm64,linux/amd64 --push --build-arg TAG=${TAG} --build-arg BUILDDATE=${BUILDDATE} -t ${IMAGEFULLNAME}:${BRANCH} .
	@docker buildx rm buildkit

update-gomod:
	go get -u
	go mod tidy	
	go mod vendor

seccheck:
	grype --add-cpes-if-none .
	trivy image ${IMAGEFULLNAME}:${BRANCH}

sboom:
	syft dir:. > sbom.txt
	syft dir:. -o json > sbom.json

all: build seccheck sboom
