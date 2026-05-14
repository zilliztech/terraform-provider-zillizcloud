#!/usr/bin/env python3
from __future__ import annotations

import json
import os
import subprocess
import sys
import urllib.error
import urllib.request


def git_short_commit() -> str:
    try:
        completed = subprocess.run(
            ["git", "rev-parse", "--short", "HEAD"],
            check=True,
            stdout=subprocess.PIPE,
            stderr=subprocess.DEVNULL,
            text=True,
        )
    except subprocess.CalledProcessError:
        return ""
    return completed.stdout.strip()


def main() -> None:
    webhook_url = os.environ.get("FEISHU_WEBHOOK_URL", "").strip()
    if not webhook_url:
        print("FEISHU_WEBHOOK_URL is empty; skipping Feishu notification")
        return

    publish_version = os.environ.get("PUBLISH_VERSION", "").strip()
    provider_name = os.environ.get("PROVIDER_NAME", "zillizcloud").strip()
    namespace = os.environ.get("TERRALIST_NAMESPACE", "zilliztech").strip()
    terralist_url = os.environ.get("TERRALIST_URL", "https://terralist.zilliz.cc").strip().rstrip("/")
    platforms = os.environ.get("PUBLISH_PLATFORMS", "").strip()
    job_name = os.environ.get("JOB_NAME", "terraform-provider-zillizcloud-terralist-publish").strip()
    build_number = os.environ.get("BUILD_NUMBER", "").strip()
    build_url = os.environ.get("BUILD_URL", "").strip()
    git_commit = git_short_commit()

    provider_host = terralist_url
    for prefix in ("https://", "http://"):
        if provider_host.startswith(prefix):
            provider_host = provider_host[len(prefix):]
            break

    provider_source = f"{provider_host}/{namespace}/{provider_name}"
    build_label = f"{job_name} #{build_number}" if build_number else job_name
    build_text = f"[{build_label}]({build_url})" if build_url else build_label
    commit_text = f"`{git_commit}`" if git_commit else "unknown"

    usage = f"""```hcl
terraform {{
  required_providers {{
    {provider_name} = {{
      source  = "{provider_source}"
      version = "{publish_version}"
    }}
  }}
}}

provider "{provider_name}" {{}}
```"""

    payload = {
        "msg_type": "interactive",
        "card": {
            "config": {
                "wide_screen_mode": True,
            },
            "header": {
                "template": "green",
                "title": {
                    "tag": "plain_text",
                    "content": f"Terraform Provider Published: {provider_name}",
                },
            },
            "elements": [
                {
                    "tag": "div",
                    "text": {
                        "tag": "lark_md",
                        "content": (
                            f"**Version:** `{publish_version}`\n"
                            f"**Provider source:** `{provider_source}`\n"
                            f"**Platforms:** `{platforms}`\n"
                            f"**Commit:** {commit_text}\n"
                            f"**Build:** {build_text}"
                        ),
                    },
                },
                {"tag": "hr"},
                {
                    "tag": "div",
                    "text": {
                        "tag": "lark_md",
                        "content": f"**How to use this provider**\n{usage}",
                    },
                },
                {
                    "tag": "div",
                    "text": {
                        "tag": "lark_md",
                        "content": f"Provider: {terralist_url}/providers/{namespace}/{provider_name}/{publish_version}",
                    },
                },
            ],
        },
    }

    request = urllib.request.Request(
        webhook_url,
        data=json.dumps(payload).encode("utf-8"),
        headers={"Content-Type": "application/json"},
        method="POST",
    )

    try:
        with urllib.request.urlopen(request, timeout=15) as response:
            body = response.read().decode("utf-8", errors="replace")
    except (urllib.error.URLError, TimeoutError) as exc:
        print(f"Feishu notification failed: {exc}", file=sys.stderr)
        return

    print(f"Feishu response: {body}")
    try:
        response_payload = json.loads(body)
    except json.JSONDecodeError:
        return

    if response_payload.get("code") not in (None, 0) or response_payload.get("StatusCode") not in (None, 0):
        print(f"Feishu notification returned non-success response: {body}", file=sys.stderr)


if __name__ == "__main__":
    main()
