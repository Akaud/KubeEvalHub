#!/usr/bin/env python3
import json
import shutil
import subprocess
import sys
import tempfile

try:
    import yaml
except ImportError:
    print("Missing dependency: pyyaml", file=sys.stderr)
    print("Install with: pip install pyyaml", file=sys.stderr)
    sys.exit(1)

NAMESPACE = "load-lab"


def run(cmd, check=True, capture_output=False):
    print("+", " ".join(cmd))
    return subprocess.run(
        cmd,
        check=check,
        text=True,
        capture_output=capture_output,
    )


def require_binary(name):
    if shutil.which(name) is None:
        print(f"Missing required binary: {name}", file=sys.stderr)
        sys.exit(1)


def minikube_running():
    try:
        result = run(
            ["minikube", "status", "-o", "json"],
            check=False,
            capture_output=True,
        )
        if result.returncode != 0:
            return False
        data = json.loads(result.stdout)
        return (
            data.get("Host") == "Running"
            and data.get("Kubelet") == "Running"
            and data.get("APIServer") == "Running"
        )
    except Exception:
        return False


def start_minikube():
    if minikube_running():
        print("Minikube is already running")
        return

    run([
        "minikube",
        "start",
        "--driver=docker",
        "--cpus=2",
        "--memory=2048",
        "--kubernetes-version=stable",
    ])

    run(["minikube", "addons", "enable", "metrics-server"], check=False)


WORKER_CODE = r'''
import hashlib
import os
import time

cpu_ms = int(os.getenv("CPU_MS", "50"))
sleep_ms = int(os.getenv("SLEEP_MS", "950"))
memory_mb = int(os.getenv("MEMORY_MB", "24"))
label = os.getenv("WORKLOAD_NAME", "worker")

block = bytearray(memory_mb * 1024 * 1024)
for i in range(0, len(block), 4096):
    block[i] = i % 251

payload = b"x" * 4096

print(
    f"starting {label}: cpu_ms={cpu_ms} sleep_ms={sleep_ms} memory_mb={memory_mb}",
    flush=True,
)

while True:
    end = time.perf_counter() + (cpu_ms / 1000.0)

    while time.perf_counter() < end:
        hashlib.sha256(payload).digest()

    for i in range(0, len(block), 4096 * 16):
        block[i] = (block[i] + 1) % 251

    time.sleep(sleep_ms / 1000.0)
'''.strip()


def make_deployment(name, cpu_ms, sleep_ms, memory_mb, req_cpu, lim_cpu, req_mem, lim_mem, replicas=1):
    return {
        "apiVersion": "apps/v1",
        "kind": "Deployment",
        "metadata": {
            "name": name,
            "namespace": NAMESPACE,
        },
        "spec": {
            "replicas": replicas,
            "selector": {
                "matchLabels": {
                    "app": name,
                }
            },
            "template": {
                "metadata": {
                    "labels": {
                        "app": name,
                    }
                },
                "spec": {
                    "containers": [
                        {
                            "name": "worker",
                            "image": "python:3.12-alpine",
                            "imagePullPolicy": "IfNotPresent",
                            "command": ["python", "-c", WORKER_CODE],
                            "env": [
                                {"name": "WORKLOAD_NAME", "value": name},
                                {"name": "CPU_MS", "value": str(cpu_ms)},
                                {"name": "SLEEP_MS", "value": str(sleep_ms)},
                                {"name": "MEMORY_MB", "value": str(memory_mb)},
                            ],
                            "resources": {
                                "requests": {
                                    "cpu": req_cpu,
                                    "memory": req_mem,
                                },
                                "limits": {
                                    "cpu": lim_cpu,
                                    "memory": lim_mem,
                                },
                            },
                        }
                    ]
                },
            },
        },
    }


def build_manifest():
    docs = [
        {
            "apiVersion": "v1",
            "kind": "Namespace",
            "metadata": {"name": NAMESPACE},
        },
        make_deployment(
            name="low-load",
            cpu_ms=20,
            sleep_ms=1500,
            memory_mb=16,
            req_cpu="20m",
            lim_cpu="80m",
            req_mem="40Mi",
            lim_mem="96Mi",
            replicas=1,
        ),
        make_deployment(
            name="medium-load",
            cpu_ms=80,
            sleep_ms=900,
            memory_mb=24,
            req_cpu="40m",
            lim_cpu="150m",
            req_mem="56Mi",
            lim_mem="128Mi",
            replicas=1,
        ),
        make_deployment(
            name="bursty-load",
            cpu_ms=200,
            sleep_ms=1800,
            memory_mb=32,
            req_cpu="50m",
            lim_cpu="200m",
            req_mem="64Mi",
            lim_mem="160Mi",
            replicas=1,
        ),
    ]
    return yaml.safe_dump_all(docs, sort_keys=False)


def apply_manifest(manifest_text):
    with tempfile.NamedTemporaryFile("w", suffix=".yaml", delete=False) as f:
        f.write(manifest_text)
        manifest_path = f.name

    try:
        run(["kubectl", "apply", "-f", manifest_path])
    finally:
        pass


def wait_for_rollouts():
    for name in ["low-load", "medium-load", "bursty-load"]:
        run([
            "kubectl",
            "rollout",
            "status",
            f"deployment/{name}",
            "-n",
            NAMESPACE,
            "--timeout=180s",
        ])


def show_status():
    run(["kubectl", "get", "pods", "-n", NAMESPACE, "-o", "wide"], check=False)
    print()
    print("Watch resource usage:")
    print(f"kubectl top pods -n {NAMESPACE}")
    print()
    print("Delete workloads:")
    print(f"kubectl delete namespace {NAMESPACE}")
    print()
    print("Stop cluster:")
    print("minikube stop")
    print("minikube delete")


def main():
    require_binary("minikube")
    require_binary("kubectl")
    require_binary("docker")

    start_minikube()
    manifest = build_manifest()
    apply_manifest(manifest)
    wait_for_rollouts()
    show_status()


if __name__ == "__main__":
    main()