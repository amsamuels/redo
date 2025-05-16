-- +migrate Down
-- Step 1: Drop global slug constraint (if still present)
ALTER TABLE links DROP CONSTRAINT IF EXISTS links_slug_key;

-- Step 2: Add globally unique short_code column
DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_name = 'links' AND column_name = 'short_code'
  ) THEN
    ALTER TABLE links
    ADD COLUMN short_code TEXT UNIQUE NOT NULL DEFAULT encode(gen_random_bytes(4), 'hex');
  END IF;
END
$$;

-- Step 3: Add user-scoped uniqueness constraint only if not already present
DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_constraint
    WHERE conname = 'unique_user_slug'
  ) THEN
    ALTER TABLE links
    ADD CONSTRAINT unique_user_slug UNIQUE (user_id, slug);
  END IF;
END
$$;
