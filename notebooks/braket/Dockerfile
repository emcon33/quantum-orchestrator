FROM ghcr.io/quantum-orchestrator/notebook-base:latest

USER root
RUN pip install --no-cache-dir \
    amazon-braket-sdk>=1.66 \
    amazon-braket-defaults>=1.26

USER ${NOTEBOOK_USER}
WORKDIR /home/${NOTEBOOK_USER}/work
