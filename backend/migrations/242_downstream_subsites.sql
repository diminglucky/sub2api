-- 下游子站（downstream subsite）基础表与归属字段。
--
-- 主站是唯一控制面：用户、余额、充值订单、API Key、模型、上游账号和网关都
-- 只有一套。子站只是带品牌的独立域名入口（V1 试点：draw.superai.sbs）。
--
-- 归属约定：subsite_id = NULL 表示主站数据。新增表为空表，旧数据无需迁移。

-- 1. subsites 子站表
CREATE TABLE IF NOT EXISTS subsites (
    id            BIGSERIAL PRIMARY KEY,
    slug          VARCHAR(64)  NOT NULL UNIQUE,
    domain        VARCHAR(255) NOT NULL UNIQUE,
    name          VARCHAR(100) NOT NULL,
    logo_url      TEXT         NOT NULL DEFAULT '',
    theme_color   VARCHAR(32)  NOT NULL DEFAULT '',
    status        VARCHAR(20)  NOT NULL DEFAULT 'active', -- active/disabled
    admin_user_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_subsites_status ON subsites(status);
CREATE INDEX IF NOT EXISTS idx_subsites_admin_user_id ON subsites(admin_user_id);

COMMENT ON TABLE subsites IS '下游子站：独立品牌域名，V1 仅支持 <slug>.superai.sbs';
COMMENT ON COLUMN subsites.slug IS '子域名标签，如 draw 对应 draw.superai.sbs';
COMMENT ON COLUMN subsites.domain IS '子站完整域名，如 draw.superai.sbs';
COMMENT ON COLUMN subsites.admin_user_id IS '子站管理员的主站用户 ID，可为空';

-- 2. subsite_members 用户与子站归属关系表
CREATE TABLE IF NOT EXISTS subsite_members (
    id         BIGSERIAL PRIMARY KEY,
    subsite_id BIGINT      NOT NULL REFERENCES subsites(id) ON DELETE CASCADE,
    user_id    BIGINT      NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role       VARCHAR(20) NOT NULL DEFAULT 'member', -- owner/admin/member
    source     VARCHAR(20) NOT NULL DEFAULT 'registration', -- registration/manual
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (subsite_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_subsite_members_subsite_id ON subsite_members(subsite_id);
CREATE INDEX IF NOT EXISTS idx_subsite_members_user_id ON subsite_members(user_id);

COMMENT ON TABLE subsite_members IS '主站用户与子站的归属关系；同一用户可归属多个子站';
COMMENT ON COLUMN subsite_members.role IS 'owner/admin 可管理子站，member 为普通用户';

-- 3. subsite_prices 子站价格覆盖表
CREATE TABLE IF NOT EXISTS subsite_prices (
    id                BIGSERIAL PRIMARY KEY,
    subsite_id        BIGINT        NOT NULL REFERENCES subsites(id) ON DELETE CASCADE,
    scope             VARCHAR(20)   NOT NULL, -- group/model
    group_id          BIGINT        REFERENCES groups(id) ON DELETE CASCADE,
    model             VARCHAR(100),
    rate_multiplier   DECIMAL(10,4) NOT NULL DEFAULT 1.0,
    input_price       DECIMAL(20,12),
    output_price      DECIMAL(20,12),
    cache_write_price DECIMAL(20,12),
    cache_read_price  DECIMAL(20,12),
    per_request_price DECIMAL(20,12),
    status            VARCHAR(20)   NOT NULL DEFAULT 'active', -- active/disabled
    created_at        TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    CHECK (
        (scope = 'group' AND group_id IS NOT NULL AND model IS NULL) OR
        (scope = 'model' AND model IS NOT NULL AND group_id IS NULL)
    )
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_subsite_prices_group
    ON subsite_prices (subsite_id, group_id)
    WHERE scope = 'group';
CREATE UNIQUE INDEX IF NOT EXISTS idx_subsite_prices_model
    ON subsite_prices (subsite_id, model)
    WHERE scope = 'model';
CREATE INDEX IF NOT EXISTS idx_subsite_prices_subsite_id ON subsite_prices(subsite_id);

COMMENT ON TABLE subsite_prices IS '子站价格覆盖；只影响该子站展示与结算，不修改上游模型配置';
COMMENT ON COLUMN subsite_prices.rate_multiplier IS '在主站价格基础上的加价倍率；显式价格为空时使用';
COMMENT ON COLUMN subsite_prices.scope IS 'group 按分组覆盖，model 按模型名覆盖';

-- 4. settlement_ledger 结算台账表
-- V1 只记录应付台账，不自动打款。
CREATE TABLE IF NOT EXISTS settlement_ledger (
    id            BIGSERIAL PRIMARY KEY,
    subsite_id    BIGINT        NOT NULL REFERENCES subsites(id) ON DELETE CASCADE,
    order_id      BIGINT REFERENCES payment_orders(id) ON DELETE SET NULL,
    user_id       BIGINT REFERENCES users(id) ON DELETE SET NULL,
    currency      VARCHAR(10)   NOT NULL DEFAULT 'USD',
    gross_amount  DECIMAL(20,8) NOT NULL DEFAULT 0,
    cost_amount   DECIMAL(20,8) NOT NULL DEFAULT 0,
    margin_amount DECIMAL(20,8) NOT NULL DEFAULT 0,
    status        VARCHAR(20)   NOT NULL DEFAULT 'pending', -- pending/settled/void
    notes         TEXT          NOT NULL DEFAULT '',
    settled_at    TIMESTAMPTZ,
    created_at    TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_settlement_ledger_order_id
    ON settlement_ledger (order_id)
    WHERE order_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_settlement_ledger_subsite_id ON settlement_ledger(subsite_id);
CREATE INDEX IF NOT EXISTS idx_settlement_ledger_status ON settlement_ledger(status);
CREATE INDEX IF NOT EXISTS idx_settlement_ledger_created_at ON settlement_ledger(created_at);

COMMENT ON TABLE settlement_ledger IS '子站结算台账：主站收入、平台成本与子站应得差额';
COMMENT ON COLUMN settlement_ledger.margin_amount IS '应付子站差额 = gross_amount - cost_amount；V1 不自动打款';

-- 5. 现有表的子站归属字段（可空，NULL 表示主站数据）
ALTER TABLE payment_orders ADD COLUMN IF NOT EXISTS subsite_id BIGINT;
ALTER TABLE usage_logs     ADD COLUMN IF NOT EXISTS subsite_id BIGINT;
ALTER TABLE api_keys       ADD COLUMN IF NOT EXISTS subsite_id BIGINT;

CREATE INDEX IF NOT EXISTS idx_payment_orders_subsite_id ON payment_orders(subsite_id);
CREATE INDEX IF NOT EXISTS idx_api_keys_subsite_id ON api_keys(subsite_id);

-- usage_logs 是最大的表，按仓库既有约定必须用 `CREATE INDEX CONCURRENTLY`
-- 且放在单独的 `_notx.sql` 迁移里，避免启动迁移时长时间持锁。
-- 该索引由后续用量归属任务（Task 4）以并发方式补充。

COMMENT ON COLUMN payment_orders.subsite_id IS '充值订单归属子站；NULL 表示主站订单';
COMMENT ON COLUMN usage_logs.subsite_id IS '调用用量归属子站；NULL 表示主站用量';
COMMENT ON COLUMN api_keys.subsite_id IS 'API Key 归属子站；NULL 表示主站 Key';
