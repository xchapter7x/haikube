package docker

const (
	dockerFileHelmInstall = `
FROM dtzar/helm-kubectl
WORKDIR /root
COPY %s /root/.kube/config 
ENV KUBECONFIG /root/.kube/config
RUN helm init 
RUN helm install https://github.com/xchapter7x/haikube-chart/releases/download/v0.0.1/default.tgz --set image.repository=%s,image.tag=%s,service.internalPort=%s -n %s || \
helm upgrade %s https://github.com/xchapter7x/haikube-chart/releases/download/v0.0.1/default.tgz --set image.repository=%s,image.tag=%s,service.internalPort=%s
RUN helm ls
`
	dockerFileBuildpackLegacy = `
FROM %s
RUN mkdir /app /cache /deps || true
WORKDIR /app
COPY %s /app
RUN mv %s /buildpack
%s
RUN /buildpack/bin/detect /app
RUN /buildpack/bin/compile /app /cache
RUN /buildpack/bin/release
ENV PORT %s 
EXPOSE %s
CMD ["%s"]
`
	dockerFileBuildpackNew = `
FROM %s
RUN mkdir /app /cache /deps || true
WORKDIR /app
COPY %s /app 
RUN mv %s /buildpack
%s
RUN /buildpack/bin/detect /app
RUN /buildpack/bin/supply /app /cache /deps 0
RUN /buildpack/bin/finalize /app /cache /deps 0
RUN /buildpack/bin/release
ENV PORT %s 
EXPOSE %s
CMD ["%s"]
`
)
