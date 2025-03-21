from quri_parts.riqu.backend.sampling import RiquSamplingBackend
from quri_parts.circuit import QuantumCircuit

import subprocess

# Check the volume size attached to SSE container.
def check_volume(volume_size):
    print("Checking the volume size attached to SSE container.")
    cmd = ['df', '-h', '/sse']
    res = subprocess.run(cmd, capture_output=True, text=True)
    print(f"df command returns code {res.returncode}")
    print(res.stdout)
    print(res.stderr)
    if res.returncode != 0:
        return res.returncode
    if volume_size in res.stdout:
        return 0
    else:
        return 1

# Check that outbound network from SSE container is restricted.
def check_network():
    print("Checking that outbound network from SSE container is restricted.")
    cmd = ['pip', 'search', 'qiskit', '--timeout', '3', '--retries', '1']
    res = subprocess.run(cmd, capture_output=True, text=True) # return code should be 2 if the connection failed
    print(f"pip command returns code {res.returncode}")
    print(res.stdout)
    print(res.stderr)

    return 0 if res.returncode==2 else 1

ret = check_volume("300M") | check_network()
print(f"Finish, the return code is {ret}")

# Execute sampling to output result.json
circuit = QuantumCircuit(2)
circuit.add_H_gate(0)
job = RiquSamplingBackend().sample(circuit, 1000, transpiler="none")

exit(ret)
