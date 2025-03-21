ARG PYTHON_VERSION=3.12
FROM python:$PYTHON_VERSION-slim
COPY --from=ghcr.io/astral-sh/uv:0.6.3 /uv /uvx /bin/

WORKDIR /app

# install tools
RUN apt-get update && apt-get install -y --no-install-recommends \
  git \
  && apt-get clean && rm -rf /var/lib/apt/lists/*

# setup user
RUN adduser --disabled-password --gecos "" appuser && chown -R appuser:appuser /app
USER appuser

COPY . /app

# install python dependencies
RUN uv sync 

# run application
CMD ["uv", "run", "python", "src/circuit_combiner/circuit_combiner.py"]
