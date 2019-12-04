
# Image URL to use all building/pushing image targets
IMG ?= repo-operator:latest
ARTIFACTORY_URL ?= https://your-artifactory-server-url/artifactory
ARTIFACTORY_USER ?= user
ARTIFACTORY_PASSWORD ?= password

# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:trivialVersions=true"

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

ifeq (,$(shell which podman))
  CONTAINER_TOOL=$(shell which docker)
else 
  CONTAINER_TOOL=$(shell which podman)
endif

ifeq (,$(CONTAINER_TOOL))
  $(error Missing podman or docker in PATH)
endif

all: manager

# Run tests
test: generate fmt vet manifests
	go test ./... -coverprofile cover.out

# Build manager binary
manager: generate fmt vet
	go build -o bin/manager main.go

# Run against the configured Kubernetes cluster in ~/.kube/config
run: generate fmt vet manifests
	go run ./main.go

# Install CRDs into a cluster
install: manifests
	kustomize build config/crd | kubectl apply -f -

# Deploy repo-operator in the configured Kubernetes cluster in ~/.kube/config
deploy: manifests
	cd config/operator && kustomize edit set image controller=${IMG} \
	&& kustomize edit add secret repo-secret --from-literal=password=$(ARTIFACTORY_PASSWORD) \
	--from-literal=username=$(ARTIFACTORY_USER) --from-literal=url=$(ARTIFACTORY_URL)
	kustomize build config/default | kubectl apply -f -
	# Don't leave secret literal values on file
	cd config/operator && rm kustomization.yaml && cp .ktemplate kustomization.yaml

# Generate manifests e.g. CRD, RBAC etc.
manifests: controller-gen
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=repo-operator webhook paths="./..." output:crd:artifacts:config=config/crd/bases

# Run go fmt against code
fmt:
	go fmt ./...

# Run go vet against code
vet:
	go vet ./...

# Generate code
generate: controller-gen
	$(CONTROLLER_GEN) object:headerFile=./hack/boilerplate.go.txt paths="./..."

# Build the docker image
image-build: test
	${CONTAINER_TOOL} build . -t ${IMG}

# Push the docker image
image-push:
	${CONTAINER_TOOL} push ${IMG}

# find or download controller-gen
# download controller-gen if necessary
controller-gen:
ifeq (, $(shell which controller-gen))
	@{ \
	set -e ;\
	CONTROLLER_GEN_TMP_DIR=$$(mktemp -d) ;\
	cd $$CONTROLLER_GEN_TMP_DIR ;\
	go mod init tmp ;\
	go get sigs.k8s.io/controller-tools/cmd/controller-gen@v0.2.2 ;\
	rm -rf $$CONTROLLER_GEN_TMP_DIR ;\
	}
CONTROLLER_GEN=$(GOBIN)/controller-gen
else
CONTROLLER_GEN=$(shell which controller-gen)
endif
