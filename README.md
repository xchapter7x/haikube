# HaikuBe

Inspired by the haiku, but made for kubernetes.
So, (Haiku) meet (Kube)rnetes in Haikube

```
here is my source code
run it on the cloud for me
i do not care how
            - Onsi Fakhouri
            (https://content.pivotal.io/blog/pivotal-cloud-foundry-s-roadmap-for-2016)
```


## features
- `build`: build your container using a buildpack
- `push`: push your container to dockerhub
- `deploy`: generate the k8s deployment and deploy it
- `make`: push will create a docker image using your source code and 
        a buildpack. it will then upload it to a docker registry.
        after that it will generate a service manifest and a deployment
        manifest. finally it will apply those manifests to your k8s cluster.

## samples

```bash
# will build the image push it to docker registry and deploy the app to k8s
$ haikube build -f file.yaml
```

### sample .haikube.yaml

```yaml
---
name: unicornapp
org: xchapter7x
tag: 1.0.0
base: cloudfoundry/cflinuxfs2
buildpack: https://github.com/cloudfoundry/go-buildpack/releases/download/v1.8.22/go-buildpack-v1.8.22.zip
ports:
  - 80
env:
  - name: ENVKEY
    value: xxxxxxxxxxxxxxxxxxxxxxxxxxxx
  - name: BLAH
    value: xxxxxxxxxxxxxxxxxxxxxxxxxxxx
deployment-patch: {}
service-patch: {}
```
