#!/usr/bin/env bash
set -euo pipefail

package=""
run_pattern="TestAccOnDemandClusterResource"
timeout="120m"
count="1"
env_file=""
extra_args=()

usage() {
  cat <<'USAGE'
Usage: run_acc_test.sh [options] [-- extra go test args]

Options:
  --package PATH    Go package pattern to test. Default: locate test with rg
  --run REGEX      Test -run pattern. Default: TestAccOnDemandClusterResource
  --timeout VALUE  Go test timeout. Default: 120m
  --count VALUE    Go test count. Default: 1
  --env-file PATH  Environment file to source. Default: .env.test, then .evn.test
  -h, --help       Show this help.
USAGE
}

locate_package() {
  local pattern="$1"
  local test_name="$pattern"
  local matches=""
  local first_file=""
  local second_file=""
  local test_dir=""

  test_name="${test_name#^}"
  test_name="${test_name%$}"

  if [[ ! "$test_name" =~ ^[A-Za-z_][A-Za-z0-9_]*$ ]]; then
    echo "Cannot auto-locate package for non-exact -run pattern: $pattern" >&2
    echo "Pass --package explicitly." >&2
    exit 2
  fi

  if ! command -v rg >/dev/null 2>&1; then
    echo "Cannot auto-locate package because rg is not installed. Pass --package explicitly." >&2
    exit 2
  fi

  matches="$(rg -l --glob '*_test.go' "func[[:space:]]+${test_name}[[:space:]]*\\(" . || true)"
  if [[ -z "$matches" ]]; then
    echo "Could not find test function $test_name with rg." >&2
    echo "Pass --package explicitly or check the -run pattern." >&2
    exit 1
  fi

  first_file="$(printf '%s\n' "$matches" | sed -n '1p')"
  second_file="$(printf '%s\n' "$matches" | sed -n '2p')"
  if [[ -n "$second_file" ]]; then
    echo "Found multiple files for $test_name:" >&2
    printf '%s\n' "$matches" >&2
    echo "Pass --package explicitly." >&2
    exit 2
  fi

  test_dir="$(dirname "$first_file")"
  test_dir="${test_dir#./}"
  echo "Located test file with rg: $first_file" >&2
  printf './%s' "$test_dir"
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --package)
      package="${2:?missing value for --package}"
      shift 2
      ;;
    --run)
      run_pattern="${2:?missing value for --run}"
      shift 2
      ;;
    --timeout)
      timeout="${2:?missing value for --timeout}"
      shift 2
      ;;
    --count)
      count="${2:?missing value for --count}"
      shift 2
      ;;
    --env-file)
      env_file="${2:?missing value for --env-file}"
      shift 2
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    --)
      shift
      extra_args=("$@")
      break
      ;;
    *)
      echo "Unknown argument: $1" >&2
      usage >&2
      exit 2
      ;;
  esac
done

if [[ -z "$package" ]]; then
  package="$(locate_package "$run_pattern")"
  echo "Located test package with rg: $package"
fi

if [[ -z "$env_file" ]]; then
  if [[ -f ".env.test" ]]; then
    env_file=".env.test"
  elif [[ -f ".evn.test" ]]; then
    env_file=".evn.test"
  else
    echo "No test env file found. Expected .env.test or .evn.test in $(pwd)." >&2
    exit 1
  fi
fi

if [[ ! -f "$env_file" ]]; then
  echo "Env file not found: $env_file" >&2
  exit 1
fi

set -a
# shellcheck source=/dev/null
source "$env_file"
set +a

echo "Loaded env file: $env_file"
echo "Running: TF_ACC=1 go test -count=$count $package -run $run_pattern -v -timeout $timeout ${extra_args[*]-}"

TF_ACC=1 go test -count="$count" "$package" -run "$run_pattern" -v -timeout "$timeout" "${extra_args[@]}"
