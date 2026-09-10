#!/usr/bin/env python3
"""Bridge for external agents. Persists original outputs; never calls a model."""
import argparse
import json
import os
from pathlib import Path
import re
from urllib.error import HTTPError
from urllib.parse import urlsplit
from urllib.request import HTTPRedirectHandler, Request, build_opener


class NoRedirects(HTTPRedirectHandler):
    def redirect_request(self, req, fp, code, msg, headers, newurl):
        return None  # Never forward the manager credential to another URL.


def request(path, body=None, key=None):
    base = os.environ.get("BANDIA_API_URL", "http://127.0.0.1:8080").rstrip("/")
    url = urlsplit(base)
    local = url.hostname in ("127.0.0.1", "localhost", "::1")
    if url.scheme not in ("http", "https") or (url.scheme != "https" and not local) or url.username or url.password or url.query or url.fragment or url.path:
        raise RuntimeError("Use an HTTPS API origin, or HTTP on localhost, without credentials/path/query")
    token = os.environ.get("BANDIA_MANAGER_TOKEN")
    if not token:
        raise RuntimeError("Set BANDIA_MANAGER_TOKEN; do not pass the token as a command argument")
    headers = {"Authorization": "Bearer " + token}
    if key:
        headers["Idempotency-Key"] = key
    data = None
    if body is not None:
        data = json.dumps(body, ensure_ascii=False).encode()
        headers["Content-Type"] = "application/json"
    try:
        with build_opener(NoRedirects).open(Request(base + path, data=data, headers=headers), timeout=15) as response:
            return json.load(response)
    except HTTPError as error:
        raise RuntimeError(f"API returned HTTP {error.code}. For ambiguous writes retry the SAME key, context and input; for 409 inspect current context.") from None


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    commands = parser.add_subparsers(dest="command", required=True)
    create = commands.add_parser("create")
    create.add_argument("--topic", required=True)
    create.add_argument("--key", required=True)
    create.add_argument("--instruction-id", action="append", default=[])
    for name in ("context", "turn", "close"):
        command = commands.add_parser(name)
        command.add_argument("meeting_id")
        if name == "context":
            command.add_argument("--out", required=True, help="New file; never overwritten")
        else:
            command.add_argument("--context", required=True)
            command.add_argument("--input", required=True)
            command.add_argument("--key", required=True)
    args = parser.parse_args()
    if args.command == "create":
        result = request("/v1/meetings", {"mode": "external", "topic": args.topic, "instruction_ids": args.instruction_id}, args.key)
    else:
        if not re.fullmatch("[a-f0-9]{32}", args.meeting_id):
            parser.error("Invalid meeting ID")
        path = "/v1/meetings/" + args.meeting_id
        if args.command == "context":
            result = request(path + "/context")
            # These files may contain private history. Keep them out of public Git.
            fd = os.open(args.out, os.O_WRONLY | os.O_CREAT | os.O_EXCL, 0o600)
            with os.fdopen(fd, "w") as file:
                json.dump(result, file, ensure_ascii=False, indent=2)
            print("Context saved; next character:", result["next_character_id"] or "none")
            return
        context = json.loads(Path(args.context).read_text())
        data = json.loads(Path(args.input).read_text())
        if context["meeting"]["id"] != args.meeting_id:
            parser.error("Context belongs to another meeting")
        if args.command == "turn":
            if set(data) != {"prompt", "content", "agent_run_id"}:
                parser.error("Turn input must contain exactly prompt, content, agent_run_id")
            body = dict(data, sequence=context["next_sequence"], history_through=context["next_sequence"] - 1,
                        character_id=context["next_character_id"], context_hash=context["context_hash"])
            result = request(path + "/turns", body, args.key)
        else:
            if set(data) != {"summary", "decisions"}:
                parser.error("Closure input must contain exactly summary, decisions")
            result = request(path + "/close", dict(data, last_sequence=len(context["messages"])), args.key)
    # Output includes original data, never credentials. Redirect to a private artifact if needed.
    print(json.dumps(result, ensure_ascii=False, indent=2))


if __name__ == "__main__":
    main()
