CREATE TABLE IF NOT EXISTS lottery_events (
    id BIGSERIAL PRIMARY KEY,
    title VARCHAR(200) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    status VARCHAR(20) NOT NULL DEFAULT 'draft',
    starts_at TIMESTAMPTZ NULL,
    draw_at TIMESTAMPTZ NOT NULL,
    drawn_at TIMESTAMPTZ NULL,
    created_by BIGINT NULL REFERENCES users(id) ON DELETE SET NULL,
    updated_by BIGINT NULL REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT lottery_events_status_check CHECK (status IN ('draft', 'active', 'drawn', 'archived')),
    CONSTRAINT lottery_events_schedule_check CHECK (starts_at IS NULL OR starts_at < draw_at)
);

CREATE TABLE IF NOT EXISTS lottery_prizes (
    id BIGSERIAL PRIMARY KEY,
    event_id BIGINT NOT NULL REFERENCES lottery_events(id) ON DELETE CASCADE,
    type VARCHAR(20) NOT NULL,
    name VARCHAR(200) NOT NULL,
    quantity INT NOT NULL DEFAULT 1,
    amount DECIMAL(20, 8) NULL,
    card_content TEXT NULL,
    sort_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT lottery_prizes_type_check CHECK (type IN ('balance', 'card')),
    CONSTRAINT lottery_prizes_quantity_check CHECK (quantity > 0),
    CONSTRAINT lottery_prizes_balance_check CHECK (type <> 'balance' OR (amount IS NOT NULL AND amount > 0)),
    CONSTRAINT lottery_prizes_card_check CHECK (type <> 'card' OR (card_content IS NOT NULL AND LENGTH(BTRIM(card_content)) > 0))
);

CREATE TABLE IF NOT EXISTS lottery_entries (
    id BIGSERIAL PRIMARY KEY,
    event_id BIGINT NOT NULL REFERENCES lottery_events(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    joined_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (event_id, user_id)
);

CREATE TABLE IF NOT EXISTS lottery_winners (
    id BIGSERIAL PRIMARY KEY,
    event_id BIGINT NOT NULL REFERENCES lottery_events(id) ON DELETE CASCADE,
    prize_id BIGINT NOT NULL REFERENCES lottery_prizes(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    prize_type VARCHAR(20) NOT NULL,
    prize_name VARCHAR(200) NOT NULL,
    amount DECIMAL(20, 8) NULL,
    card_content TEXT NULL,
    delivered_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (event_id, user_id),
    CONSTRAINT lottery_winners_type_check CHECK (prize_type IN ('balance', 'card'))
);

CREATE INDEX IF NOT EXISTS idx_lottery_events_status_draw_at ON lottery_events(status, draw_at);
CREATE INDEX IF NOT EXISTS idx_lottery_entries_event_id ON lottery_entries(event_id);
CREATE INDEX IF NOT EXISTS idx_lottery_entries_user_id ON lottery_entries(user_id);
CREATE INDEX IF NOT EXISTS idx_lottery_winners_event_id ON lottery_winners(event_id);
CREATE INDEX IF NOT EXISTS idx_lottery_winners_user_id ON lottery_winners(user_id);
