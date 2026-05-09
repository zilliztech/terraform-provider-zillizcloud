#!/usr/bin/env python3
from __future__ import annotations

import argparse
import json
import urllib.parse
import urllib.request


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Fail if a provider version already exists in Terralist.")
    parser.add_argument("--terralist-url", required=True)
    parser.add_argument("--namespace", required=True)
    parser.add_argument("--provider-name", required=True)
    parser.add_argument("--version", required=True)
    return parser.parse_args()


def main() -> None:
    args = parse_args()
    base_url = args.terralist_url.rstrip("/")
    namespace = urllib.parse.quote(args.namespace)
    provider_name = urllib.parse.quote(args.provider_name)
    url = f"{base_url}/v1/providers/{namespace}/{provider_name}/versions"

    request = urllib.request.Request(url, method="GET")
    opener = urllib.request.build_opener(urllib.request.ProxyHandler({}))
    with opener.open(request, timeout=60) as response:
        payload = json.loads(response.read().decode("utf-8"))

    versions = payload.get("versions", []) if isinstance(payload, dict) else []
    published = {
        item.get("version")
        for item in versions
        if isinstance(item, dict) and item.get("version")
    }

    if args.version in published:
        raise SystemExit(f"version {args.version} already exists in Terralist")

    print(f"version {args.version} is not present in Terralist")


if __name__ == "__main__":
    main()
