-- 子站按尺寸的生图一口价。
--
-- 主站生图价来自分组字段 image_price_1k/2k/4k（按尺寸的一口价）。子站需要
-- 能对每个尺寸单独定价，因此 substite_prices 增加同样的三个字段；NULL 表示
-- 未覆盖，回退到主站价格。

ALTER TABLE subsite_prices
    ADD COLUMN IF NOT EXISTS image_price_1k DECIMAL(20,12),
    ADD COLUMN IF NOT EXISTS image_price_2k DECIMAL(20,12),
    ADD COLUMN IF NOT EXISTS image_price_4k DECIMAL(20,12);

COMMENT ON COLUMN subsite_prices.image_price_1k IS '子站生图 1K 一口价（USD/张）；NULL 回退主站价';
COMMENT ON COLUMN subsite_prices.image_price_2k IS '子站生图 2K 一口价（USD/张）；NULL 回退主站价';
COMMENT ON COLUMN subsite_prices.image_price_4k IS '子站生图 4K 一口价（USD/张）；NULL 回退主站价';
