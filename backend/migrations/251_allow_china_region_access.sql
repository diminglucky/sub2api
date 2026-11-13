-- Remove the former default China-region restrictions from persisted settings.
-- Persist an empty list too, so legacy environment/file defaults cannot restore it.
-- Preserve any restrictions configured for other countries, in their original order.
INSERT INTO settings (key, value, updated_at)
SELECT 'region_block_blocked_countries',
       COALESCE(jsonb_agg(country ORDER BY ordinal), '[]'::jsonb)::text,
       NOW()
FROM jsonb_array_elements_text(COALESCE(
    (SELECT NULLIF(BTRIM(value), '')::jsonb
     FROM settings WHERE key = 'region_block_blocked_countries'),
    '[]'::jsonb
)) WITH ORDINALITY AS countries(country, ordinal)
WHERE UPPER(BTRIM(country)) NOT IN ('CN', 'HK', 'MO', 'TW')
ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value, updated_at = EXCLUDED.updated_at;

-- With no remaining blocked countries, also disable any legacy enabled setting.
INSERT INTO settings (key, value, updated_at)
SELECT 'region_block_enabled', 'false', NOW()
WHERE (SELECT value::jsonb FROM settings WHERE key = 'region_block_blocked_countries') = '[]'::jsonb
ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value, updated_at = EXCLUDED.updated_at;
