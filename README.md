# HaikuBe
### If you like the Cloud Foundry and Heroku 'push' experiences, but are using Kubernetes, try this!

---


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

## Overview

- haikube takes the same approach as cloudfoundry & heroku and uses buildpacks to create a working image from just your code
  - buildpacks can detect if your code is supported, and then can create a fully functional application container image from just your code
  - buildpacks will run any required scripts/processes to make your code ready to be deployed (ie. ruby bundle, go dep, etc)
- haikube can use any dockerhost its configured for to build the container images
- haikube can store the container image in dockerhub by default, but can be configured to use any docker registry
- haikube can create a k8s deployment using the created docker image
- haikube can create a k8s service which points to the created deployment

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
$ hk build -c haikube.yml -s pathtosource


# will build the image & upload it to docker registry
$ hk upload -c haikube.yml -s pathtosource
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

- `DOCKER_HOST`: to set the url to the docker server. **(optional)**
- `DOCKER_API_VERSION`: to set the version of the API to reach, leave empty for latest. **(optional)**
- `DOCKER_CERT_PATH`: to load the TLS certificates from. **(optional)**
- `DOCKER_TLS_VERIFY`: to enable or disable TLS verification, off by default. **(optional)**
- `DOCKER_USERNAME`: to set which account to use when uploading to docker registry. **(required for upload calls)**
- `DOCKER_PASSWORD`: to set password of docker registry account. **(required for upload calls)**
- `DOCKER_REGISTRY_URL`: to set a private registry host (optional: dockerhub is used by default) **(optional)**


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

#### Test related toggles
- `K8S_CLUSTER`: setting this var to `false` when running tests will disable integration tests and e2e tests which require a kubernetes cluster to succeed. If you do not have a k8s cluster to test against, or are not testing functionality that requires a cluster, set this variable to false.



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

## Similar tools in the space
- [Helm](https://helm.sh) : The Kubernetes Package Manager
  - helm is a great tool, but it plays in a different space than haikube.
  - helm is meant to bring structure to templatizing, versioning and distributing your k8s deployments. Haikube explicitely is meant for those
  who do not care how their code is run on k8s, they just want to push it and have it work. One can use haikube to generate an image, which gets
  deployed via a helm chart.
- [Skaffold](https://github.com/GoogleContainerTools/skaffold) : Skaffold is a command line tool that facilitates continuous development for Kubernetes applications
  - More similar to haikube, but takes a different approach. Skaffold is a sort of framework to organize and structure how you create your images, 
  what your deployments, services, etc look like. You can even use it with helm. Haikube isnt a framework, it assumes you dont want to worry about
  how to build your images or architect your deployments. Haikube forms its own opinion, so you don't have to, where as Skaffold gives engineers
  who wish to build everything themselves a standard structure and tooling to play safely. It's a great tool, but meets the needs of a different 
  user persona.
