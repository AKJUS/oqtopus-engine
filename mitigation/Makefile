SHELL := bash
.SHELLFLAGS := -eu -o pipefail -c
.DEFAULT_GOAL := help
.PHONY: help 

build : 
	docker compose --env-file .env build

up : 
	@if [ ! -d $(SHARE_DIR)/$(LOG_DIR) ]; then \
		echo ";; mkdir -p $(SHARE_DIR)/$(LOG_DIR)"; mkdir -p $(SHARE_DIR)/$(LOG_DIR); \
	fi
	docker compose --env-file .env up -d mitigator

down : 
	docker compose --env-file .env down mitigator

ruff :
	@echo "...checking lint check"
	@ruff check src
	@echo "...checking format check"
	@ruff format src --diff --check

clean :
	@rm .venv -rf
	@rm .ruff_cache -rf
	@rm .pytest_cache -rf

help :
