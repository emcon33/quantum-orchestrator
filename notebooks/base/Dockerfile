FROM jupyter/scipy-notebook:latest

USER root
# Install quantum SDKs
RUN pip install --no-cache-dir \
    qiskit>=1.0 \
    cirq>=1.2 \
    pennylane>=0.36 \
    pyquil>=3.25 \
    qbraid>=0.11 \
    openqaoa>=0.6 \
    qibo>=0.2

USER ${NOTEBOOK_USER}
WORKDIR /home/${NOTEBOOK_USER}/work

# Pre-configure Jupyter
RUN jupyter notebook --generate-config && \
    echo "c.Server.token = ''" >> ~/.jupyter/jupyter_server_config.py && \
    echo "c.Server.ip = '0.0.0.0'" >> ~/.jupyter/jupyter_server_config.py

EXPOSE 8888
CMD ["start-notebook.sh", "jupyter", "notebook", "--no-browser"]
