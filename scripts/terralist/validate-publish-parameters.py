#!/usr/bin/env python3
import os
import re


VERSION_PATTERN = re.compile(r"^[0-9]+[.][0-9]+[.][0-9]+(-[0-9A-Za-z][0-9A-Za-z.-]*)?$")
PLATFORM_PATTERN = re.compile(r"^[a-z0-9]+/[a-z0-9_]+$")


def fail(message: str) -> None:
    raise SystemExit(message)


def validate_version(raw_version: str) -> str:
    version = raw_version.strip()
    if not version:
        fail("VERSION is required")
    if version.startswith("v"):
        version = version[1:]
    if not VERSION_PATTERN.fullmatch(version):
        fail(f"VERSION must look like 0.6.37 or 0.6.37-rc5, got: {version}")
    return version


def validate_platforms(raw_platforms: str) -> str:
    platforms = [platform.strip() for platform in raw_platforms.split(",") if platform.strip()]
    if not platforms:
        fail("PLATFORMS must include at least one GOOS/GOARCH target")

    for platform in platforms:
        if not PLATFORM_PATTERN.fullmatch(platform):
            fail(f"Invalid platform '{platform}', expected GOOS/GOARCH")

    return ",".join(platforms)


def main() -> None:
    publish_version = validate_version(os.environ.get("VERSION", ""))
    publish_platforms = validate_platforms(os.environ.get("PLATFORMS", ""))

    print(f"PUBLISH_VERSION={publish_version}")
    print(f"PUBLISH_PLATFORMS={publish_platforms}")


if __name__ == "__main__":
    main()
