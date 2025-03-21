import matplotlib
import numpy
import yaml
import networkx
import pandas
import tqdm
import scipy
import sklearn
import quri_parts
import quri_parts.riqu
import qiskit
import qulacs
import skqulacs
import pyqubo
import openjij
import cirq
import pennylane
import openfermion

import time

from quri_parts.riqu.backend.sampling import RiquSamplingBackend, RiquConfig
from quri_parts.circuit import QuantumCircuit


for i in range(3):
  print(f"## Start iteration {i} ##")
  circuit = QuantumCircuit(2)
  circuit.add_H_gate(0)
  circuit.add_X_gate(0)
  circuit.add_CNOT_gate(0, 1)
  circuit.add_RY_gate(1, 0.1*i)
  job = RiquSamplingBackend().sample(circuit, 10*i+100, transpiler="none")
  print("#### Job1:")
  print(job)
  result = job.result()
  print("#### Result:")
  print(result.counts)

  job2 = RiquSamplingBackend(RiquConfig("dummy", "dummy")).sample(circuit, 10*i+100, transpiler="none")
  print("#### Job2:")
  print(job2)
  result2 = job2.result()
  print("#### Result:")
  print(result2.counts)
print("## Finish ##")
