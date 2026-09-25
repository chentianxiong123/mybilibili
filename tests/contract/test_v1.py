"""契约测试：打真后端，校验响应符合 contracts/openapi/v1.yaml。

跑法：
  pytest tests/contract/
  API_BASE_URL=http://staging/api/v1 pytest tests/contract/

所有测试独立（不依赖顺序），共享：
  - spec_is_valid    契约文件本身合法（session 级）
  - login_token      登录 string/123456 拿 JWT（session 级）
"""
import pytest
import requests


# ---------- 契约文件本身 ----------
def test_openapi_spec_is_valid(spec_is_valid):
    """契约文件本身必须符合 OpenAPI 3.1。"""
    assert spec_is_valid


# ---------- user/login ----------
class TestUserLogin:
    def test_login_success(self, http, base_url, spec, schema_for, assert_contract):
        resp = http.post(f"{base_url}/user/login",
                         json={"username": "string", "password": "123456"})
        assert resp.status_code == 200
        body = resp.json()
        schema = schema_for("/user/login", "post")
        assert_contract(body, schema)

    def test_login_wrong_password(self, http, base_url):
        resp = http.post(f"{base_url}/user/login",
                         json={"username": "string", "password": "wrong"})
        assert resp.status_code == 401


# ---------- user/me ----------
class TestUserMe:
    def test_me_with_token(self, http, base_url, auth_headers, schema_for, assert_contract):
        resp = http.get(f"{base_url}/user/me", headers=auth_headers)
        assert resp.status_code == 200
        body = resp.json()
        schema = schema_for("/user/me", "get")
        assert_contract(body, schema)

    def test_me_without_token(self, http, base_url):
        resp = http.get(f"{base_url}/user/me")
        assert resp.status_code == 401


# ---------- user/{id} ----------
class TestUserByID:
    def test_user_id_4_exists(self, http, base_url, schema_for, assert_contract):
        resp = http.get(f"{base_url}/user/4")
        assert resp.status_code == 200
        body = resp.json()
        schema = schema_for("/user/{id}", "get")
        assert_contract(body, schema)

    def test_user_id_999_not_found(self, http, base_url):
        resp = http.get(f"{base_url}/user/999")
        assert resp.status_code == 404


# ---------- manuscript/recommended ----------
class TestManuscriptRecommended:
    def test_returns_list(self, http, base_url, schema_for, assert_contract):
        resp = http.get(f"{base_url}/manuscript/recommended")
        assert resp.status_code == 200
        body = resp.json()
        assert body["code"] == 200
        schema = schema_for("/manuscript/recommended", "get")
        assert_contract(body, schema)
        assert isinstance(body["data"], list)

    def test_refresh_time_changes_order(self, http, base_url):
        """不同 refresh_time 应返回不同顺序的稿件（推荐流轮播行为）。"""
        r1 = http.get(f"{base_url}/manuscript/recommended?refresh_time=0").json()["data"]
        r2 = http.get(f"{base_url}/manuscript/recommended?refresh_time=1").json()["data"]
        # 列表顺序（至少前几个 id）应该不同
        ids1 = [m["id"] for m in r1[:5]]
        ids2 = [m["id"] for m in r2[:5]]
        assert ids1 != ids2 or len(r1) < 5, f"order unchanged: {ids1}"


# ---------- manuscript/hot ----------
class TestManuscriptHot:
    def test_returns_list(self, http, base_url, schema_for, assert_contract):
        resp = http.get(f"{base_url}/manuscript/hot?page_size=5")
        assert resp.status_code == 200
        body = resp.json()
        schema = schema_for("/manuscript/hot", "get")
        assert_contract(body, schema)

    def test_pagination_accepts_offset(self, http, base_url, schema_for, assert_contract):
        """offset 参数应被接受（schema 声明），业务分页由后端实现。"""
        resp = http.get(f"{base_url}/manuscript/hot?offset=0")
        assert resp.status_code == 200
        schema = schema_for("/manuscript/hot", "get")
        assert_contract(resp.json(), schema)


# ---------- manuscript/detail/{id} ----------
class TestManuscriptDetail:
    def test_detail_10(self, http, base_url, schema_for, assert_contract):
        """已修过的 bug：detail 路由原本 return 404，现在必须能取到详情。"""
        resp = http.get(f"{base_url}/manuscript/detail/10")
        assert resp.status_code == 200
        body = resp.json()
        schema = schema_for("/manuscript/detail/{id}", "get")
        assert_contract(body, schema)
        assert body["data"]["id"] == 10

    def test_detail_video_id_fallback(self, http, base_url, schema_for, assert_contract):
        """用视频 id（firstVideoId=25）兜底取稿件 10。"""
        resp = http.get(f"{base_url}/manuscript/detail/25")
        assert resp.status_code == 200
        body = resp.json()
        schema = schema_for("/manuscript/detail/{id}", "get")
        assert_contract(body, schema)

    def test_detail_not_found(self, http, base_url):
        resp = http.get(f"{base_url}/manuscript/detail/999999")
        assert resp.status_code == 404


# ---------- comment/list ----------
class TestCommentList:
    def test_comment_list_manuscript_10(self, http, base_url, schema_for, assert_contract):
        resp = http.get(f"{base_url}/comment/list?manuscriptId=10&page_size=5")
        assert resp.status_code == 200
        body = resp.json()
        schema = schema_for("/comment/list", "get")
        assert_contract(body, schema)
        if body["data"]:
            # manuscriptId 实际为 string
            assert all(c["manuscriptId"] == "10" for c in body["data"])

    def test_comment_list_missing_manuscriptId(self, http, base_url):
        resp = http.get(f"{base_url}/comment/list")
        # 缺 manuscriptId 应返回 400 或空 list（契约不强制 400）
        assert resp.status_code in (200, 400)


# ---------- comment/get-up-like ----------
class TestCommentUpLike:
    def test_get_up_like_uid_4(self, http, base_url, schema_for, assert_contract):
        resp = http.get(f"{base_url}/comment/get-up-like?uid=4")
        assert resp.status_code == 200
        body = resp.json()
        schema = schema_for("/comment/get-up-like", "get")
        assert_contract(body, schema)
        assert isinstance(body["data"], list)
        # 每个元素应该是整数
        assert all(isinstance(x, int) for x in body["data"])


# ---------- favorites/check ----------
class TestFavoriteCheck:
    def test_check_with_token(self, http, base_url, auth_headers, schema_for, assert_contract):
        resp = http.get(f"{base_url}/favorites/check?manuscript_id=10", headers=auth_headers)
        assert resp.status_code == 200
        body = resp.json()
        schema = schema_for("/favorites/check", "get")
        assert_contract(body, schema)
        assert isinstance(body["data"]["favorited"], bool)

    def test_check_without_token(self, http, base_url):
        resp = http.get(f"{base_url}/favorites/check?manuscript_id=10")
        assert resp.status_code == 401

    def test_check_param_alias(self, http, base_url, auth_headers):
        """manuscriptId / vid 三种参数名都应工作（兼容层契约）。"""
        for param in ["manuscript_id=10", "manuscriptId=10", "vid=25"]:
            resp = http.get(f"{base_url}/favorites/check?{param}", headers=auth_headers)
            assert resp.status_code == 200, f"param '{param}' failed: {resp.status_code}"


# ---------- favorites/list ----------
class TestFavoriteList:
    def test_list_with_token(self, http, base_url, auth_headers, schema_for, assert_contract):
        resp = http.get(f"{base_url}/favorites/list?page_size=5", headers=auth_headers)
        assert resp.status_code == 200
        body = resp.json()
        schema = schema_for("/favorites/list", "get")
        assert_contract(body, schema)

    def test_list_without_token(self, http, base_url):
        resp = http.get(f"{base_url}/favorites/list")
        assert resp.status_code == 401


# ---------- favorites/manuscript/{id} ----------
class TestFavoriteManuscriptFolders:
    def test_manuscript_10(self, http, base_url, schema_for, assert_contract):
        resp = http.get(f"{base_url}/favorites/manuscript/10")
        assert resp.status_code == 200
        body = resp.json()
        schema = schema_for("/favorites/manuscript/{id}", "get")
        assert_contract(body, schema)
        assert isinstance(body["data"], list)


# ---------- 健康度 ----------
class TestContractHealth:
    def test_all_paths_have_response_schema(self, spec):
        """每个 path+method 都必须声明 response schema。"""
        for path, item in spec["paths"].items():
            for method in item:
                if method in ("get", "post", "put", "delete", "patch"):
                    responses = item[method].get("responses", {})
                    has_schema = False
                    for status, resp in responses.items():
                        if "content" in resp and "schema" in resp["content"].get("application/json", {}):
                            has_schema = True
                            break
                    assert has_schema, f"{method.upper()} {path} 缺 response schema"