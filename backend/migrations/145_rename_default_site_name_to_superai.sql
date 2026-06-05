-- Rename the legacy default brand only. Custom site names are preserved.
UPDATE settings
SET value = 'SuperAI',
    updated_at = NOW()
WHERE key = 'site_name'
  AND lower(trim(value)) IN ('sub2api', 'sub2 api', 'sub2-api');
