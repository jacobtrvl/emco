[//]: # "SPDX-License-Identifier: Apache-2.0"
[//]: # "Copyright (c) 2019-2022 Intel Corporation"

# Build & Deploy

- [Build & Deploy](#build--deploy)
  - [Build from source](#build-from-source)
    - [Set up the environment](#set-up-the-environment)
    - [Update the base container images, if needed](#update-the-base-container-images-if-needed)
    - [Create the build base image in the EMCODOCKERREPO registry](#create-the-build-base-image-in-the-emcodockerrepo-registry)
    - [Registries and Images](#registries-and-images)
    - [Deploy EMCO locally](#deploy-emco-locally)
    - [Deploy EMCO in a Kubernetes cluster](#deploy-emco-in-a-kubernetes-cluster)
    - [Deploy sample test cases with EMCO](#deploy-sample-test-cases-with-emco)
  - [Top-level build system (Makefile)](#top-level-build-system-makefile)

## Build from source

### Set up the environment

Depending on your scenario, you may have to set different environment variables.
If behind an HTTP proxy for the Internet, make sure to set and export:

```
export HTTP_PROXY=${http_proxy}
export HTTPS_PROXY=${https_proxy}

```

If you are running a local container registry, typically using Docker on port TCP:5000, you don't need to set any other variables at this point.

Otherwise, the `EMCODOCKERREPO` variables needs to be set:

```
export EMCODOCKERREPO=${container_registry_url}/
```

Note that the value for the container registry URL must end with a `/`.

Or you may set the variable in `config/config.txt`, where the default value of `EMCODOCKERREPO` is already defined.

### Update the base container images, if needed

Additional external dependencies for EMCO, and other key parameters, are also captured in the configuration file `config/config.txt`.

The configuration file specifies the following important parameters:
  * `GO_VERSION`: The version of the Go programming language to use.
  * `HELM_VERSION`: The version of the Helm package manager to use.
  * `BUILD_BASE_IMAGE_NAME`: The name of the base image used for Helm
  * `BUILD_BASE_IMAGE_VERSION`: The version of the base image used for Helm
  * `SERVICE_BASE_IMAGE_NAME`: The name of the base image for each of the EMCO microservice images.
  * `SERVICE_BASE_IMAGE_VERSION`: The name and version of the base image used for each of the EMCO microservices' images.
  * `EMCODOCKERREPO`: The container registry URL (must end with a `/`) where newly-built EMCO container images should be pushed to (default: `localhost:5000/`).
  * `MAINDOCKERREPO`: The container registry URL (must end with a `/`), where existing container images should be pulled from (default is `""` which, which results in Docker Hub).

See the file `config/config.txt` for the current default values of all the parameters above.

### Create the build base image in the EMCODOCKERREPO registry

Run the following to create the final build container image and populate that
in the `EMCODOCKERREPO` registry.

```
make build-base
```

This will build the image that can then be used to create Helm charts for every EMCO microservice.

### Registries and Images
The image names outlined in `config/config.txt` will be searched in `MAINDOCKERREPO`.
`BUILD_BASE_IMAGE_NAME` which defaults to `emco-build-base` must be first built and pushed to `EMCODOCKERREPO`. Additionally, the `mongodb` and `etcd` images will also be pulled from `MAINDOCKERREPO`.

Both `EMCODOCKERREPO` and `MAINDOCKERREPO` can be set to the same exat container registry, in which case both pulling and pushing will always be done in the same registry.

Here are some of the images and versions that have been validated as of this writing:
  1.	alpine:3.12 (this is the default `SERVICE_BASE_IMAGE_*`, the base for EMCO microservice images)
  2.	golang:1.17.8 (for building EMCO Go components)
  2.	emco-build-base:1.3 (built after `make build-base`, builds the base image used for Helm - its base being alpine:3.12 like other EMCO microservice images)
  3.	mongo:4.x
  4.	etcd:3.x

### Deploy EMCO locally
You can build and deploy the EMCO components in your local environment (and
use them to deploy your workload in a set of remote Kubernetes clusters).

This is done in two stages:

 * Build the EMCO components:
   ```make all```
   This spawns a build container that generates the needed EMCO binaries and
   container images.
 * Deploy EMCO components locally:
   ```docker-compose up```
   using `deployments/docker/docker-compose.yml`. This spawns a set of
   containers, each running one EMCO component.

See [this tutorial](docs/user/Tutorial_Local_Install.md) for further details.

### Deploy EMCO in a Kubernetes cluster
Alternatively, you can build EMCO locally and deploy EMCO components in a
Kubernetes cluster using Helm charts (and use them to deploy your workload in
another set of Kubernetes clusters).

This requires the locally built container images to be pushed to the
container registry `EMCODOCKERREPO` with the appropriate tag, so that the
Kubernetes cluster can pull images from that container registry. The tag can
be set based on whether it is a developer/test build or a release build.

Do the following steps:

 * Set up the environment:

   * For development/testing:
     ```export BUILD_CAUSE=DEV_TEST```
     This sets the image tags to the form `${USER}-latest`.

   * For release:
     ```
        export BUILD_CAUSE=RELEASE
        export EMCOSRV_RELEASE_TAG=emco-${release_number}
     ```
     This sets the image tags to the specified tag. Note that if you set
     `BUILD_CAUSE=RELEASE` but do not set `EMCOSRV_RELEASE_TAG`, the image tags
     will be set to any tag defined on the git `HEAD` of the current git
     branch. If no git tag is defined on the HEAD, the build will fail.

 * Build and deploy EMCO:
   ```make deploy```

 * The helm charts and other build artefacts are put in a new folder called `bin/`
   - Helm chart: `bin/helm/`
   - EMCO control utility: `bin/emcoctl/emcoctl`

 * Transfer those files over to your target (assuming you built EMCO on a separate development system). To finalize your installation on the target:
   - Copy the `emcoctl` file somewhere on the path
   - Run the script in the `bin/helm` folder: `./emco-base-helm-install.sh install` (`./emco-base-helm-install.sh -h` for more details)

### Deploy sample test cases with EMCO
See [this Readme](examples/single-cluster/Readme.md) on how to setup an environment and run a few test cases with EMCO.

See [this tutorial](docs/user/install/Tutorial_Helm.md) for further details.


## Top-level build system (Makefile)

The following outlines and describes all important top-level `Makefile` targets. Each target has been expanded to include all the targets executed through dependency, iteratively until the last non-dependent target. As such, you will see that a great number of lines/steps are replicating throughout the targets outlined below. The list of targets and tasks executed is current as of EMCO 22.03.

`check-env:`

* enforces that `EMCODOCKERREPO` is set, for those target that really require it

`clean:`

* cleans up all compiled binaries
* does NOT clean up any artifacts, such as config.json, json-schemas and ref-schemas

`clean-all:`

* cleans up all compiled binaries
* removes the entire bin/ directories, which include config.json, json-schemas and ref-schemas

`test:`

* run unit tests for all services

`tidy:`

* runs `go mod tidy` for all services

`pre-compile: clean`

* cleans up all compiled binaries
* does NOT clean up any artifacts, such as config.json, json-schemas and ref-schemas
* replaces artifacts, such config.json, json-schemas and ref-schemas with default ones

`deploy-compile: check-env`

* enforces that `EMCODOCKERREPO` is set, for those target that really require it
* same as `make develop-compile`, it launches `make compile` inside a container (golang:v.vv):
    - cleans up all compiled binaries
    - does NOT clean up any artifacts, such as config.json, json-schemas and ref-schemas
    - replaces artifacts, such config.json, json-schemas and ref-schemas with default ones
    - compiles each of the EMCO binaries

`build-containers:`

* creates container images with each of the compiled EMCO binaries (emco-servicename:yy.mm extended from alpine:v.vv)

`develop-compile: check-env`

* enforces that `EMCODOCKERREPO` is set, for those target that really require it
* same as `make deploy-compile`, it launches `make compile` inside a container (golang:v.vv):
    - cleans up all compiled binaries
    - does NOT clean up any artifacts, such as config.json, json-schemas and ref-schemas
    - replaces artifacts, such config.json, json-schemas and ref-schemas with default ones
    - compiles each of the EMCO binaries
* but also sets `--env GOPATH=/repo/bin` in docker to reuse downloaded dependencies

`compile: pre-compile`

* cleans up all compiled binaries
* does NOT clean up any artifacts, such as config.json, json-schemas and ref-schemas
* replaces artifacts, such config.json, json-schemas and ref-schemas with default ones
* compiles each of the EMCO binaries

`deploy: check-env deploy-compile build-containers`

* enforces that `EMCODOCKERREPO` is set, for those target that really require it
* same as `make develop-compile`, it launches `make compile` inside a container (golang:v.vv):
    - cleans up all compiled binaries
    - does NOT clean up any artifacts, such as config.json, json-schemas and ref-schemas
    - replaces artifacts, such config.json, json-schemas and ref-schemas with default ones
    - compiles each of the EMCO binaries
* creates container images with each of the compiled EMCO binaries (emco-servicename:yy.mm extended from alpine:v.vv)
* creates helm charts (assumes build-base image exists even though it's not a checked dependency)
* pushes microservices to registry
* copies docker-compose files if BUILD_CAUSE set to DEV_TEST

`build-base:`

* creates build base image (emco-build-base:: extended from alpine:v.vv) and pushes it to registry

`develop: develop-compile build-containers`

* same as `make deploy-compile`, it launches `make compile` inside a container (golang:v.vv):
    - cleans up all compiled binaries
    - does NOT clean up any artifacts, such as config.json, json-schemas and ref-schemas
    - replaces artifacts, such config.json, json-schemas and ref-schemas with default ones
    - compiles each of the EMCO binaries
* but also sets `--env GOPATH=/repo/bin` in docker to reuse downloaded dependencies
* creates container images with each of the compiled EMCO binaries (emco-servicename:yy.mm extended from alpine:v.vv)
* fixes the names of some services
* pushes microservices to registry

`all: check-env compile build-containers`

* enforces that `EMCODOCKERREPO` is set, for those target that really require it
* launches `make compile` inside a container (golang:v.vv):
    - cleans up all compiled binaries
    - does NOT clean up any artifacts, such as config.json, json-schemas and ref-schemas
    - replaces artifacts, such config.json, json-schemas and ref-schemas with default ones
    - compiles each of the EMCO binaries
* creates container images with each of the compiled EMCO binaries (emco-servicename:yy.mm extended from alpine:v.vv)
* (nothing else is done)
