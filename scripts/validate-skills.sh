#!/usr/bin/env bash

set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
skill_count=0

for skill_dir in "$repo_root"/skills/*; do
  [[ -d "$skill_dir" ]] || continue

  skill_count=$((skill_count + 1))
  folder_name="$(basename "$skill_dir")"
  skill_file="$skill_dir/SKILL.md"
  agent_file="$skill_dir/agents/openai.yaml"

  [[ -f "$skill_file" ]] || {
    echo "missing $skill_file" >&2
    exit 1
  }

  declared_name="$(sed -n 's/^name:[[:space:]]*//p' "$skill_file" | head -n 1 | tr -d '\"')"
  [[ "$declared_name" == "$folder_name" ]] || {
    echo "$skill_file declares '$declared_name'; expected '$folder_name'" >&2
    exit 1
  }

  grep -Eq '^description:[[:space:]]*"?.+"?$' "$skill_file" || {
    echo "$skill_file is missing a description" >&2
    exit 1
  }

  if grep -q 'TODO' "$skill_file"; then
    echo "$skill_file still contains TODO text" >&2
    exit 1
  fi

  [[ -f "$agent_file" ]] || {
    echo "missing $agent_file" >&2
    exit 1
  }

  grep -Fq "\$$folder_name" "$agent_file" || {
    echo "$agent_file default prompt must reference \$$folder_name" >&2
    exit 1
  }
done

[[ "$skill_count" -gt 0 ]] || {
  echo "no skills found" >&2
  exit 1
}

echo "validated $skill_count skill packages"
