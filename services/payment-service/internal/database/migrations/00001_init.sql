-- +goose Up
-- +goose StatementBegin

-- 支付订单
CREATE TABLE IF NOT EXISTS payment_orders (
    id SERIAL PRIMARY KEY,
    order_no VARCHAR(64) UNIQUE NOT NULL,
    user_id VARCHAR(64) NOT NULL,
    amount DECIMAL(10,2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'CNY',
    subject VARCHAR(256) NOT NULL,
    description TEXT,
    status VARCHAR(32) DEFAULT 'pending',
    payment_channel VARCHAR(32),
    channel_trade_no VARCHAR(128),
    client_ip INET,
    notify_url TEXT,
    return_url TEXT,
    expire_time TIMESTAMP,
    paid_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    metadata JSONB
);

-- 退款记录
CREATE TABLE IF NOT EXISTS refund_records (
    id SERIAL PRIMARY KEY,
    refund_no VARCHAR(64) UNIQUE NOT NULL,
    order_no VARCHAR(64) NOT NULL,
    channel_refund_no VARCHAR(128),
    amount DECIMAL(10,2) NOT NULL,
    reason VARCHAR(512),
    status VARCHAR(32) DEFAULT 'pending',
    refunded_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (order_no) REFERENCES payment_orders(order_no)
);

-- 会员套餐
CREATE TABLE IF NOT EXISTS membership_plans (
    id SERIAL PRIMARY KEY,
    plan_code VARCHAR(32) UNIQUE NOT NULL,
    name VARCHAR(128) NOT NULL,
    description TEXT,
    price DECIMAL(10,2) NOT NULL,
    duration_days INTEGER NOT NULL,
    features JSONB,
    max_queries INTEGER DEFAULT 0,
    max_downloads INTEGER DEFAULT 0,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 用户会员
CREATE TABLE IF NOT EXISTS user_memberships (
    id SERIAL PRIMARY KEY,
    user_id VARCHAR(64) NOT NULL,
    plan_code VARCHAR(32) NOT NULL,
    order_no VARCHAR(64) NOT NULL,
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP NOT NULL,
    status VARCHAR(32) DEFAULT 'active',
    auto_renew BOOLEAN DEFAULT false,
    used_queries INTEGER DEFAULT 0,
    used_downloads INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (plan_code) REFERENCES membership_plans(plan_code),
    FOREIGN KEY (order_no) REFERENCES payment_orders(order_no)
);

-- 支付回调日志
CREATE TABLE IF NOT EXISTS payment_callbacks (
    id SERIAL PRIMARY KEY,
    order_no VARCHAR(64) NOT NULL,
    channel VARCHAR(32) NOT NULL,
    callback_data TEXT NOT NULL,
    signature VARCHAR(512),
    verified BOOLEAN DEFAULT false,
    processed BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 许可证信息
CREATE TABLE IF NOT EXISTS license_info (
    id SERIAL PRIMARY KEY,
    user_id VARCHAR(64) NOT NULL,
    device_id VARCHAR(128) NOT NULL,
    device_fingerprint TEXT NOT NULL,
    license_key VARCHAR(512) NOT NULL,
    encrypted_data TEXT NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    status VARCHAR(32) DEFAULT 'active',
    bind_count INTEGER DEFAULT 1,
    max_bind_count INTEGER DEFAULT 3,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, device_id)
);

-- 索引
CREATE INDEX IF NOT EXISTS idx_payment_orders_user_id    ON payment_orders(user_id);
CREATE INDEX IF NOT EXISTS idx_payment_orders_status     ON payment_orders(status);
CREATE INDEX IF NOT EXISTS idx_payment_orders_created_at ON payment_orders(created_at);
CREATE INDEX IF NOT EXISTS idx_refund_records_order_no   ON refund_records(order_no);
CREATE INDEX IF NOT EXISTS idx_user_memberships_user_id  ON user_memberships(user_id);
CREATE INDEX IF NOT EXISTS idx_user_memberships_status   ON user_memberships(status);
CREATE INDEX IF NOT EXISTS idx_license_info_user_id      ON license_info(user_id);
CREATE INDEX IF NOT EXISTS idx_license_info_device_id    ON license_info(device_id);

-- 默认会员套餐 seed
-- 注：seed 嵌入 baseline migration 是因为现有 createTables() 行为如此，保持兼容；
-- 后续若要把 seed 拆出可新增一条 data migration 单独处理。
INSERT INTO membership_plans (plan_code, name, description, price, duration_days, features, max_queries, max_downloads)
VALUES
    ('basic',    '基础版', '基础功能套餐，适合初级用户',                29.90,  30,  '{"basic_query": true, "data_export": false, "ai_recommendation": false}'::jsonb, 100,  0),
    ('premium',  '高级版', '高级功能套餐，包含AI推荐',                  99.90,  90,  '{"basic_query": true, "data_export": true, "ai_recommendation": true, "priority_support": true}'::jsonb, 1000, 50),
    ('ultimate', '旗舰版', '全功能套餐，无限制使用',                    299.90, 365, '{"basic_query": true, "data_export": true, "ai_recommendation": true, "priority_support": true, "unlimited": true}'::jsonb, -1,   -1)
ON CONFLICT (plan_code) DO NOTHING;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS license_info       CASCADE;
DROP TABLE IF EXISTS payment_callbacks  CASCADE;
DROP TABLE IF EXISTS user_memberships   CASCADE;
DROP TABLE IF EXISTS membership_plans   CASCADE;
DROP TABLE IF EXISTS refund_records     CASCADE;
DROP TABLE IF EXISTS payment_orders     CASCADE;

-- +goose StatementEnd
