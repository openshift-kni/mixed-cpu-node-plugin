FROM docker.io/kindest/node:v1.26.0

COPY 99-nri.conf /etc/crio/crio.conf.d/

RUN echo "Installing helper packages..." \
    && apt update \
    && apt install -y less vim
RUN echo "Installing crio ..." \
    # https://github.com/cri-o/cri-o#installing-cri-o
    && curl https://raw.githubusercontent.com/cri-o/cri-o/main/scripts/get | /bin/bash \
    && rm -rf /etc/crictl.yaml \
    #  delete crun configuration it doesn't work in kind setup and needs to be investigated
    && rm -f /etc/crio/crio.conf.d/10-crun.conf \

# do not disable containerd becasue it's needed for kind load docker-image command
RUN systemctl enable crio
