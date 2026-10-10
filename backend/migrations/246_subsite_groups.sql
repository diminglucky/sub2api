-- 子站与分组的归属关系。
--
-- 之前用 subsite_prices 的 group 行充当"这个分组开放给子站"，会和定价耦合。
-- 现在把归属拆出来：subsite_groups 只表示"这个分组开放给这个子站"，价格默认
-- 完全跟随主站；subsite_prices 只在子站显式覆盖价格时才写入。

CREATE TABLE IF NOT EXISTS subsite_groups (
    id         BIGSERIAL PRIMARY KEY,
    subsite_id BIGINT      NOT NULL REFERENCES subsites(id) ON DELETE CASCADE,
    group_id   BIGINT      NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (subsite_id, group_id)
);

CREATE INDEX IF NOT EXISTS idx_subsite_groups_subsite_id ON subsite_groups(subsite_id);
CREATE INDEX IF NOT EXISTS idx_subsite_groups_group_id ON subsite_groups(group_id);

COMMENT ON TABLE subsite_groups IS '子站开放的分组；价格默认完全跟随主站';
