#!/usr/bin/env python3
"""Native model integration oracle; fixture success is not live-provider acceptance."""
from __future__ import annotations
import argparse
import hashlib
import json
import os
from pathlib import Path
import platform
import re
import signal
import subprocess
import sys
import time

TEST_FILES = [
    "nanovms_deletion_safety_test.go",
    "nanovms_infrastructure_adapter_test.go",
    "disposable_lifecycle_test.go",
    "nanovms_http_lifecycle_test.go",
    "nanovms_http_transport_test.go",
]

def digest(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()

def classify(code, events, expected):
    terminal = [e for e in events if e.get("Test") in expected and e.get("Action") in ("pass", "fail", "skip")]
    failures = [e for e in events if e.get("Action") in ("fail", "skip")]
    package_pass = [e for e in events if e.get("Action") == "pass" and not e.get("Test")]
    return (code == 0 and bool(expected) and not failures and len(package_pass) == 1
            and len(terminal) == len(expected)
            and {e["Test"] for e in terminal} == set(expected)
            and all(e["Action"] == "pass" for e in terminal))

def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--candidate", required=True)
    parser.add_argument("--output", type=Path, required=True)
    args = parser.parse_args()
    output = args.output.resolve()
    output.mkdir(parents=True, exist_ok=True)
    receipt = {"schema":"byteport-deletion-safety/v1", "product":"KooshaPari/BytePort",
               "candidate":args.candidate, "result":"COLLECTOR_FAILURE", "product_accepted":False,
               "live_provider_accepted":False,"scope":"native models with fixture transports",
               "run":os.getenv("GITHUB_RUN_ID"),"run_attempt":os.getenv("GITHUB_RUN_ATTEMPT"),
               "timestamp":time.strftime("%Y-%m-%dT%H:%M:%SZ",time.gmtime()),
               "environment":{"platform":platform.platform(),"python":sys.version},
               "verifier_sha256":digest(Path(__file__))}
    rc = 1
    try:
        actual = subprocess.check_output(["git","rev-parse","HEAD"],text=True).strip()
        receipt["tested_checkout"] = actual
        if actual != args.candidate or not re.fullmatch(r"[0-9a-f]{40}",actual):
            receipt["result"] = "WRONG_CANDIDATE"
            raise ValueError("candidate differs from tested checkout")
        if subprocess.check_output(["git","diff","HEAD","--name-only"],text=True).strip():
            raise ValueError("tracked source is dirty")
        root = Path("backend/byteport")
        expected = []
        for name in TEST_FILES:
            expected.extend(re.findall(r"(?m)^func (Test\w+)\s*\(",(root/"models"/name).read_text()))
        if not expected or len(expected) != len(set(expected)):
            raise ValueError("missing or duplicated named test inventory")
        receipt["expected_tests"] = expected
        paths = [root/"models"/name for name in TEST_FILES]
        paths += [root/"models"/name for name in ["nanovms_infrastructure_adapter.go","nanovms_http_transport.go","infrastructure_reconcile_execution.go","infrastructure_target_adapter.go","infrastructure_graph.go"]]
        paths += [root/"go.mod", root/"go.sum"]
        receipt["sources"] = {str(p):digest(p) for p in paths}
        receipt["environment"]["go"] = subprocess.check_output(["go","version"],text=True).strip()
        pattern = "^("+"|".join(re.escape(n) for n in expected)+")$"
        command = ["go","test","./models","-race","-count=1","-timeout=180s","-run",pattern,"-json"]
        receipt["command"] = command
        receipt["working_directory"] = str(root)
        with (output/"go-test.jsonl").open("w") as log:
            process = subprocess.Popen(command,cwd=root,stdout=log,stderr=subprocess.STDOUT,start_new_session=True)
            try:
                code = process.wait(timeout=900)
            except subprocess.TimeoutExpired:
                os.killpg(process.pid,signal.SIGKILL)
                process.wait()
                raise ValueError("native build/test timeout")
        receipt["exit_code"] = code
        raw = output/"go-test.jsonl"
        if raw.stat().st_size > 32*1024*1024:
            raise ValueError("raw output exceeded parser limit")
        events = []
        for line in raw.read_text(errors="replace").splitlines():
            try:
                event = json.loads(line)
                if isinstance(event,dict): events.append(event)
            except json.JSONDecodeError:
                continue
        receipt["tests"] = [{k:e[k] for k in ("Test","Action","Elapsed") if k in e}
                             for e in events if e.get("Test") and e.get("Action") in ("pass","fail","skip")]
        if classify(code,events,expected):
            receipt["result"] = "PASS"
            rc = 0
        else:
            receipt["result"] = "FAIL" if any(e.get("Test") and e.get("Action")=="fail" for e in events) else "INVALID_ORACLE_OR_COLLECTOR_FAILURE"
    except (OSError,ValueError,subprocess.SubprocessError) as exc:
        receipt["error"] = str(exc)
    finally:
        receipt["raw_artifacts"] = {p.name:digest(p) for p in output.glob("*.jsonl")}
        (output/"receipt.json").write_text(json.dumps(receipt,indent=2)+"\n")
        print(json.dumps(receipt,indent=2))
    return rc

if __name__ == "__main__":
    raise SystemExit(main())
