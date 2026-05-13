-- +goose Up
-- +goose StatementBegin

-- 扩展
-- pgcrypto: gen_random_uuid() ——所有 uuid PK 使用
-- pg_trgm:  保留接入点；当前 user-service 无 LIKE 搜索路径，未启用 trgm GIN 索引
CREATE EXTENSION IF NOT EXISTS pgcrypto;
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- 用户
CREATE TABLE IF NOT EXISTS users (
    id                 uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    username           varchar(50)  NOT NULL,
    email              varchar(100) NOT NULL,
    phone              varchar(20),
    password           varchar(255) NOT NULL,
    nickname           varchar(50),
    avatar             varchar(255),
    gender             varchar(10),
    birthday           timestamp with time zone,
    province           varchar(50),
    city               varchar(50),
    school             varchar(100),
    grade              varchar(20),
    status             varchar(20)  DEFAULT 'active',
    is_verified        boolean      DEFAULT false,
    membership_level   varchar(20)  DEFAULT 'free',
    membership_expiry  timestamp with time zone,
    max_devices        bigint       DEFAULT 1,
    trial_used         boolean      DEFAULT false,
    trial_expiry       timestamp with time zone,
    last_login_at      timestamp with time zone,
    last_login_ip      varchar(45),
    login_count        bigint       DEFAULT 0,
    created_at         timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at         timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    deleted_at         timestamp with time zone,
    CONSTRAINT users_username_key UNIQUE (username),
    CONSTRAINT users_email_key    UNIQUE (email),
    CONSTRAINT users_phone_key    UNIQUE (phone)
);

-- 角色
CREATE TABLE IF NOT EXISTS roles (
    id          bigserial PRIMARY KEY,
    name        varchar(50)  NOT NULL,
    description varchar(255),
    is_system   boolean      DEFAULT false,
    created_at  timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at  timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT roles_name_key UNIQUE (name)
);

-- 权限
CREATE TABLE IF NOT EXISTS permissions (
    id          bigserial PRIMARY KEY,
    name        varchar(100) NOT NULL,
    description varchar(255),
    resource    varchar(50),
    action      varchar(50),
    created_at  timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at  timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT permissions_name_key UNIQUE (name)
);

-- 用户-角色多对多
CREATE TABLE IF NOT EXISTS user_roles (
    user_id    uuid   NOT NULL,
    role_id    bigint NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, role_id)
);

-- 角色-权限多对多
CREATE TABLE IF NOT EXISTS role_permissions (
    role_id       bigint NOT NULL,
    permission_id bigint NOT NULL,
    created_at    timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (role_id, permission_id)
);

-- 登录尝试记录
CREATE TABLE IF NOT EXISTS login_attempts (
    id         bigserial PRIMARY KEY,
    username   varchar(50)  NOT NULL,
    ip         varchar(45)  NOT NULL,
    user_agent varchar(500),
    success    boolean,
    reason     varchar(255),
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);

-- 审计日志
CREATE TABLE IF NOT EXISTS audit_logs (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     uuid,
    action      varchar(100) NOT NULL,
    resource    varchar(100),
    resource_id varchar(100),
    details     text,
    ip          varchar(45),
    user_agent  varchar(500),
    status      varchar(20),
    created_at  timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);

-- 刷新令牌
CREATE TABLE IF NOT EXISTS refresh_tokens (
    id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    uuid         NOT NULL,
    token      varchar(255) NOT NULL,
    expires_at timestamp with time zone NOT NULL,
    is_revoked boolean      DEFAULT false,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT refresh_tokens_token_key UNIQUE (token)
);

-- 设备指纹
CREATE TABLE IF NOT EXISTS device_fingerprints (
    id                uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id           uuid         NOT NULL,
    device_id         varchar(255) NOT NULL,
    device_name       varchar(100),
    device_type       varchar(20),
    platform          varchar(50),
    browser           varchar(100),
    browser_version   varchar(50),
    os                varchar(50),
    os_version        varchar(50),
    screen_resolution varchar(20),
    timezone          varchar(50),
    language          varchar(10),
    user_agent        varchar(500),
    ip_address        varchar(45),
    location          varchar(100),
    is_active         boolean      DEFAULT true,
    is_trusted        boolean      DEFAULT false,
    last_seen_at      timestamp with time zone,
    created_at        timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at        timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    deleted_at        timestamp with time zone,
    CONSTRAINT device_fingerprints_device_id_key UNIQUE (device_id)
);

-- 设备许可证
CREATE TABLE IF NOT EXISTS device_licenses (
    id            uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id       uuid         NOT NULL,
    device_id     varchar(255) NOT NULL,
    license_data  text         NOT NULL,
    status        varchar(20)  NOT NULL,
    issued_at     timestamp with time zone NOT NULL,
    expires_at    timestamp with time zone,
    revoked_at    timestamp with time zone,
    revoke_reason varchar(255),
    created_at    timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at    timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    deleted_at    timestamp with time zone
);

-- 会员订单
CREATE TABLE IF NOT EXISTS membership_orders (
    id               uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id          uuid         NOT NULL,
    order_no         varchar(50)  NOT NULL,
    product_name     varchar(100) NOT NULL,
    membership_level varchar(20)  NOT NULL,
    duration         bigint       NOT NULL,
    original_price   bigint       NOT NULL,
    discount_price   bigint       DEFAULT 0,
    final_price      bigint       NOT NULL,
    currency         varchar(10)  DEFAULT 'CNY',
    payment_method   varchar(50),
    payment_provider varchar(50),
    payment_id       varchar(100),
    discount_code    varchar(50),
    status           varchar(20)  NOT NULL,
    paid_at          timestamp with time zone,
    expired_at       timestamp with time zone,
    refunded_at      timestamp with time zone,
    refund_amount    bigint       DEFAULT 0,
    refund_reason    varchar(255),
    notes            text,
    created_at       timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at       timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    deleted_at       timestamp with time zone,
    CONSTRAINT membership_orders_order_no_key UNIQUE (order_no)
);

-- 用户会话
CREATE TABLE IF NOT EXISTS user_sessions (
    id                  uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id             uuid         NOT NULL,
    device_id           varchar(255),
    session_token       varchar(255) NOT NULL,
    refresh_token       varchar(255),
    ip_address          varchar(45),
    user_agent          varchar(500),
    location            varchar(100),
    is_active           boolean      DEFAULT true,
    expires_at          timestamp with time zone NOT NULL,
    refresh_expires_at  timestamp with time zone,
    last_activity_at    timestamp with time zone NOT NULL,
    created_at          timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at          timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    deleted_at          timestamp with time zone,
    CONSTRAINT user_sessions_session_token_key UNIQUE (session_token),
    CONSTRAINT user_sessions_refresh_token_key UNIQUE (refresh_token)
);

-- 索引（GORM `gorm:"index"` tag 列）
-- users
CREATE INDEX IF NOT EXISTS idx_users_province            ON users(province);
CREATE INDEX IF NOT EXISTS idx_users_status              ON users(status);
CREATE INDEX IF NOT EXISTS idx_users_is_verified         ON users(is_verified);
CREATE INDEX IF NOT EXISTS idx_users_membership_level    ON users(membership_level);
CREATE INDEX IF NOT EXISTS idx_users_membership_expiry   ON users(membership_expiry);
CREATE INDEX IF NOT EXISTS idx_users_deleted_at          ON users(deleted_at);

-- login_attempts
CREATE INDEX IF NOT EXISTS idx_login_attempts_username   ON login_attempts(username);
CREATE INDEX IF NOT EXISTS idx_login_attempts_ip         ON login_attempts(ip);
CREATE INDEX IF NOT EXISTS idx_login_attempts_success    ON login_attempts(success);

-- audit_logs
CREATE INDEX IF NOT EXISTS idx_audit_logs_user_id        ON audit_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_action         ON audit_logs(action);
CREATE INDEX IF NOT EXISTS idx_audit_logs_resource       ON audit_logs(resource);
CREATE INDEX IF NOT EXISTS idx_audit_logs_resource_id    ON audit_logs(resource_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_ip             ON audit_logs(ip);
CREATE INDEX IF NOT EXISTS idx_audit_logs_status         ON audit_logs(status);

-- refresh_tokens
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_id    ON refresh_tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_expires_at ON refresh_tokens(expires_at);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_is_revoked ON refresh_tokens(is_revoked);

-- device_fingerprints
CREATE INDEX IF NOT EXISTS idx_device_fingerprints_user_id      ON device_fingerprints(user_id);
CREATE INDEX IF NOT EXISTS idx_device_fingerprints_device_type  ON device_fingerprints(device_type);
CREATE INDEX IF NOT EXISTS idx_device_fingerprints_ip_address   ON device_fingerprints(ip_address);
CREATE INDEX IF NOT EXISTS idx_device_fingerprints_is_active    ON device_fingerprints(is_active);
CREATE INDEX IF NOT EXISTS idx_device_fingerprints_last_seen_at ON device_fingerprints(last_seen_at);
CREATE INDEX IF NOT EXISTS idx_device_fingerprints_deleted_at   ON device_fingerprints(deleted_at);

-- device_licenses
CREATE INDEX IF NOT EXISTS idx_device_licenses_user_id    ON device_licenses(user_id);
CREATE INDEX IF NOT EXISTS idx_device_licenses_device_id  ON device_licenses(device_id);
CREATE INDEX IF NOT EXISTS idx_device_licenses_status     ON device_licenses(status);
CREATE INDEX IF NOT EXISTS idx_device_licenses_expires_at ON device_licenses(expires_at);
CREATE INDEX IF NOT EXISTS idx_device_licenses_deleted_at ON device_licenses(deleted_at);

-- membership_orders
CREATE INDEX IF NOT EXISTS idx_membership_orders_user_id          ON membership_orders(user_id);
CREATE INDEX IF NOT EXISTS idx_membership_orders_membership_level ON membership_orders(membership_level);
CREATE INDEX IF NOT EXISTS idx_membership_orders_payment_method   ON membership_orders(payment_method);
CREATE INDEX IF NOT EXISTS idx_membership_orders_payment_id       ON membership_orders(payment_id);
CREATE INDEX IF NOT EXISTS idx_membership_orders_status           ON membership_orders(status);
CREATE INDEX IF NOT EXISTS idx_membership_orders_paid_at          ON membership_orders(paid_at);
CREATE INDEX IF NOT EXISTS idx_membership_orders_expired_at       ON membership_orders(expired_at);
CREATE INDEX IF NOT EXISTS idx_membership_orders_deleted_at       ON membership_orders(deleted_at);

-- user_sessions
CREATE INDEX IF NOT EXISTS idx_user_sessions_user_id            ON user_sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_user_sessions_device_id          ON user_sessions(device_id);
CREATE INDEX IF NOT EXISTS idx_user_sessions_ip_address         ON user_sessions(ip_address);
CREATE INDEX IF NOT EXISTS idx_user_sessions_is_active          ON user_sessions(is_active);
CREATE INDEX IF NOT EXISTS idx_user_sessions_expires_at         ON user_sessions(expires_at);
CREATE INDEX IF NOT EXISTS idx_user_sessions_refresh_expires_at ON user_sessions(refresh_expires_at);
CREATE INDEX IF NOT EXISTS idx_user_sessions_last_activity_at   ON user_sessions(last_activity_at);
CREATE INDEX IF NOT EXISTS idx_user_sessions_deleted_at         ON user_sessions(deleted_at);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS user_sessions       CASCADE;
DROP TABLE IF EXISTS membership_orders   CASCADE;
DROP TABLE IF EXISTS device_licenses     CASCADE;
DROP TABLE IF EXISTS device_fingerprints CASCADE;
DROP TABLE IF EXISTS refresh_tokens      CASCADE;
DROP TABLE IF EXISTS audit_logs          CASCADE;
DROP TABLE IF EXISTS login_attempts      CASCADE;
DROP TABLE IF EXISTS role_permissions    CASCADE;
DROP TABLE IF EXISTS user_roles          CASCADE;
DROP TABLE IF EXISTS permissions         CASCADE;
DROP TABLE IF EXISTS roles               CASCADE;
DROP TABLE IF EXISTS users               CASCADE;

-- +goose StatementEnd
