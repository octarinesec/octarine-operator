FROM registry.access.redhat.com/ubi8/ubi-minimal:latest

ENV OPERATOR=/usr/local/bin/octarine-operator \
    USER_UID=1001 \
    USER_NAME=octarine-operator \
    HOME=/opt/helm
COPY watches.yaml ${HOME}/watches.yaml

# install operator binary
COPY build/_output/bin/octarine-operator ${OPERATOR}

COPY build/bin /usr/local/bin
RUN  /usr/local/bin/user_setup
COPY helm-charts/octarine ${HOME}/helm-charts/octarine

ENTRYPOINT ["/usr/local/bin/entrypoint"]

USER ${USER_UID}