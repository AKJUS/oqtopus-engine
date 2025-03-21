ARG PYTHON_VERSION=3.11.9
FROM python:$PYTHON_VERSION-slim
COPY --from=ghcr.io/astral-sh/uv:0.6.3 /uv /uvx /bin/

# install tools
RUN apt-get update && apt-get install -y --no-install-recommends \
  vim less gcc g++ cargo iptables \
  && apt-get clean && rm -rf /var/lib/apt/lists/*

COPY init.sh /root/
RUN chmod 100 /root/init.sh

WORKDIR /app

# setup user
RUN adduser --disabled-password --gecos "" appuser && chown -R appuser:appuser /app
USER appuser

COPY pyproject.toml /app
COPY src/ /app/src/

# install python dependencies
RUN uv sync

# set environment variables
ENV PYTHONPATH=/app/src/sse_runtime/
ENV OQTOPUS_ENV=sse_container

# isolate user's program to tmpfs
WORKDIR /sse
