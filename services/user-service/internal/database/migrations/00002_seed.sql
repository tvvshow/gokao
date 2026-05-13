-- +goose Up
-- +goose StatementBegin

-- 默认权限种子（idempotent，使用 ON CONFLICT 跳过既存项）。
-- 旧 seedDefaultData() 等价：28 条权限按资源域分组。
INSERT INTO permissions (name, description) VALUES
    ('user:read',          '查看用户信息'),
    ('user:write',         '修改用户信息'),
    ('user:delete',        '删除用户'),
    ('user:verify',        '验证用户身份'),
    ('role:read',          '查看角色信息'),
    ('role:write',         '修改角色信息'),
    ('role:delete',        '删除角色'),
    ('permission:manage',  '管理权限'),
    ('membership:read',    '查看会员信息'),
    ('membership:write',   '修改会员信息'),
    ('membership:upgrade', '会员升级'),
    ('membership:order',   '创建会员订单'),
    ('order:read',         '查看订单信息'),
    ('order:process',      '处理订单'),
    ('order:refund',       '订单退款'),
    ('device:read',        '查看设备信息'),
    ('device:manage',      '管理设备绑定'),
    ('device:trust',       '设置信任设备'),
    ('device:revoke',      '撤销设备授权'),
    ('session:read',       '查看会话信息'),
    ('session:manage',     '管理用户会话'),
    ('session:revoke',     '撤销会话'),
    ('audit:read',         '查看审计日志'),
    ('audit:export',       '导出审计日志'),
    ('system:monitor',     '系统监控'),
    ('system:stats',       '查看系统统计'),
    ('admin:all',          '管理员全部权限')
ON CONFLICT (name) DO NOTHING;

-- 默认角色种子。
INSERT INTO roles (name, description, is_system) VALUES
    ('admin',      '系统管理员',  true),
    ('user',       '普通用户',    true),
    ('basic',      '基础会员',    true),
    ('premium',    '高级会员',    true),
    ('enterprise', '企业会员',    true),
    ('moderator',  '内容审核员',  false)
ON CONFLICT (name) DO NOTHING;

-- admin 角色 → 全部权限（用 SELECT 派生避免硬编码 ID）。
-- 旧实现按行 First/Create，新实现单条 INSERT...SELECT + ON CONFLICT 完成。
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
CROSS JOIN permissions p
WHERE r.name = 'admin'
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- user 角色 → 基础只读权限
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.name = 'user'
  AND p.name IN ('user:read', 'device:read', 'session:read')
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- basic 角色 → 自助管理 + 下单
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.name = 'basic'
  AND p.name IN (
      'user:read', 'user:write',
      'device:read', 'device:manage',
      'session:read', 'session:manage',
      'membership:read', 'membership:order',
      'order:read'
  )
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- premium 角色 → 基础 + 信任设备 + 审计读
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.name = 'premium'
  AND p.name IN (
      'user:read', 'user:write',
      'device:read', 'device:manage', 'device:trust',
      'session:read', 'session:manage',
      'membership:read', 'membership:write', 'membership:order', 'membership:upgrade',
      'order:read',
      'audit:read'
  )
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- enterprise 角色 → premium + 撤销 + 订单处理 + 审计导出 + 统计
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.name = 'enterprise'
  AND p.name IN (
      'user:read', 'user:write', 'user:verify',
      'device:read', 'device:manage', 'device:trust', 'device:revoke',
      'session:read', 'session:manage', 'session:revoke',
      'membership:read', 'membership:write', 'membership:order', 'membership:upgrade',
      'order:read', 'order:process',
      'audit:read', 'audit:export',
      'system:stats'
  )
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- moderator 角色 → 审核类只读 + 系统监控
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.name = 'moderator'
  AND p.name IN (
      'user:read', 'user:verify',
      'audit:read',
      'system:monitor'
  )
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- 回滚顺序：先清依赖（role_permissions），再清主表种子行。
-- 注：DELETE 限定 system seeds + 6 角色名，避免误删手工新增数据。
DELETE FROM role_permissions
WHERE role_id IN (SELECT id FROM roles WHERE name IN ('admin','user','basic','premium','enterprise','moderator'));

DELETE FROM roles
WHERE name IN ('admin','user','basic','premium','enterprise','moderator')
  AND is_system = true;

DELETE FROM permissions WHERE name IN (
    'user:read','user:write','user:delete','user:verify',
    'role:read','role:write','role:delete','permission:manage',
    'membership:read','membership:write','membership:upgrade','membership:order',
    'order:read','order:process','order:refund',
    'device:read','device:manage','device:trust','device:revoke',
    'session:read','session:manage','session:revoke',
    'audit:read','audit:export',
    'system:monitor','system:stats',
    'admin:all'
);

-- +goose StatementEnd
