# NVIDIA CUDA-QX QEC notebook
# For quantum error correction research: decoders, surface codes, LDPC.
FROM nvcr.io/nvidia/cuda-quantum:latest

RUN python -c "import cudaqx; print('CUDA-QX:', cudaqx.__version__)" || \
    pip install --no-cache-dir cudaqx

RUN pip install --no-cache-dir \
    stim \
    pymatching

RUN jupyter notebook --generate-config 2>/dev/null || true
RUN echo "c.Server.token = ''" >> /etc/jupyter/jupyter_server_config.py 2>/dev/null || true
RUN echo "c.Server.ip = '0.0.0.0'" >> /etc/jupyter/jupyter_server_config.py 2>/dev/null || true

ENV NVIDIA_VISIBLE_DEVICES=all

EXPOSE 8888
CMD ["jupyter", "notebook", "--no-browser", "--ip=0.0.0.0"]
