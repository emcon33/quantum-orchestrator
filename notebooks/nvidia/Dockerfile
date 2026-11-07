# NVIDIA CUDA-Q + cuQuantum notebook
# Requires NVIDIA GPU. Based on NGC CUDA-Q container.
FROM nvcr.io/nvidia/cuda-quantum:latest

RUN pip install --no-cache-dir \
    qiskit>=1.0 \
    pennylane>=0.36 \
    qibo>=0.2 \
    cirq>=1.2

RUN python -c "import cudaqx; print(cudaqx.__version__)" || \
    pip install --no-cache-dir cudaqx

RUN jupyter notebook --generate-config 2>/dev/null || true
RUN echo "c.Server.token = ''" >> /etc/jupyter/jupyter_server_config.py 2>/dev/null || true
RUN echo "c.Server.ip = '0.0.0.0'" >> /etc/jupyter/jupyter_server_config.py 2>/dev/null || true

ENV NVIDIA_VISIBLE_DEVICES=all
ENV NVIDIA_DRIVER_CAPABILITIES=compute,utility

EXPOSE 8888
CMD ["jupyter", "notebook", "--no-browser", "--ip=0.0.0.0"]
