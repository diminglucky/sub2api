-- 下游子站：把结算台账从“充值订单口径”改为“实际 API 用量口径”。
--
-- 业务规则（2026-10 澄清）：
--   主站模型价 = 子站批发成本；子站售价默认等于主站价。
--   子站管理员可设置更高售价，差额子站应得 = 子站售价收入 - 主站成本价。
--   结算只来自真实 API 用量，不再来自充值订单金额或到账美元余额。
--
-- 幂等键：每一条 usage 用 (subsite_id, usage_request_id) 唯一。
-- usage_request_id 取 usage_logs.request_id（同一请求稳定），避免依赖异步
-- 批量插入后才能拿到的 usage_logs 主键。

ALTER TABLE settlement_ledger ADD COLUMN IF NOT EXISTS usage_request_id TEXT;
ALTER TABLE settlement_ledger ADD COLUMN IF NOT EXISTS billing_mode VARCHAR(20);
ALTER TABLE settlement_ledger ALTER COLUMN currency SET DEFAULT 'USD';

CREATE UNIQUE INDEX IF NOT EXISTS idx_settlement_ledger_usage
    ON settlement_ledger (subsite_id, usage_request_id)
    WHERE usage_request_id IS NOT NULL;

COMMENT ON COLUMN settlement_ledger.usage_request_id IS '产生该结算行的 API 请求 ID；V1 结算取自真实用量';
COMMENT ON COLUMN settlement_ledger.billing_mode IS '计费模式（token/per_request/image/video）';
COMMENT ON COLUMN settlement_ledger.gross_amount IS '子站售价收入（用户实际扣费，USD）';
COMMENT ON COLUMN settlement_ledger.cost_amount IS '主站批发成本（主站价，USD）';
