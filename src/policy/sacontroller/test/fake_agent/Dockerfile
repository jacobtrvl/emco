ARG MAINDOCKERREPO
ARG SERVICE_BASE_IMAGE_NAME
ARG SERVICE_BASE_IMAGE_VERSION

#FROM ${MAINDOCKERREPO}${SERVICE_BASE_IMAGE_NAME}:${SERVICE_BASE_IMAGE_VERSION}
FROM jacobaarna/alpine:3.12

WORKDIR /opt/emco/policy

RUN addgroup -S emco && adduser -S -G emco emco
RUN chown emco:emco . -R

COPY --chown=emco ./agent ./

USER emco

ENTRYPOINT ["./agent"]

