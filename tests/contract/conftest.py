"""契约测试共享 fixture。

提供：
  - base_url           后端 base URL（默认 http://localhost:8080/api/v1）
  - spec               加载 contracts/openapi/v1.yaml
  - validator          校验 OpenAPI 3.1 合法
  - schema_for()       按 path+method 取出 response schema（支持 path 参数）
  - login_token()      登录 string/123456，返回 JWT
"""
import os
import re
import pytest
import requests
import yaml

from openapi_spec_validator import validate as validate_spec
from jsonschema import Draft7Validator


CONTRACT_PATH = "contracts/openapi/v1.yaml"


@pytest.fixture(scope="session")
def base_url() -> str:
    return os.environ.get("API_BASE_URL", "http://localhost:8080/api/v1")


@pytest.fixture(scope="session")
def spec():
    """加载权威契约 YAML。"""
    with open(CONTRACT_PATH) as f:
        return yaml.safe_load(f)


@pytest.fixture(scope="session")
def spec_is_valid(spec):
    """契约文件本身必须合法（OpenAPI 3.1）。"""
    validate_spec(spec)
    return True


@pytest.fixture
def http(base_url):
    """requests.Session，预设 base_url。"""
    s = requests.Session()
    s.headers.update({"Content-Type": "application/json"})
    return s


@pytest.fixture(scope="session")
def login_token(base_url):
    """用 string/123456 登录拿到 JWT；session 级只登录一次。"""
    resp = requests.post(
        f"{base_url}/user/login",
        json={"username": "string", "password": "123456"},
        timeout=5,
    )
    resp.raise_for_status()
    body = resp.json()
    assert body.get("code") == 200, f"login failed: {body}"
    return body["data"]["token"]


@pytest.fixture
def auth_headers(login_token):
    return {"Authorization": f"Bearer {login_token}"}


def resolve_response_schema(spec, path, method, status="200"):
    """按 (path, method, status) 取出 response schema（递归解析所有 $ref）。

    支持路径模板（如 /manuscript/detail/{id}）匹配具体路径。
    返回的 schema 是内联的（所有 $ref 已展开）。
    """
    path_item = spec["paths"].get(path)
    if path_item is None:
        for template_path, item in spec["paths"].items():
            if "{" in template_path:
                pattern = re.sub(r"\{[^}]+\}", r"[^/]+", template_path)
                if re.fullmatch(pattern, path):
                    path_item = item
                    break
    if path_item is None:
        raise KeyError(f"path not found in spec: {path}")

    op = path_item.get(method.lower())
    if op is None:
        raise KeyError(f"method {method} not on {path}")

    resp = op.get("responses", {}).get(status)
    if resp is None:
        raise KeyError(f"status {status} not declared for {method.upper()} {path}")

    # $ref 指向 components/responses/...
    if "$ref" in resp:
        ref_path = resp["$ref"].lstrip("#/").split("/")
        node = spec
        for key in ref_path:
            node = node[key]
        resp = node

    schema = resp.get("content", {}).get("application/json", {}).get("schema")
    if schema is None:
        return None
    return _inline(spec, schema)


def _inline(spec, schema):
    """递归把所有 $ref 展开为内联 schema（合并 allOf）。"""
    if not isinstance(schema, dict):
        return schema
    if "$ref" in schema:
        ref_path = schema["$ref"].lstrip("#/").split("/")
        node = spec
        for key in ref_path:
            node = node[key]
        return _inline(spec, node)
    if "allOf" in schema:
        merged = {"type": "object", "properties": {}, "required": []}
        for sub in schema["allOf"]:
            sub_inlined = _inline(spec, sub)
            if "properties" in sub_inlined:
                merged["properties"].update(sub_inlined["properties"])
            if "required" in sub_inlined:
                merged["required"] = sorted(set(merged["required"]) | set(sub_inlined["required"]))
            if "type" in sub_inlined:
                merged["type"] = sub_inlined["type"]
        # 递归处理 properties 内部可能有的 $ref
        merged["properties"] = {k: _inline(spec, v) for k, v in merged["properties"].items()}
        return merged
    # 普通 object：递归 properties / items / additionalProperties 等
    result = dict(schema)
    for key in ("properties", "patternProperties"):
        if key in result and isinstance(result[key], dict):
            result[key] = {k: _inline(spec, v) for k, v in result[key].items()}
    for key in ("items", "additionalProperties"):
        if key in result:
            result[key] = _inline(spec, result[key])
    if "oneOf" in result:
        result["oneOf"] = [_inline(spec, s) for s in result["oneOf"]]
    if "anyOf" in result:
        result["anyOf"] = [_inline(spec, s) for s in result["anyOf"]]
    return result


def validate_payload(spec, schema, payload):
    """校验 payload 符合 schema。失败抛 AssertionError。"""
    validator = Draft7Validator(schema)
    errors = list(validator.iter_errors(payload))
    if errors:
        msgs = []
        for e in errors[:5]:
            path = "/".join(str(p) for p in e.absolute_path) or "<root>"
            msgs.append(f"  {path}: {e.message}")
        raise AssertionError(
            f"payload not conform to schema:\n" + "\n".join(msgs)
            + (f"\n  ... +{len(errors)-5} more" if len(errors) > 5 else "")
        )


@pytest.fixture
def schema_for(spec):
    return lambda path, method, status="200": resolve_response_schema(spec, path, method, status)


@pytest.fixture
def assert_contract(spec):
    def _check(payload, schema):
        validate_payload(spec, schema, payload)
    return _check