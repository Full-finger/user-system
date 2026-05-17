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

**Request Body (JSON):**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| username | string | 是 | 用户名，3-50字符 |
| password | string | 是 | 密码，最少6位 |

**Response:** 返回创建的用户信息（不含密码）

---

## 2. 用户登录

`POST /api/login`

**Request Body (JSON):**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| username | string | 是 | 用户名 |
| password | string | 是 | 密码 |

**Response:**

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIs..."
  }
}
```

---

## 3. 获取个人信息

`GET /api/profile`

**Headers:** `Authorization: Bearer <token>`

**Response:** 返回当前登录用户信息

---

## 4. 修改个人信息

`PUT /api/profile`

**Headers:** `Authorization: Bearer <token>`

**Request Body (JSON):**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| password | string | 否 | 新密码，最少6位 |

> 注意：普通用户不能修改自己的角色

---

## 5. 获取用户列表（管理员）

`GET /api/users`

**Headers:** `Authorization: Bearer <token>`（需 admin 角色）

**Response:** 返回所有用户列表

---

## 6. 获取指定用户（管理员）

`GET /api/users/:id`

**Headers:** `Authorization: Bearer <token>`（需 admin 角色）

**Path 参数:** `id` - 用户ID

---

## 7. 更新指定用户（管理员）

`PUT /api/users/:id`

**Headers:** `Authorization: Bearer <token>`（需 admin 角色）

**Path 参数:** `id` - 用户ID

**Request Body (JSON):**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| password | string | 否 | 新密码，最少6位 |
| role | string | 否 | 角色，`admin` 或 `user` |

---

## 8. 删除指定用户（管理员）

`DELETE /api/users/:id`

**Headers:** `Authorization: Bearer <token>`（需 admin 角色）

**Path 参数:** `id` - 用户ID

**Response:** data 为 null

---

## 错误码说明

| code | 说明 |
|------|------|
| 200 | 成功 |
| 400 | 参数错误 / 业务错误 |
| 401 | 未认证 / token无效 |
| 403 | 权限不足 |
| 404 | 资源不存在 |
| 500 | 服务器内部错误 |