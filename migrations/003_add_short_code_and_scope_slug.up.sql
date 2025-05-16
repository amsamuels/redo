-- Step 1: Drop global slug constraint (if it still exists)
ALTER TABLE links DROP CONSTRAINT IF EXISTS links_slug_key;

-- Step 2: Add new globally unique short_code for canonical link resolution
ALTER TABLE links
ADD COLUMN short_code TEXT UNIQUE NOT NULL DEFAULT encode(gen_random_bytes(4), 'hex');

-- Step 3: Make slug scoped to user (or domain)
ALTER TABLE links
ADD CONSTRAINT unique_user_slug UNIQUE (user_id, slug);