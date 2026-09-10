#!/usr/bin/env python3
"""Exercise the local API. --restart also restarts this project's Compose services."""
import argparse
import json
import os
from pathlib import Path
import subprocess
import time
from urllib.error import HTTPError, URLError
from urllib.request import Request, urlopen
import uuid

ROOT = Path(__file__).resolve().parents[1]
BASE = "http://127.0.0.1:8080"

def load_token():
    token = os.environ.get("BANDIA_MANAGER_TOKEN")
    if not token and (ROOT / ".env").exists():
        for line in (ROOT / ".env").read_text().splitlines():
            if line.startswith("BANDIA_MANAGER_TOKEN="):
                token = line.split("=", 1)[1]
    if not token:
        raise RuntimeError("Run scripts/init-local.py first or set BANDIA_MANAGER_TOKEN")
    return token

def call(path, token=None, body=None, key=None, expected=200):
    headers = {}
    if token:
        headers["Authorization"] = "Bearer " + token
    if key:
        headers["Idempotency-Key"] = key
    data = None
    if body is not None:
        data = json.dumps(body).encode()
        headers["Content-Type"] = "application/json"
    request = Request(BASE + path, data=data, headers=headers)
    try:
        response = urlopen(request, timeout=6)
    except HTTPError as error:
        response = error
    with response:
        result = json.load(response)
        if response.status != expected:
            code = result.get("error", {}).get("code", "unknown")
            raise RuntimeError(f"{path}: HTTP {response.status}, expected {expected} ({code})")
        return result

def ready():
    deadline = time.monotonic() + 45
    while time.monotonic() < deadline:
        try:
            call("/readyz")
            return
        except (RuntimeError, URLError, TimeoutError):
            time.sleep(0.5)
    raise RuntimeError("API did not become ready in 45 seconds")

def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--restart", action="store_true")
    args = parser.parse_args()
    token = load_token()
    ready()
    call("/v1/system", expected=401)
    capabilities = call("/v1/system", token)["capabilities"]
    assert capabilities["persistent_meetings"] and not capabilities["agents"]
    prefix = "smoke-" + uuid.uuid4().hex
    instruction = call("/v1/manager-instructions", token,
                       {"text": "Prueba local: definir próximos proyectos"}, prefix + "-i", 201)
    body = {"topic": "Smoke test: rumbo inicial", "instruction_ids": [instruction["id"]]}
    accepted = call("/v1/meetings", token, body, prefix + "-m", 202)
    assert call("/v1/meetings", token, body, prefix + "-m", 202) == accepted
    call("/v1/meetings", token, {"topic": "Otro contenido"}, prefix + "-m", 409)
    meeting_path = "/v1/meetings/" + accepted["meeting_id"]
    job_path = "/v1/jobs/" + accepted["job_id"]
    deadline = time.monotonic() + 15
    while time.monotonic() < deadline:
        job = call(job_path, token)
        if job["status"] == "blocked":
            break
        time.sleep(0.25)
    assert job["status"] == "blocked", "Worker did not process the job"
    assert job["attempts"] == 1
    assert job["blocked_reason"] == "agent_provider_not_implemented"
    meeting = call(meeting_path, token)
    assert meeting["status"] == "blocked" and len(meeting["participants"]) == 5
    assert meeting["instructions"][0]["id"] == instruction["id"]
    assert call(meeting_path + "/messages", token) == {"data": []}
    if args.restart:
        subprocess.run(["docker", "compose", "restart", "api", "postgres"], cwd=ROOT, check=True)
        ready()
        assert call(meeting_path, token) == meeting
        assert call(job_path, token) == job
        assert call("/v1/meetings", token, body, prefix + "-m", 202) == accepted
    print("PASS: persistence, authentication, idempotency, snapshots and honest capability blocking")
    if args.restart:
        print("PASS: same data and job after API/PostgreSQL restart")
    print("meeting_id=" + accepted["meeting_id"])
    print("job_id=" + accepted["job_id"])

if __name__ == "__main__":
    main()
