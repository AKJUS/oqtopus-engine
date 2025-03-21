SHELL := bash
.SHELLFLAGS := -eu -o pipefail -c
.DEFAULT_GOAL := help

.PHONY: lint
lint:
	buf lint

.PHONY: generate-job
generate-job:
	buf generate --template buf.gen.job.yaml --path proto/job_interface

.PHONY: generate-qpu
generate-qpu:
	buf generate --template buf.gen.qpu.yaml --path proto/qpu_interface

.PHONY: generate-sse
generate-sse:
	buf generate --template buf.gen.sse.yaml --path proto/sse_interface

.PHONY: generate-multiprog
generate-multiprog:
	buf generate --template buf.gen.multiprog.yaml --path proto/multiprog_interface

.PHONY: generate-mitig
generate-mitig:
	buf generate --template buf.gen.mitig.yaml --path proto/mitigation_interface

.PHONY: generate-est
generate-est:
	buf generate --template buf.gen.est.yaml --path proto/estimation_interface

.PHONY: generate-all
generate-all: lint generate-job generate-qpu generate-sse generate-multiprog generate-mitig generate-est
