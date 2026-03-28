#!/usr/bin/env python3
"""Validate Intercom Go SDK endpoint coverage against OpenAPI spec.

Parses raw/api.intercom.io.yaml to extract all operations, scans *.go files
to extract SDK methods with their HTTP method + URL path from NewRequest calls,
and produces an endpoint coverage mapping report.
"""

import os
import re
import sys
from collections import defaultdict

import yaml


def load_openapi_spec(path):
    """Parse the OpenAPI YAML file and extract all operations."""
    with open(path, "r") as f:
        spec = yaml.safe_load(f)

    operations = []
    for path_str, path_item in spec.get("paths", {}).items():
        for method in ("get", "post", "put", "patch", "delete"):
            if method not in path_item:
                continue
            op = path_item[method]
            operations.append({
                "method": method.upper(),
                "path": path_str,
                "operationId": op.get("operationId", ""),
                "tags": op.get("tags", []),
                "summary": op.get("summary", ""),
                "has_request_body": "requestBody" in op,
                "responses": list(op.get("responses", {}).keys()),
                "parameters": [
                    p.get("name") for p in op.get("parameters", [])
                    if p.get("name") != "Intercom-Version"
                ],
            })
    return operations


def normalize_path(go_path):
    """Normalize a Go SDK URL path to match OpenAPI path format.

    Converts patterns like:
      fmt.Sprintf("contacts/%s/tags", ...) -> /contacts/{param}/tags
      fmt.Sprintf("data_attributes/%d", ...) -> /data_attributes/{param}
      "admins" -> /admins
      "visitors?user_id=%s" -> /visitors
    """
    # Already normalized
    if go_path.startswith("/"):
        result = go_path
    else:
        result = "/" + go_path

    # Strip query string
    if "?" in result:
        result = result.split("?")[0]

    # Replace format verbs (%s, %d, %v, etc.) with {param} placeholder
    param_idx = 0
    while re.search(r'%[sdvfgq]', result):
        result = re.sub(r'%[sdvfgq]', "{param" + str(param_idx) + "}", result, count=1)
        param_idx += 1
    return result


def scan_go_sources(root_dir):
    """Scan Go source files for NewRequest calls to extract SDK method -> HTTP endpoint mapping."""
    sdk_methods = []

    for filename in sorted(os.listdir(root_dir)):
        if not filename.endswith(".go") or filename.endswith("_test.go"):
            continue
        if filename in ("doc.go", "intercom.go", "pagination.go", "search.go",
                        "errors.go", "result.go", "testutil_test.go"):
            continue

        filepath = os.path.join(root_dir, filename)
        with open(filepath, "r") as f:
            content = f.read()

        # Find all functions/methods with NewRequest calls
        # Pattern: func (s *XxxService) MethodName(...) ... NewRequest(http.MethodXxx, "path", ...)
        # or NewRequest(http.MethodXxx, fmt.Sprintf("path/%s/...", ...), ...)

        # Split by function boundaries
        func_pattern = re.compile(
            r'func\s+\(s\s+\*(\w+Service)\)\s+(\w+)\s*\([^)]*\)[^{]*\{',
            re.MULTILINE
        )

        # Pattern 1: NewRequest(http.MethodXxx, "path", ...) or fmt.Sprintf(...)
        new_request_inline = re.compile(
            r'NewRequest\(\s*http\.Method(\w+)\s*,\s*'
            r'(?:fmt\.Sprintf\(\s*"([^"]+)"'  # fmt.Sprintf("path", ...)
            r'|"([^"]+)")'                      # or literal "path"
        )

        # Pattern 2: NewRequest(http.MethodXxx, path, ...) where path is a variable
        new_request_var = re.compile(
            r'NewRequest\(\s*http\.Method(\w+)\s*,\s*path\s*,'
        )

        # Patterns to resolve `path` variable assignment
        add_query_opts = re.compile(
            r'path\s*,\s*(?:_|err)\s*:?=\s*addQueryOptions\(\s*"([^"]+)"'
        )
        add_query_opts_sprintf = re.compile(
            r'path\s*,\s*(?:_|err)\s*:?=\s*addQueryOptions\(\s*fmt\.Sprintf\(\s*"([^"]+)"'
        )
        path_sprintf = re.compile(
            r'path\s*:?=\s*fmt\.Sprintf\(\s*"([^"]+)"'
        )
        path_literal = re.compile(
            r'path\s*:?=\s*"([^"]+)"'
        )

        matches = list(func_pattern.finditer(content))
        for i, match in enumerate(matches):
            service_name = match.group(1)
            method_name = match.group(2)
            func_start = match.start()

            # Find the end of this function (next func or EOF)
            if i + 1 < len(matches):
                func_end = matches[i + 1].start()
            else:
                func_end = len(content)

            func_body = content[func_start:func_end]

            # Try Pattern 1: inline path in NewRequest
            nr_match = new_request_inline.search(func_body)
            if nr_match:
                http_method = nr_match.group(1).upper()
                raw_path = nr_match.group(2) or nr_match.group(3)
                normalized = normalize_path(raw_path)
            else:
                # Try Pattern 2: path variable
                nr_var = new_request_var.search(func_body)
                if not nr_var:
                    continue
                http_method = nr_var.group(1).upper()

                # Resolve path variable - try patterns in order of specificity
                raw_path = None
                for pat in [add_query_opts_sprintf, add_query_opts,
                            path_sprintf, path_literal]:
                    m = pat.search(func_body)
                    if m:
                        raw_path = m.group(1)
                        break

                if not raw_path:
                    continue
                normalized = normalize_path(raw_path)

            # Map Go http.Method constants
            method_map = {"GET": "GET", "POST": "POST", "PUT": "PUT",
                          "PATCH": "PATCH", "DELETE": "DELETE"}
            http_method = method_map.get(http_method, http_method)

            sdk_methods.append({
                "service": service_name,
                "method": method_name,
                "http_method": http_method,
                "raw_path": raw_path,
                "normalized_path": normalized,
                "file": filename,
            })

    return sdk_methods


def match_path(openapi_path, sdk_path):
    """Check if an OpenAPI path matches an SDK normalized path.

    OpenAPI: /contacts/{contact_id}/tags
    SDK:     /contacts/{param0}/tags
    """
    # Split both paths and compare segment by segment
    openapi_parts = openapi_path.strip("/").split("/")
    sdk_parts = sdk_path.strip("/").split("/")

    if len(openapi_parts) != len(sdk_parts):
        return False

    for op, sp in zip(openapi_parts, sdk_parts):
        # Both are path params
        if op.startswith("{") and sp.startswith("{"):
            continue
        # Both are literals and must match
        if op == sp:
            continue
        return False

    return True


def build_coverage_report(operations, sdk_methods):
    """Match OpenAPI operations to SDK methods and produce coverage report."""
    matched = []
    missing = []
    extra_sdk = list(sdk_methods)  # Will remove matched ones

    for op in operations:
        found = None
        for sdk in extra_sdk:
            if (sdk["http_method"] == op["method"] and
                    match_path(op["path"], sdk["normalized_path"])):
                found = sdk
                break

        if found:
            matched.append({"operation": op, "sdk": found})
            extra_sdk.remove(found)
        else:
            missing.append(op)

    return matched, missing, extra_sdk


def format_report(operations, matched, missing, extra_sdk):
    """Format the coverage report as markdown."""
    lines = []
    lines.append("# OpenAPI Endpoint Coverage Report")
    lines.append("")
    lines.append(f"Generated by `scripts/validate_openapi.py`")
    lines.append("")
    lines.append("## Summary")
    lines.append("")
    lines.append(f"- Total OpenAPI operations: {len(operations)}")
    lines.append(f"- Matched to SDK methods: {len(matched)}")
    lines.append(f"- Missing from SDK: {len(missing)}")
    lines.append(f"- Extra SDK methods (not in schema): {len(extra_sdk)}")
    pct = len(matched) / len(operations) * 100 if operations else 0
    lines.append(f"- Coverage: {pct:.1f}%")
    lines.append("")

    # Group matched by tag
    lines.append("## Matched Operations")
    lines.append("")
    by_tag = defaultdict(list)
    for m in matched:
        tag = m["operation"]["tags"][0] if m["operation"]["tags"] else "Untagged"
        by_tag[tag].append(m)

    for tag in sorted(by_tag.keys()):
        lines.append(f"### {tag}")
        lines.append("")
        for m in sorted(by_tag[tag], key=lambda x: x["operation"]["path"]):
            op = m["operation"]
            sdk = m["sdk"]
            lines.append(
                f"- `{op['method']} {op['path']}` ({op['operationId']}) "
                f"-> {sdk['service']}.{sdk['method']}"
            )
        lines.append("")

    if missing:
        lines.append("## Missing from SDK")
        lines.append("")
        for op in sorted(missing, key=lambda x: (x["tags"][0] if x["tags"] else "", x["path"])):
            tag = op["tags"][0] if op["tags"] else "Untagged"
            lines.append(
                f"- **{op['method']} {op['path']}** ({op['operationId']}) "
                f"[{tag}] — {op['summary']}"
            )
        lines.append("")

    if extra_sdk:
        lines.append("## Extra SDK Methods (not in schema)")
        lines.append("")
        for sdk in sorted(extra_sdk, key=lambda x: (x["service"], x["method"])):
            lines.append(
                f"- {sdk['service']}.{sdk['method']} — "
                f"`{sdk['http_method']} {sdk['raw_path']}`"
            )
        lines.append("")

    return "\n".join(lines)


def main():
    root_dir = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
    spec_path = os.path.join(root_dir, "raw", "api.intercom.io.yaml")
    report_path = os.path.join(root_dir, "docs", "reports", "openapi-endpoint-coverage.md")

    if not os.path.exists(spec_path):
        print(f"Error: OpenAPI spec not found at {spec_path}", file=sys.stderr)
        sys.exit(1)

    print(f"Loading OpenAPI spec from {spec_path}...")
    operations = load_openapi_spec(spec_path)
    print(f"Found {len(operations)} operations")

    print(f"Scanning Go source files in {root_dir}...")
    sdk_methods = scan_go_sources(root_dir)
    print(f"Found {len(sdk_methods)} SDK methods with NewRequest calls")

    matched, missing, extra_sdk = build_coverage_report(operations, sdk_methods)

    report = format_report(operations, matched, missing, extra_sdk)

    os.makedirs(os.path.dirname(report_path), exist_ok=True)
    with open(report_path, "w") as f:
        f.write(report)

    print(f"\nReport saved to {report_path}")
    print(f"\nCoverage: {len(matched)}/{len(operations)} "
          f"({len(matched)/len(operations)*100:.1f}%)")

    if missing:
        print(f"\nMissing operations ({len(missing)}):")
        for op in missing:
            print(f"  - {op['method']} {op['path']} ({op['operationId']})")

    if extra_sdk:
        print(f"\nExtra SDK methods ({len(extra_sdk)}):")
        for sdk in extra_sdk:
            print(f"  - {sdk['service']}.{sdk['method']}")


if __name__ == "__main__":
    main()
