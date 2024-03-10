FROM mcr.microsoft.com/devcontainers/go:1-1.22-bookworm

ARG BUILD_USER="vscode"
ENV BUILD_USER="${BUILD_USER}"

ARG BUILD_HOME="/home/${BUILD_USER}"
ENV BUILD_HOME="${BUILD_HOME}"

# [Optional] Uncomment this section to install additional OS packages.
RUN apt-get update && export DEBIAN_FRONTEND=noninteractive \
    && apt-get -y install --no-install-recommends direnv postgresql-client tree

# Add some useful additions to bashrc
COPY "bash_additions.sh" "/tmp/"
RUN set -x \
    && for f in .bash_profile .bashrc .profile; do printf "\n" >> "${BUILD_HOME}/${f}"; done \
    && for f in .bash_profile .bashrc .profile; do cat "/tmp/bash_additions.sh" >> "${BUILD_HOME}/${f}"; done \
    && rm "/tmp/bash_additions.sh"

# [Optional] Uncomment the next lines to use go get to install anything else you need
# USER vscode
# RUN go get -x <your-dependency-or-tool>
# USER root

# [Optional] Uncomment this line to install global node packages.
# RUN su vscode -c "source /usr/local/share/nvm/nvm.sh && npm install -g <your-package-here>" 2>&1
