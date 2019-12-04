# Installing repo-operator

:white_check_mark: Pre-Requisite
* [go](https://golang.org/dl/) version v1.13+.
* [docker](https://docs.docker.com/install/) version 17.03+ or [podman](https://github.com/containers/libpod).
* [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) and access to a [Kubernetes](https://kubernetes.io/docs/concepts/configuration/organize-cluster-access-kubeconfig/) cluster version 1.11 or later.
* [kustomize](https://sigs.k8s.io/kustomize/docs/INSTALL.md) v3.1.0+.
* [Kubebuilder](https://book.kubebuilder.io/quick-start.html) (Required only if you want to make changes to the CRD structure)
* GNU Make
```sudo apt-get install build-essential```
* Artifactory-Pro 6.12.2+ (If you want to actually manage repositories with the operator) 

## Local development
To run repo-operator for testing during local development, you can launch it outside k8s. 
:rocket:
```
# install CRD into cluster
make install
# run the operator locally against configured cluster
make run
``` 

## Install to cluster

To run in a k8s cluster, repo-operator needs to be built as a container image and deployed with some additional resources.
1. _**Build Container image**_
```bash
# set image tag if you need to push to remote registry
export IMG=my.registry.host.domain/repo/image:tag
# build image
make image-build
``` 
> ¤ The default image tag is ```repo-operator:latest```, for local development you can probably leave it at that.   
  ¤ Set **GOPROXY** variable in the Dockerfile if you are running behind firewall.    
  ¤ ```make image-push``` if you need to push the image to docker registry accessible from your kubernetes cluster.   

2. _**Install repo-operator into the Kubernetes cluster**_ 
```
make deploy
```   
> ¤ You need to have cluster-admin role to perform this task.           
  ¤ This will install the ```repository``` CRD and deploy repo-operator with required additional config into a new namespace called ```repo-operator-system``` in the cluster configured in ~/.kube/config

:point_right: You're now all set up to [use repository resources](using.md)

:point_left: Back to [Home](../README.md)