FROM python:3.11-slim

ENV PYTHONFAULTHANDLER=1 \
    PYTHONUNBUFFERED=1 \
    PYTHONHASHSEED=random \
    PIP_NO_CACHE_DIR=off \
    PIP_DISABLE_PIP_VERSION_CHECK=on \
    PIP_DEFAULT_TIMEOUT=100 \
    POETRY_VERSION=1.0.0

RUN rm -rf /var/lib/apt/lists/*
RUN apt-get update
RUN apt-get install -y gcc python3-dev curl graphviz graphviz-dev
RUN rm -rf /var/lib/apt/lists/*
RUN pip install "poetry==1.3.2"

COPY ./ds/pyproject.toml ./

RUN poetry config virtualenvs.create false \
    && poetry install --only main --no-interaction --no-ansi

COPY ./ds /backend

WORKDIR /backend

CMD ["python3", "main.py"]
