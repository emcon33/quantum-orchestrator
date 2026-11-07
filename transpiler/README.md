# Quantum Circuit Transpiler Sidecar

A lightweight HTTP service that converts quantum circuits between formats.
Runs as a sidecar container next to the operator.

## Supported Conversions

| From \ To | QASM 3 | QIR | Qiskit | Cirq | PennyLane | Quil |
|-----------|--------|-----|--------|------|-----------|------|
| QASM 3    | ✅     | ✅  | ✅     | ✅   | ✅        | ✅   |
| QIR       | ✅     | —   | ✅     | —    | —         | —    |
| Qiskit    | ✅     | ✅  | —      | ✅   | ✅        | ✅   |

✅ = available (requires the corresponding `pip` extra)

## Run Locally

```bash
pip install -r requirements.txt
python server.py
# → http://localhost:8080
```

## API

### POST /convert

```json
{
  "input": { "format": "qasm3", "source": "OPENQASM 3.0; ..." },
  "target_format": "qir",
  "device_id": null
}
```

### GET /health

Returns status + list of available conversions.

### GET /capabilities

Returns the full conversion matrix.
