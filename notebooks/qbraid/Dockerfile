FROM ghcr.io/quantum-orchestrator/notebook-base:latest

USER root
RUN pip install --no-cache-dir \
    qbraid>=0.11 \
    qbraid-qir>=0.3

USER ${NOTEBOOK_USER}
WORKDIR /home/${NOTEBOOK_USER}/work
