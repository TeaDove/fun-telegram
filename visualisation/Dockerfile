FROM python:3.11-slim

ENV PYTHONFAULTHANDLER=1 \
    PYTHONUNBUFFERED=1 \
    PYTHONHASHSEED=random \
    PIP_NO_CACHE_DIR=off \
    PIP_DISABLE_PIP_VERSION_CHECK=on \
    PIP_DEFAULT_TIMEOUT=100 \
    POETRY_VERSION=1.0.0

RUN pip install "poetry==1.3.2"

COPY . /backend

WORKDIR /backend

RUN poetry input virtualenvs.create false \
    && poetry install --only main --no-interaction --no-ansi

CMD ["python3", "draw.py"]
