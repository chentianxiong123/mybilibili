FROM pgvector/pgvector:pg16

RUN apt-get update && apt-get install -y --no-install-recommends \
    curl ca-certificates \
    && mkdir -p /etc/apt/keyrings \
    && curl -fsSL https://repo.pigsty.io/key | gpg --dearmor -o /etc/apt/keyrings/pigsty.gpg \
    && echo "deb [signed-by=/etc/apt/keyrings/pigsty.gpg] https://repo.pigsty.io/apt/pgsql/bookworm bookworm main" > /etc/apt/sources.list.d/pigsty.list \
    && apt-get update \
    && apt-get install -y --no-install-recommends postgresql-16-zhparser \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*
