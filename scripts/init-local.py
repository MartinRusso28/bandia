#!/usr/bin/env python3
"""Create local-only secrets once. Never overwrite credentials or print them."""
import os
from pathlib import Path
import secrets

root = Path(__file__).resolve().parents[1]
target = root / ".env"
password = secrets.token_hex(24)
token = secrets.token_hex(32)
content = (
    "# Generated for local development. Do not commit.\n"
    f"POSTGRES_PASSWORD={password}\n"
    f"BANDIA_MANAGER_TOKEN={token}\n"
    f"BANDIA_DATABASE_URL=postgres://bandia:{password}@127.0.0.1:5432/bandia?sslmode=disable\n"
    "BANDIA_HTTP_ADDR=127.0.0.1:8080\n"
)
try:
    fd = os.open(target, os.O_WRONLY | os.O_CREAT | os.O_EXCL, 0o600)
except FileExistsError:
    print("Local configuration already exists; preserved without changes.")
else:
    with os.fdopen(fd, "w") as file:
        file.write(content)
    print("Local configuration created. Run: docker compose up --build -d")
