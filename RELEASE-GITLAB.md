# Release steps for emco-base on GitLab

This document outlines the release process for emco-base on GitLab.
The version as of this writing is `22.03.1`. Please modify to the correct version according to the release undertaken.

## Tagging the source code

    git tag v22.03.1
    git push --tags

## Pushing container images

    cd emco-base # this is the git repo dir
    docker login registry.gitlab.com

    export BUILD_CAUSE=RELEASE
    export EMCOSRV_RELEASE_TAG=v22.03.1
    export EMCODOCKERREPO=registry.gitlab.com/project-emco/core/emco-base/
    make deploy

## Updating Helm charts

Usually, there is no need to update the public Helm charts hosted on the GitLab Package Registry, since they are designed to be compatible with any verison of EMCO hosted on the GitLab Container Registry.
However, if indeed there was a bump the any Helm chart/package `version`, including when new microservices get added (which require a bump to the `version` of the encompassing package, such as `emco` or `emco-services`), then here is how to generate and upload new packages:

    PROJECT_ACCESS_TOKEN=""
    cd emco-base/deployments/helm/emcoBase
    make clean
    make all
    curl --request POST --form chart=@dist/packages/emco-db-1.0.0.tgz --user emco-base:$PROJECT_ACCESS_TOKEN https://gitlab.com/api/v4/projects/29353813/packages/helm/api/stable/charts
    curl --request POST --form chart=@dist/packages/emco-tools-1.0.0.tgz --user emco-base:$PROJECT_ACCESS_TOKEN https://gitlab.com/api/v4/projects/29353813/packages/helm/api/stable/charts
    curl --request POST --form chart=@dist/packages/emco-services-1.0.0.tgz --user emco-base:$PROJECT_ACCESS_TOKEN https://gitlab.com/api/v4/projects/29353813/packages/helm/api/stable/charts
    curl --request POST --form chart=@dist/packages/emco-1.0.0.tgz --user emco-base:$PROJECT_ACCESS_TOKEN https://gitlab.com/api/v4/projects/29353813/packages/helm/api/stable/charts
    cd ..
    helm package monitor
    curl --request POST --form chart=@monitor-1.0.0.tgz --user emco-base:$PROJECT_ACCESS_TOKEN https://gitlab.com/api/v4/projects/29353813/packages/helm/api/stable/charts

The commands above are an example for the already-released Helm charts' `version: 1.0.0`.
Make sure to replace set `$PROJECT_ACCESS_TOKEN` to a [GitLab project access token](https://docs.gitlab.com/ee/user/project/settings/project_access_tokens.html) for EMCO.

Currently only the `stable` channel of the package registry is being used, but we may decide to have multiple channels in the future once we need to support charts of different versions that shouldn't clash with each other (for example to prevent a stable version from upgrading to another stable version, and instead only upgrade to point releases).

## Creating release page and tarballs

Go to the [GitLab Releases page](https://gitlab.com/project-emco/core/emco-base/-/releases), and click *New release*.
Choose the tag name `v22.03.1` created earlier. Name the release title simply as `22.03.1`. Choose the corresponding milestone, i.e. 22.03.1. Finally, write the release notes, add any extra release assets, and click *Create release*.
