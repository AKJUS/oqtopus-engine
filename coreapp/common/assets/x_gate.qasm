OPENQASM 3;

qubit[1] q;
bit[1] c;

x q[0];

c[0] = measure q[0];