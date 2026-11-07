# IBM Qiskit + Nighthawk/Heron notebook
# Based on the official IBM Quantum notebook image.
FROM ibmquantum/qiskit:latest

# Qiskit v2.5 with C API, PBC, Clifford+T
RUN pip install --no-cache-dir \
    qiskit>=2.5 \
    qiskit-ibm-runtime>=0.34 \
    qiskit-algorithms>=0.3

# Additional frameworks for cross-vendor work
RUN pip install --no-cache-dir \
    pennylane>=0.36 \
    cirq>=1.2 \
    qibo>=0.2

# QEC tools
RUN pip install --no-cache-dir \
    stim \
    pymatching \
    qec-circuits

# Pre-configure Jupyter
RUN jupyter notebook --generate-config 2>/dev/null || true
RUN echo "c.Server.token = ''" >> /etc/jupyter/jupyter_server_config.py 2>/dev/null || true
RUN echo "c.Server.ip = '0.0.0.0'" >> /etc/jupyter/jupyter_server_config.py 2>/dev/null || true

EXPOSE 8888
CMD ["jupyter", "notebook", "--no-browser", "--ip=0.0.0.0"]
