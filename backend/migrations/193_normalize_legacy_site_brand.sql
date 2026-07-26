-- Replace only the source project's historical default brand.
-- Operator-authored site names and all business data remain untouched.
UPDATE settings
SET value = '3API', updated_at = NOW()
WHERE key = 'site_name'
  AND LOWER(BTRIM(value)) = 'sub2api';

UPDATE settings
SET value = 'AI API gateway for unified model access', updated_at = NOW()
WHERE key = 'site_subtitle'
  AND LOWER(BTRIM(value)) = 'subscription to api conversion platform';
