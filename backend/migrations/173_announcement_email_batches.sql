CREATE TABLE IF NOT EXISTS announcement_email_batches (
    id BIGSERIAL PRIMARY KEY,
    announcement_id BIGINT NOT NULL REFERENCES announcements(id) ON DELETE CASCADE,
    campaign_id UUID NOT NULL UNIQUE,
    title VARCHAR(200) NOT NULL,
    content TEXT NOT NULL,
    recipients JSONB NOT NULL DEFAULT '[]'::jsonb,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    attempt_count INTEGER NOT NULL DEFAULT 0,
    max_attempts INTEGER NOT NULL DEFAULT 5,
    total_count INTEGER NOT NULL DEFAULT 0,
    processed_count INTEGER NOT NULL DEFAULT 0,
    failed_count INTEGER NOT NULL DEFAULT 0,
    last_error TEXT,
    next_attempt_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    locked_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT announcement_email_batches_status_check
        CHECK (status IN ('pending', 'processing', 'retrying', 'completed', 'failed')),
    CONSTRAINT announcement_email_batches_attempts_check
        CHECK (attempt_count >= 0 AND max_attempts > 0 AND attempt_count <= max_attempts),
    CONSTRAINT announcement_email_batches_counts_check
        CHECK (total_count >= 0 AND processed_count >= 0 AND failed_count >= 0)
);

CREATE INDEX IF NOT EXISTS idx_announcement_email_batches_due
    ON announcement_email_batches (next_attempt_at, id)
    WHERE status IN ('pending', 'retrying');

CREATE INDEX IF NOT EXISTS idx_announcement_email_batches_processing
    ON announcement_email_batches (locked_at, id)
    WHERE status = 'processing';

CREATE INDEX IF NOT EXISTS idx_announcement_email_batches_announcement
    ON announcement_email_batches (announcement_id, created_at DESC, id DESC);

COMMENT ON TABLE announcement_email_batches IS '公告群发邮件持久化任务与发送状态';
COMMENT ON COLUMN announcement_email_batches.campaign_id IS '通知投递去重使用的稳定批次标识';
COMMENT ON COLUMN announcement_email_batches.recipients IS '创建任务时冻结的收件人邮箱列表';
