# 用户管理系统 API 文档

Base URL: `http://localhost:1323/api`

## 统一响应格式

```json
{
  "code": 200,
  "message": "success",
  "data": {}
}
```

## 1. 用户注册

`POST /api/register`

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| username | string | 是 | 用户名，3-50字符 |
| password | string | 是 | 密码，最少6位 |

---

## 2. 用户登录

`POST /api/login`

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| username | string | 是 | 用户名或邮箱 |
| password | string | 是 | 密码 |

Response:

```json
{ "code": 200, "data": { "token": "eyJhbGciOiJIUzI1NiIs..." } }
```

---

## 3. 发送验证码

`POST /api/send-code`

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| email | string | 是 | 邮箱地址 |

> 有频率限制，默认 1 分钟内不可重复发送。

---

## 4. 验证码登录

`POST /api/code-login`

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| email | string | 是 | 邮箱地址 |
| code | string | 是 | 验证码 |

Response: 同登录接口，返回 token。

---

## 5. 获取个人信息

`GET /api/profile`

Headers: `Authorization: Bearer <token>`

---

## 6. 修改个人信息

`PUT /api/profile`

Headers: `Authorization: Bearer <token>`

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| password | string | 否 | 新密码，最少6位 |

> 普通用户不能修改自己的角色。

---

## 7. 绑定邮箱

`PUT /api/profile/email`

Headers: `Authorization: Bearer <token>`

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| email | string | 是 | 邮箱地址 |
| code | string | 是 | 验证码 |

---

## 8. 获取用户列表（管理员）

`GET /api/users?page=1&page_size=20`

Headers: `Authorization: Bearer <token>`（需 admin 角色）

Query 参数:

| 参数 | 默认值 | 说明 |
|------|--------|------|
| page | 1 | 页码 |
| page_size | 20 | 每页数量（最大100）|

Response:

```json
{
  "code": 200,
  "data": {
    "list": [...],
    "total": 100,
    "page": 1,
    "page_size": 20
  }
}
```

---

## 9. 获取指定用户（管理员）

`GET /api/users/:id`

Headers: `Authorization: Bearer <token>`（需 admin 角色）

---

## 10. 更新指定用户（管理员）

`PUT /api/users/:id`

Headers: `Authorization: Bearer <token>`（需 admin 角色）

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| password | string | 否 | 新密码，最少6位 |
| role | string | 否 | `admin` 或 `user` |

---

## 11. 删除指定用户（管理员）

`DELETE /api/users/:id`

Headers: `Authorization: Bearer <token>`（需 admin 角色）

> 软删除，数据可恢复。

---

## 错误码

| code | 说明 |
|------|------|
| 200 | 成功 |
| 400 | 参数错误 / 业务错误 |
| 401 | 未认证 / token无效 |
| 403 | 权限不足 |
| 404 | 资源不存在 |
| 429 | 请求过于频繁 |
| 500 | 服务器内部错误 |
