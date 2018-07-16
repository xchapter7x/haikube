# HaikuBe

[![CircleCI](https://circleci.com/gh/xchapter7x/haikube/tree/master.svg?style=svg)](https://circleci.com/gh/xchapter7x/haikube/tree/master)

Inspired by the haiku, but made for kubernetes.
So, (Haiku) meet (Kube)rnetes in Haikube

```
here is my source code
run it on the cloud for me
i do not care how
            - Onsi Fakhouri
            (https://content.pivotal.io/blog/pivotal-cloud-foundry-s-roadmap-for-2016)
```

## Download

### Binaries available for linux, osx & windows
https://github.com/xchapter7x/haikube/releases/latest


## features
- `build`: build your container using a buildpack
- `upload`: push your container to dockerhub
- `deploy`: generate the k8s deployment and deploy it
- `push`: push will create a docker image using your source code and 
        a buildpack. it will then upload it to a docker registry.
        after that it will generate a service manifest and a deployment
        manifest. finally it will apply those manifests to your k8s cluster.

## samples

```bash
# will build the image using a buildpack
$ "hk build -c haikube.yml -s pathtosource


# will build the image & upload it to docker registry
$ "hk upload -c haikube.yml -s pathtosource
```

### sample .haikube.yaml

```yaml
---
name: unicornapp
instances: 4
image: xchapter7x/myapp
tag: 1.0.0
baseimage: cloudfoundry/cflinuxfs2
buildpack: https://github.com/cloudfoundry/go-buildpack/releases/download/v1.8.22/go-buildpack-v1.8.22.zip
ports:
  - 80
env:
  CF_STACK: cflinuxfs2
  GOPACKAGENAME: main

```

## HK Usage

### some local env vars you can set 

`DOCKER_HOST`: to set the url to the docker server.
`DOCKER_API_VERSION`: to set the version of the API to reach, leave empty for latest.
`DOCKER_CERT_PATH`: to load the TLS certificates from.
`DOCKER_TLS_VERIFY`: to enable or disable TLS verification, off by default.
`DOCKER_USERNAME`: to set which account to use when uploading to docker registry.
`DOCKER_PASSWORD`: to set password of docker registry account.
`DOCKER_REGISTRY_URL`: to set a private registry host (optional: dockerhub is used by default)


## Contributions

### Simply send over a PR or submit an issue

### Running the tests
```bash
# run everything
$ make test

# run units only (no external deps)
$ make unit

# run e2e tests (will likely require docker and k8s configured on your machine)
$ make e2e

# run integrations only ( will likely require docker and k8s configured on your machine)
$ make integrations
```

### build the binaries
```bash

# build cross all platform
$ make build

# build individual platform
$ make build-(darwin|linux|win)
```

### updating dependencies
```bash
$ make dep
```
