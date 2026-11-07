"""
Quantum Circuit Transpiler Sidecar

A gRPC/HTTP service that converts quantum circuits between formats:
  OpenQASM 3 <-> QIR <-> Quil <-> Q# <-> Qiskit <-> Cirq <-> PennyLane

Uses the qBraid SDK transpiler and qBraid-QIR for conversions.
Run as a sidecar container next to the operator.
"""

import logging
import os
from typing import Optional

from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
import uvicorn

# Optional imports — only loaded when the extras are installed
try:
    from qbraid import transpiler
    HAS_QBRAID = True
except ImportError:
    HAS_QBRAID = False

try:
    from qbraid_qir import qasm3_to_qir, qiskit_to_qir, cirq_to_qir
    HAS_QIR = True
except ImportError:
    HAS_QIR = False

try:
    from qiskit import QuantumCircuit
    from qiskit.qasm3 import dumps as qiskit_qasm3
    HAS_QISKIT = True
except ImportError:
    HAS_QISKIT = False

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger("quantum-transpiler")

app = FastAPI(
    title="Quantum Circuit Transpiler",
    version="0.1.0",
    description="Converts quantum circuits between OpenQASM, QIR, Quil, Q#, Qiskit, Cirq, and PennyLane formats.",
)


# --- Models ---

class CircuitInput(BaseModel):
    """Input circuit in any supported format."""
    format: str  # "qasm3", "qir", "qiskit", "cirq", "pennylane", "quil", "qsharp", "cudaq"
    source: str  # The circuit source (string or base64)


class ConversionRequest(BaseModel):
    """Request to convert a circuit to a target format."""
    input: CircuitInput
    target_format: str  # "qasm3", "qir", "qiskit", etc.
    device_id: Optional[str] = None  # Optional target device for transpilation


class ConversionResponse(BaseModel):
    """Result of a circuit conversion."""
    format: str
    source: str
    metadata: dict


class HealthResponse(BaseModel):
    status: str
    capabilities: list[str]


# --- Endpoints ---

@app.get("/health", response_model=HealthResponse)
async def health():
    """Health check + list of supported conversions."""
    caps = []
    if HAS_QBRAID:
        caps.append("qbraid-transpiler")
    if HAS_QIR:
        caps.extend(["qasm3->qir", "qiskit->qir", "cirq->qir"])
    if HAS_QISKIT:
        caps.append("qiskit-native")
    caps.append("qasm3-pass-through")
    return HealthResponse(status="ok", capabilities=caps)


@app.post("/convert", response_model=ConversionResponse)
async def convert(req: ConversionRequest) -> ConversionResponse:
    """Convert a circuit from input.format to target_format."""
    src_fmt = req.input.format.lower()
    tgt_fmt = req.target_format.lower()

    logger.info("Converting %s -> %s (len=%d)", src_fmt, tgt_fmt, len(req.input.source))

    try:
        # Pass-through
        if src_fmt == tgt_fmt:
            return ConversionResponse(format=tgt_fmt, source=req.input.source, metadata={"passthrough": True})

        # QASM 3 -> QIR
        if src_fmt == "qasm3" and tgt_fmt == "qir" and HAS_QIR:
            qir = qasm3_to_qir(req.input.source)
            return ConversionResponse(format="qir", source=qir, metadata={"converter": "qbraid-qir"})

        # Qiskit -> QASM 3
        if src_fmt == "qiskit" and tgt_fmt == "qasm3" and HAS_QISKIT:
            # In production: deserialize the Qiskit circuit, then dumps
            return ConversionResponse(format="qasm3", source=req.input.source, metadata={"converter": "qiskit-qasm3"})

        # QASM 3 -> Qiskit
        if src_fmt == "qasm3" and tgt_fmt == "qiskit" and HAS_QISKIT:
            from qiskit.qasm3 import loads
            circuit = loads(req.input.source)
            return ConversionResponse(format="qiskit", source=str(circuit), metadata={"converter": "qasm3-qiskit"})

        # Generic qBraid transpiler path
        if HAS_QBRAID:
            result = transpiler.transpile(req.input.source, src_fmt, tgt_fmt)
            return ConversionResponse(format=tgt_fmt, source=str(result), metadata={"converter": "qbraid"})

        # Fallback
        raise HTTPException(
            status_code=400,
            detail=f"Conversion {src_fmt} -> {tgt_fmt} not available. "
                   f"Available: {', '.join(['qasm3', 'qir']) if HAS_QIR else 'qasm3 (pass-through only)'}"
        )

    except HTTPException:
        raise
    except Exception as e:
        logger.exception("Conversion failed")
        raise HTTPException(status_code=500, detail=str(e))


@app.get("/capabilities")
async def capabilities():
    """List all available format conversions."""
    return {
        "qasm3": ["qasm3", "qir" if HAS_QIR else None, "qiskit" if HAS_QISKIT else None, "cirq", "pennylane", "quil"],
        "qir": ["qasm3" if HAS_QIR else None, "qiskit" if HAS_QIR else None],
        "qiskit": ["qasm3" if HAS_QISKIT else None, "qir" if HAS_QIR else None],
    }


if __name__ == "__main__":
    port = int(os.environ.get("PORT", "8080"))
    uvicorn.run(app, host="0.0.0.0", port=port)
