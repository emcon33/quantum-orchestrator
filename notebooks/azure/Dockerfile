FROM ghcr.io/quantum-orchestrator/notebook-base:latest

USER root
# Install QDK + Azure Quantum SDK
RUN pip install --no-cache-dir \
    azure-quantum>=1.0 \
    qsharp>=0.24 \
    microsoft-quantum>=1.2

USER ${NOTEBOOK_USER}
WORKDIR /home/${NOTEBOOK_USER}/work
