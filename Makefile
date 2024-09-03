# Image URL to use all building/pushing image targets
REGISTRY ?= tanzu-build-docker-prod-local.usw1.packages.broadcom.com
REPOSITORY ?= build-controller

VERSION ?= 0.0.0-dev
TAG ?= 0.0.0-dev
BASE_IMAGE_WITH_TAG ?= "paketobuildpacks/run-jammy-tiny:latest"

# Suppress kapp prompts with KAPP_ARGS="--yes"
KAPP_ARGS ?= "--yes=false"

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
GOOS=$(shell go env GOOS)
else
GOBIN=$(shell go env GOBIN)
GOOS=$(shell go env GOOS)
endif

# Setting SHELL to bash allows bash commands to be executed by recipes.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

.PHONY: all
all: build

##@ General

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'. The awk command is responsible for reading the
# entire set of makefiles included in this invocation, looking for lines of the
# file as xyz: ## something, and then pretty-format the target and help. Then,
# if there's a line with ##@ something, that gets pretty-printed as a category.
# More info on the usage of ANSI control characters for terminal formatting:
# https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_parameters
# More info on the awk command:
# http://linuxcommand.org/lc3_adv_awk.php

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

.PHONY: manifests
manifests: controller-gen ## Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects.
	$(CONTROLLER_GEN) rbac:roleName=manager-role crd webhook paths="./..." output:crd:artifacts:config=config/crd/bases

.PHONY: generate
generate: diegen controller-gen ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."
	$(DIEGEN) die:headerFile=./hack/boilerplate.go.txt paths="./..."

.PHONY: fmt
fmt: ## Run go fmt against code.
	go fmt ./...

.PHONY: vet
vet: ## Run go vet against code.
	go vet ./...

.PHONY: test
test: manifests generate fmt vet ## Run unit tests only.
	go test ./... -short -coverprofile cover.out

.PHONY: lint
lint: golangci-lint ## Run golangci-lint linter
	$(GOLANGCI_LINT) run

.PHONY: lint-fix
lint-fix: golangci-lint ## Run golangci-lint linter and perform fixes
	$(GOLANGCI_LINT) run --fix

##@ Build

.PHONY: ko-yaml
ko-yaml: ko-setup ## Generates .ko.yaml
	$(YTT) -f build-templates/ko.yml -f build-templates/values-schema.yaml -v build.base_image=$(BASE_IMAGE_WITH_TAG) > .ko.yaml

.PHONY: build
build: manifests generate fmt vet ## Build manager binary.
	go build -o bin/manager cmd/controller/main.go

.PHONY: run
run: manifests generate fmt vet test ## Run a controller from your host.
	go run ./cmd/controller/main.go

.PHONY: dist
dist: manifests generate kustomize ## Generate a consolidated YAML with CRDs and deployment.
	mkdir -p dist
	$(KUSTOMIZE) build config/default > dist/image-sync-controller.yaml

.PHONY: package-build
package-build: dist ko-yaml carvel-tools ## Generates builds and pushes to registry with 0.0.0-dev version/tag. Requires carvel-tools
	$(YTT) -f build-templates/kbld-config.yaml -f build-templates/values-schema.yaml -v build.registry_host=$(REGISTRY) -v build.registry_project=$(REPOSITORY) > kbld-config.yaml
	$(YTT) -f build-templates/package-build.yml -f build-templates/values-schema.yaml -v build.registry_host=$(REGISTRY) -v build.registry_project=$(REPOSITORY) > package-build.yml
	$(YTT) -f build-templates/package-resources.yml > package-resources.yml
	KO_CONFIG_PATH="$(shell pwd)/.ko.yaml" $(KCTRL) package release -v $(VERSION) -t $(VERSION) --chdir=$(shell pwd) --debug -y

##@ Deployment

ifndef ignore-not-found
  ignore-not-found = false
endif

.PHONY: install
install: manifests kustomize ## Install CRDs into the K8s cluster specified in ~/.kube/config.
	$(KAPP) deploy -a image-sync-controller -n kube-system -f <($(KUSTOMIZE) build config/crd) $(KAPP_ARGS)

.PHONY: uninstall
uninstall: manifests kustomize ## Uninstall CRDs from the K8s cluster specified in ~/.kube/config. Call with ignore-not-found=true to ignore resource not found errors during deletion.
	$(KAPP) delete -a image-sync-controller -n kube-system $(KAPP_ARGS)

.PHONY: sample
sample:
	$(KAPP) deploy -a sample -f config/samples/v1* $(KAPP_ARGS)

.PHONY: deploy
deploy: dist ko-yaml ## Deploy controller to the K8s cluster specified in ~/.kube/config.
	$(YTT) -f build-templates/ko.yml -f build-templates/values-schema.yaml -v build.base_image=$(BASE_IMAGE_WITH_TAG) > .ko.yaml
	$(KAPP) deploy -a image-sync-controller -n kube-system -f <( $(KO) resolve -f dist/image-sync-controller.yaml) $(KAPP_ARGS)

.PHONY: undeploy
undeploy: kustomize ## Undeploy controller from the K8s cluster specified in ~/.kube/config. Call with ignore-not-found=true to ignore resource not found errors during deletion.
	$(KAPP) delete -a image-sync-controller -n kube-system $(KAPP_ARGS)

.PHONY: package-deploy
package-deploy: ## Deploys carvel package of the controller to the K8s cluster specified in ~/.kube/config
	@if [ -d carvel-artifacts ]; then \
  		$(KAPP) deploy -a image-sync-controller-pkg -f carvel-artifacts/ -n kube-system; \
		$(KCTRL) package install -i image-sync-controller -p controller.build.tanzu.vmware.com -v $(VERSION) -n kube-system; \
	fi;

.PHONY: package-undeploy
package-undeploy: ## Deletes image-sync-controller package from the K8s cluster specified in ~/.kube/config
	$(KCTRL) package installed delete -i image-sync-controller -n kube-system
	$(KAPP) delete -a image-sync-controller-pkg -n kube-system $(KAPP_ARGS)
##@ Dependencies

## Location to install dependencies to
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

##@ Tool Binaries
KUBECTL ?= kubectl
YTT ?= $(LOCALBIN)/ytt
KCTRL ?= $(LOCALBIN)/kctrl
KAPP ?= $(LOCALBIN)/kapp
IMGPKG ?= $(LOCALBIN)/imgpkg
KUSTOMIZE ?= $(LOCALBIN)/kustomize-$(KUSTOMIZE_VERSION)
CONTROLLER_GEN ?= $(LOCALBIN)/controller-gen-$(CONTROLLER_TOOLS_VERSION)
GOLANGCI_LINT = $(LOCALBIN)/golangci-lint-$(GOLANGCI_LINT_VERSION)
KO ?= $(LOCALBIN)/ko
DIEGEN ?= $(LOCALBIN)/diegen

## Tool Versions
KUSTOMIZE_VERSION ?= v5.4.1
CONTROLLER_TOOLS_VERSION ?= v0.16.2
GOLANGCI_LINT_VERSION ?= v1.57.2
KO_VERSION ?= 0.15.4
DIEGEN_VERSION=v0.13.0
GOOS ?= darwin

.PHONY: kustomize
kustomize: $(KUSTOMIZE) ## Download kustomize locally if necessary.
$(KUSTOMIZE): $(LOCALBIN)
	$(call go-install-tool,$(KUSTOMIZE),sigs.k8s.io/kustomize/kustomize/v5,$(KUSTOMIZE_VERSION))

.PHONY: controller-gen
controller-gen: $(CONTROLLER_GEN) ## Download controller-gen locally if necessary.
$(CONTROLLER_GEN): $(LOCALBIN)
	$(call go-install-tool,$(CONTROLLER_GEN),sigs.k8s.io/controller-tools/cmd/controller-gen,$(CONTROLLER_TOOLS_VERSION))

.PHONY: golangci-lint
golangci-lint: $(GOLANGCI_LINT) ## Download golangci-lint locally if necessary.
$(GOLANGCI_LINT): $(LOCALBIN)
	$(call go-install-tool,$(GOLANGCI_LINT),github.com/golangci/golangci-lint/cmd/golangci-lint,${GOLANGCI_LINT_VERSION})

.PHONY: carvel-tools
carvel-tools: $(LOCALBIN) ## Downloads Carvel CLI tools locally
	if [[ ! -f $(YTT) ]]; then \
		curl -L https://carvel.dev/install.sh | K14SIO_INSTALL_BIN_DIR=$(LOCALBIN) bash; \
	fi

.PHONY: ko-setup
ko-setup: $(KO) ## Setup for ko binary
$(KO): $(LOCALBIN)
	@if [ ! -f $(KO) ]; then \
		echo curl -sSfL "https://github.com/ko-build/ko/releases/download/v$(KO_VERSION)/ko_$(KO_VERSION)_$(GOOS)_x86_64.tar.gz"; \
		curl -sSfL "https://github.com/ko-build/ko/releases/download/v$(KO_VERSION)/ko_$(KO_VERSION)_$(GOOS)_x86_64.tar.gz" > $(LOCALBIN)/ko.tar.gz; \
		tar xzf $(LOCALBIN)/ko.tar.gz -C $(LOCALBIN)/; \
		chmod +x $(LOCALBIN)/ko; \
	fi;

.PHONY: diegen
diegen: $(DIEGEN) ## Download die-gen locally
$(DIEGEN): $(LOCALBIN)
	@echo "# installing $(@)"
	GOBIN=$(LOCALBIN) go install reconciler.io/dies/diegen@$(DIEGEN_VERSION)

.PHONY: clean
clean: ## Remove local downloaded and generated binaries
	rm -rf $(LOCALBIN)
	fi

# go-install-tool will 'go install' any package with custom target and name of binary, if it doesn't exist
# $1 - target path with name of binary (ideally with version)
# $2 - package url which can be installed
# $3 - specific version of package
define go-install-tool
@[ -f $(1) ] || { \
set -e; \
package=$(2)@$(3) ;\
echo "Downloading $${package}" ;\
GOBIN=$(LOCALBIN) go install $${package} ;\
mv "$$(echo "$(1)" | sed "s/-$(3)$$//")" $(1) ;\
}
endef

