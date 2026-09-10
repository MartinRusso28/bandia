#!/usr/bin/env python3
"""Storage/HTTP fixture test, NOT an AI conversation. No model calls."""
import subprocess
import uuid
from smoke import ROOT, call, load_token, ready


def main():
    ready()
    token = load_token()
    key = "external-fixture-" + uuid.uuid4().hex
    accepted = call("/v1/meetings", token, {"mode": "external", "topic": "CI test fixtures, not agent output"}, key, 202)
    assert accepted["status"] == "waiting_external"
    path = "/v1/meetings/" + accepted["meeting_id"]
    first = None
    original = None
    for sequence in range(1, 12):
        context = call(path + "/context", token)
        assert context["next_sequence"] == sequence
        data = {"sequence": sequence, "history_through": sequence - 1,
                "character_id": context["next_character_id"], "context_hash": context["context_hash"],
                "agent_run_id": key + "-fixture-" + str(sequence),
                "prompt": "  TEST FIXTURE prompt\n", "content": "TEST FIXTURE response\n"}
        result = call(path + "/turns", token, data, key + str(sequence), 201)
        assert result["prompt"] == data["prompt"] and result["content"] == data["content"]
        assert call(path + "/turns", token, data, key + str(sequence), 201) == result
        if sequence == 1:
            first, original = result, data
        if sequence == 3:
            subprocess.run(["docker", "compose", "restart", "api", "postgres"], cwd=ROOT, check=True)
            ready()
    call(path + "/turns", token, original, key + "stale", 409)
    body = {"last_sequence": 11, "summary": "TEST FIXTURE summary", "decisions": ["TEST FIXTURE decision"]}
    closed = call(path + "/close", token, body, key + "close")
    assert closed["status"] == "completed"
    assert call(path + "/close", token, body, key + "close") == closed
    assert call(path + "/turns", token, original, key + "1", 201) == first
    assert call("/v1/jobs/" + accepted["job_id"], token)["status"] == "completed"
    assert len(call(path + "/messages", token)["data"]) == 11
    print("PASS: 11 fixture turns, context linkage, replay, mid-meeting restart and durable closure; no AI executed")


if __name__ == "__main__":
    main()
