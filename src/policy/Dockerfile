ARG MAINDOCKERREPO
ARG SERVICE_BASE_IMAGE_NAME
ARG SERVICE_BASE_IMAGE_VERSION

#FROM ${MAINDOCKERREPO}${SERVICE_BASE_IMAGE_NAME}:${SERVICE_BASE_IMAGE_VERSION}
FROM jacobaarna/alpine:3.12

WORKDIR /opt/emco/policy

RUN addgroup -S emco && adduser -S -G emco emco
RUN chown emco:emco . -R

COPY --chown=emco ./sacontroller ./
COPY --chown=emco ./config.json ./
COPY --chown=emco ./json-schemas ./json-schemas
COPY --chown=emco ./ref-schemas ./ref-schemas

USER emco

ENTRYPOINT ["./sacontroller", "-workflowmgr", "10.98.208.2:9095"]
