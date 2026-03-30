#!/usr/bin/env python3
"""Validate Intercom Go SDK endpoint coverage against OpenAPI spec.

Parses raw/api.intercom.io.yaml to extract all operations, scans *.go files
to extract SDK methods with their HTTP method + URL path from NewRequest calls,
and produces an endpoint coverage mapping report.

Extended to also compare request/response struct fields between the OpenAPI
schema and the Go SDK structs (Task 2).
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


def resolve_ref(spec, ref):
    """Resolve a $ref string to the actual schema object."""
    if not ref.startswith("#/"):
        return {}
    parts = ref.lstrip("#/").split("/")
    obj = spec
    for p in parts:
        obj = obj.get(p, {})
    return obj


def get_schema_properties(spec, schema):
    """Extract property names and types from a schema, following $ref chains.

    Returns a dict of {property_name: {"type": ..., "required": bool}}.
    Handles $ref, oneOf (takes first), allOf (merges), and inline properties.
    """
    if not schema:
        return {}

    # Follow $ref
    if "$ref" in schema:
        resolved = resolve_ref(spec, schema["$ref"])
        return get_schema_properties(spec, resolved)

    # oneOf: take the first option, plus any top-level properties
    if "oneOf" in schema:
        merged = {}
        for option in schema["oneOf"]:
            props = get_schema_properties(spec, option)
            if props:
                merged = props
                break
        # Also include any top-level properties alongside oneOf
        if "properties" in schema:
            top_required = set(schema.get("required", []))
            for name, prop in schema["properties"].items():
                prop_type = prop.get("type", "")
                if "$ref" in prop:
                    prop_type = f"ref:{prop['$ref'].split('/')[-1]}"
                merged[name] = {
                    "type": prop_type,
                    "required": name in top_required,
                    "nullable": prop.get("nullable", False),
                }
        return merged

    # allOf: merge all schemas, plus any top-level properties
    if "allOf" in schema:
        merged = {}
        for sub in schema["allOf"]:
            merged.update(get_schema_properties(spec, sub))
        # Also include any top-level properties alongside allOf
        if "properties" in schema:
            top_required = set(schema.get("required", []))
            for name, prop in schema["properties"].items():
                prop_type = prop.get("type", "")
                if "$ref" in prop:
                    prop_type = f"ref:{prop['$ref'].split('/')[-1]}"
                merged[name] = {
                    "type": prop_type,
                    "required": name in top_required,
                    "nullable": prop.get("nullable", False),
                }
        return merged

    # Inline properties
    properties = schema.get("properties", {})
    required_fields = set(schema.get("required", []))
    result = {}
    for name, prop in properties.items():
        prop_type = prop.get("type", "")
        if "$ref" in prop:
            ref_name = prop["$ref"].split("/")[-1]
            prop_type = f"ref:{ref_name}"
        elif "oneOf" in prop:
            prop_type = "oneOf"
        elif prop_type == "array":
            items = prop.get("items", {})
            if "$ref" in items:
                item_type = items["$ref"].split("/")[-1]
                prop_type = f"array[ref:{item_type}]"
            else:
                prop_type = f"array[{items.get('type', 'any')}]"
        result[name] = {
            "type": prop_type,
            "required": name in required_fields,
            "nullable": prop.get("nullable", False),
        }
    return result


def extract_request_body_schema(spec, operation):
    """Extract request body schema properties for an operation."""
    rb = operation.get("requestBody", {})
    content = rb.get("content", {}).get("application/json", {})
    schema = content.get("schema", {})
    return get_schema_properties(spec, schema)


def extract_response_schema(spec, operation):
    """Extract the success (2xx) response schema properties for an operation."""
    responses = operation.get("responses", {})
    for status in ("200", "201", "202"):
        if status in responses:
            resp = responses[status]
            content = resp.get("content", {}).get("application/json", {})
            schema = content.get("schema", {})
            return get_schema_properties(spec, schema)
    return {}


def snake_to_pascal(name):
    """Convert snake_case to PascalCase (for matching schema names to Go structs)."""
    return "".join(word.capitalize() for word in name.split("_"))


def extract_go_structs(root_dir):
    """Extract all Go struct definitions with their JSON-tagged fields.

    Returns a dict of {StructName: {"file": str, "fields": {json_name: {"go_name": str, "go_type": str}}}}.
    """
    structs = {}
    for filename in sorted(os.listdir(root_dir)):
        if not filename.endswith(".go") or filename.endswith("_test.go"):
            continue
        filepath = os.path.join(root_dir, filename)
        with open(filepath, "r") as f:
            content = f.read()

        # Find struct definitions
        pattern = re.compile(r'type\s+(\w+)\s+struct\s*\{', re.MULTILINE)
        for m in pattern.finditer(content):
            name = m.group(1)
            # Find matching closing brace (handles nested structs)
            start = m.end()
            depth = 1
            i = start
            while i < len(content) and depth > 0:
                if content[i] == '{':
                    depth += 1
                elif content[i] == '}':
                    depth -= 1
                i += 1
            body = content[start:i - 1]

            fields = {}
            for line in body.split("\n"):
                line = line.strip()
                # Match: FieldName Type `json:"tag"` or `json:"tag" url:"..."`
                fm = re.match(
                    r'(\w+)\s+(\S+(?:\s*\{[^}]*\})?)\s+`(?:[^`]*?)json:"([^"]+)"',
                    line,
                )
                if fm:
                    go_name = fm.group(1)
                    go_type = fm.group(2)
                    json_tag = fm.group(3)
                    json_name = json_tag.split(",")[0]
                    if json_name == "-":
                        continue
                    fields[json_name] = {
                        "go_name": go_name,
                        "go_type": go_type,
                        "omitempty": "omitempty" in json_tag,
                    }
            if fields:
                structs[name] = {"file": filename, "fields": fields}

    return structs


def collect_refs_from_schema(schema):
    """Collect all $ref names from a schema (including allOf, oneOf)."""
    refs = []
    if "$ref" in schema:
        refs.append(schema["$ref"].split("/")[-1])
    for key in ("allOf", "oneOf", "anyOf"):
        for sub in schema.get(key, []):
            refs.extend(collect_refs_from_schema(sub))
    return refs


def find_sdk_struct_for_operation(sdk_method_info, go_structs, spec, operation):
    """Try to find the Go request struct that corresponds to an OpenAPI operation.

    Uses multiple strategies:
    1. Match by $ref schema name -> PascalCase Go struct name (including allOf/oneOf)
    2. Match by SDK method name convention (e.g., Create -> CreateXxxRequest)
    """
    rb = operation.get("requestBody", {})
    content = rb.get("content", {}).get("application/json", {})
    schema = content.get("schema", {})

    # Strategy 1: match by $ref name (including from allOf/oneOf)
    refs = collect_refs_from_schema(schema)
    for ref_name in refs:
        pascal = snake_to_pascal(ref_name)
        if pascal in go_structs:
            return pascal

    # Strategy 2: match by SDK method name convention
    if sdk_method_info:
        method_name = sdk_method_info["method"]
        service = sdk_method_info["service"]
        # Strip "Service" suffix and "Raw" suffix
        service_base = service.replace("Service", "")
        method_base = method_name.replace("Raw", "")

        # Common patterns: CreateContactRequest, UpdateArticleRequest, etc.
        candidates = [
            f"{method_base}{service_base}Request",
            f"{method_base}Request",
        ]
        for candidate in candidates:
            if candidate in go_structs:
                return candidate

    return None


def compare_fields(schema_props, go_fields):
    """Compare schema properties against Go struct fields.

    Returns: (missing_from_sdk, extra_in_sdk, type_notes)
    """
    missing = []  # In schema but not in Go struct
    extra = []  # In Go struct but not in schema
    type_notes = []  # Type mismatch notes

    schema_names = set(schema_props.keys())
    go_names = set(go_fields.keys())

    for name in sorted(schema_names - go_names):
        prop = schema_props[name]
        missing.append({
            "name": name,
            "schema_type": prop["type"],
            "required": prop["required"],
        })

    for name in sorted(go_names - schema_names):
        field = go_fields[name]
        extra.append({
            "name": name,
            "go_type": field["go_type"],
        })

    # Type comparison for common fields
    for name in sorted(schema_names & go_names):
        schema_type = schema_props[name]["type"]
        go_type = go_fields[name]["go_type"]
        # Basic type mapping check
        note = check_type_compatibility(schema_type, go_type)
        if note:
            type_notes.append({
                "name": name,
                "schema_type": schema_type,
                "go_type": go_type,
                "note": note,
            })

    return missing, extra, type_notes


def check_type_compatibility(schema_type, go_type):
    """Check if a Go type is compatible with an OpenAPI schema type.

    Returns a note string if there's a potential mismatch, None if compatible.
    """
    compatible = {
        "string": {"string", "*string"},
        "integer": {"int", "int64", "*int64", "int32", "*int32", "*int"},
        "number": {"float64", "*float64", "float32", "*float32", "int", "int64", "*int64"},
        "boolean": {"bool", "*bool"},
        "object": {"map[string]interface{}", "map[string]any", "json.RawMessage"},
    }

    # Skip complex types (refs, arrays, oneOf) — they need deeper analysis
    if schema_type.startswith("ref:") or schema_type.startswith("array[") or schema_type == "oneOf":
        return None

    expected_go_types = compatible.get(schema_type)
    if expected_go_types is None:
        return None  # Unknown schema type, skip

    if go_type not in expected_go_types:
        # Check if it's a pointer variant or slice
        if go_type.startswith("*") and go_type[1:] in expected_go_types:
            return None
        if go_type.startswith("[]"):
            return None  # Array types handled separately
        return f"possible mismatch: schema={schema_type}, go={go_type}"

    return None


def build_field_validation(spec, operations, matched, go_structs):
    """Build field-level validation comparing schema properties to Go struct fields.

    Returns a list of validation results per operation.
    """
    results = []

    for m in matched:
        op = m["operation"]
        sdk = m["sdk"]
        operation_id = op["operationId"]

        # Get the full operation from the spec
        path_item = spec.get("paths", {}).get(op["path"], {})
        full_op = path_item.get(op["method"].lower(), {})

        # Request body comparison
        request_result = None
        if "requestBody" in full_op:
            schema_props = extract_request_body_schema(spec, full_op)
            struct_name = find_sdk_struct_for_operation(sdk, go_structs, spec, full_op)

            if struct_name and struct_name in go_structs:
                go_fields = go_structs[struct_name]["fields"]
                missing, extra, type_notes = compare_fields(schema_props, go_fields)
                request_result = {
                    "struct_name": struct_name,
                    "struct_file": go_structs[struct_name]["file"],
                    "schema_field_count": len(schema_props),
                    "go_field_count": len(go_fields),
                    "missing_from_sdk": missing,
                    "extra_in_sdk": extra,
                    "type_notes": type_notes,
                }
            elif schema_props:
                request_result = {
                    "struct_name": None,
                    "schema_field_count": len(schema_props),
                    "note": f"No matching Go struct found for {operation_id}",
                }

        results.append({
            "operation_id": operation_id,
            "method": op["method"],
            "path": op["path"],
            "sdk_service": sdk["service"],
            "sdk_method": sdk["method"],
            "request": request_result,
        })

    return results


def format_field_report(field_results):
    """Format the field validation report as markdown."""
    lines = []
    lines.append("# OpenAPI Full Validation Report")
    lines.append("")
    lines.append("Generated by `scripts/validate_openapi.py`")
    lines.append("")

    # Summary statistics
    total_ops = len(field_results)
    ops_with_request = sum(1 for r in field_results if r["request"])
    ops_matched = sum(1 for r in field_results if r["request"] and r["request"].get("struct_name"))
    ops_with_missing = sum(
        1 for r in field_results
        if r["request"] and r["request"].get("missing_from_sdk")
    )
    ops_with_extra = sum(
        1 for r in field_results
        if r["request"] and r["request"].get("extra_in_sdk")
    )
    ops_with_type_notes = sum(
        1 for r in field_results
        if r["request"] and r["request"].get("type_notes")
    )
    total_missing = sum(
        len(r["request"]["missing_from_sdk"])
        for r in field_results
        if r["request"] and r["request"].get("missing_from_sdk")
    )
    total_extra = sum(
        len(r["request"]["extra_in_sdk"])
        for r in field_results
        if r["request"] and r["request"].get("extra_in_sdk")
    )

    lines.append("## Summary")
    lines.append("")
    lines.append(f"- Operations analyzed: {total_ops}")
    lines.append(f"- Operations with request body: {ops_with_request}")
    lines.append(f"- Request structs matched: {ops_matched}")
    lines.append(f"- Operations with missing fields: {ops_with_missing} ({total_missing} total fields)")
    lines.append(f"- Operations with extra SDK fields: {ops_with_extra} ({total_extra} total fields)")
    lines.append(f"- Operations with type notes: {ops_with_type_notes}")
    lines.append("")

    # Missing fields detail
    if ops_with_missing:
        lines.append("## Missing Fields (in schema but not in SDK)")
        lines.append("")
        for r in field_results:
            req = r.get("request")
            if not req or not req.get("missing_from_sdk"):
                continue
            lines.append(
                f"### `{r['method']} {r['path']}` ({r['operation_id']}) "
                f"-> {r['sdk_service']}.{r['sdk_method']}"
            )
            lines.append(f"Go struct: `{req['struct_name']}` in `{req['struct_file']}`")
            lines.append("")
            for field in req["missing_from_sdk"]:
                req_marker = " **[required]**" if field["required"] else ""
                lines.append(f"- `{field['name']}` ({field['schema_type']}){req_marker}")
            lines.append("")

    # Extra fields detail
    if ops_with_extra:
        lines.append("## Extra SDK Fields (in SDK but not in schema)")
        lines.append("")
        for r in field_results:
            req = r.get("request")
            if not req or not req.get("extra_in_sdk"):
                continue
            lines.append(
                f"### `{r['method']} {r['path']}` ({r['operation_id']}) "
                f"-> {r['sdk_service']}.{r['sdk_method']}"
            )
            lines.append(f"Go struct: `{req['struct_name']}` in `{req['struct_file']}`")
            lines.append("")
            for field in req["extra_in_sdk"]:
                lines.append(f"- `{field['name']}` ({field['go_type']})")
            lines.append("")

    # Type notes
    if ops_with_type_notes:
        lines.append("## Type Notes")
        lines.append("")
        for r in field_results:
            req = r.get("request")
            if not req or not req.get("type_notes"):
                continue
            lines.append(
                f"### `{r['method']} {r['path']}` ({r['operation_id']}) "
                f"-> {r['sdk_service']}.{r['sdk_method']}"
            )
            lines.append("")
            for note in req["type_notes"]:
                lines.append(
                    f"- `{note['name']}`: {note['note']}"
                )
            lines.append("")

    # Unmatched request structs
    unmatched = [
        r for r in field_results
        if r["request"] and not r["request"].get("struct_name") and r["request"].get("note")
    ]
    if unmatched:
        lines.append("## Unmatched Operations (request body but no Go struct found)")
        lines.append("")
        for r in unmatched:
            lines.append(
                f"- `{r['method']} {r['path']}` ({r['operation_id']}): "
                f"{r['request']['note']}"
            )
        lines.append("")

    # Clean operations (matched with no issues)
    clean = [
        r for r in field_results
        if r["request"] and r["request"].get("struct_name")
        and not r["request"].get("missing_from_sdk")
        and not r["request"].get("extra_in_sdk")
        and not r["request"].get("type_notes")
    ]
    if clean:
        lines.append("## Clean Operations (fields match)")
        lines.append("")
        for r in clean:
            req = r["request"]
            lines.append(
                f"- `{r['method']} {r['path']}` ({r['operation_id']}) "
                f"-> `{req['struct_name']}` ({req['schema_field_count']} fields)"
            )
        lines.append("")

    return "\n".join(lines)


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
    full_report_path = os.path.join(root_dir, "docs", "reports", "openapi-full-validation.md")

    if not os.path.exists(spec_path):
        print(f"Error: OpenAPI spec not found at {spec_path}", file=sys.stderr)
        sys.exit(1)

    print(f"Loading OpenAPI spec from {spec_path}...")
    with open(spec_path, "r") as f:
        spec = yaml.safe_load(f)

    operations = load_openapi_spec(spec_path)
    print(f"Found {len(operations)} operations")

    print(f"Scanning Go source files in {root_dir}...")
    sdk_methods = scan_go_sources(root_dir)
    print(f"Found {len(sdk_methods)} SDK methods with NewRequest calls")

    matched, missing, extra_sdk = build_coverage_report(operations, sdk_methods)

    # Endpoint coverage report
    report = format_report(operations, matched, missing, extra_sdk)

    os.makedirs(os.path.dirname(report_path), exist_ok=True)
    with open(report_path, "w") as f:
        f.write(report)

    print(f"\nEndpoint report saved to {report_path}")
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

    # Field-level validation (Task 2)
    print(f"\n--- Field-level validation ---")
    print(f"Extracting Go struct definitions...")
    go_structs = extract_go_structs(root_dir)
    print(f"Found {len(go_structs)} Go structs with JSON fields")

    print(f"Comparing request body schemas to Go structs...")
    field_results = build_field_validation(spec, operations, matched, go_structs)

    full_report = format_field_report(field_results)
    with open(full_report_path, "w") as f:
        f.write(full_report)

    print(f"Full validation report saved to {full_report_path}")

    # Print field validation summary
    ops_with_issues = sum(
        1 for r in field_results
        if r["request"] and (
            r["request"].get("missing_from_sdk")
            or r["request"].get("extra_in_sdk")
            or r["request"].get("type_notes")
        )
    )
    print(f"Operations with field issues: {ops_with_issues}")


if __name__ == "__main__":
    main()
