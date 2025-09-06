# 高考志愿填报系统 - 设备认证服务API文档

## 1. 概述

设备认证服务提供设备指纹采集、验证和许可证管理功能。所有API请求和响应均使用JSON格式。

### 1.1. 基础URL

```
http://device-auth-service:8085/api/v1
```

### 1.2. 状态码

| 状态码 | 说明             |
| ------ | ---------------- |
| 200    | 请求成功         |
| 400    | 请求参数错误     |
| 401    | 未授权           |
| 500    | 服务器内部错误   |

## 2. 设备指纹相关接口

### 2.1. 采集设备指纹

#### 请求

```
POST /device/fingerprint
```

#### 响应

```json
{
  "device_id": "string",
  "device_type": "string",
  "cpu_id": "string",
  "cpu_model": "string",
  "cpu_cores": 0,
  "total_memory": 0,
  "os_type": "string",
  "os_version": "string",
  "hostname": "string",
  "username": "string",
  "screen_resolution": "string",
  "fingerprint_hash": "string",
  "confidence_score": 0,
  "collected_at": "2025-01-01T00:00:00Z"
}
```

### 2.2. 注册设备

#### 请求

```
POST /device/register
```

#### 响应

```json
{
  "device_id": "string"
}
```

## 3. 许可证相关接口

### 3.1. 验证许可证

#### 请求

```
POST /license/validate
```

**请求体:**

```json
{
  "license_data": "string",
  "device_id": "string"
}
```

#### 响应

```json
{
  "device_id": "string",
  "license_type": "string",
  "issued_at": "2025-01-01T00:00:00Z",
  "expires_at": "2025-01-01T00:00:00Z",
  "max_devices": 0,
  "features": ["string"],
  "is_valid": true,
  "error_message": "string"
}
```

### 3.2. 生成许可证

#### 请求

```
POST /license/generate
```

**请求体:**

```json
{
  "device_id": "string",
  "expires_at": "2025-01-01T00:00:00Z",
  "private_key": "string",
  "license_type": "string",
  "features": ["string"]
}
```

#### 响应

```json
{
  "license_data": "string"
}
```