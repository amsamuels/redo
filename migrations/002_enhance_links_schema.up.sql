-- +migrate Up

-- --- make_slug_user_scoped: allow duplicate slugs across users, but not within the same user ---
ALTER TABLE links
DROP CONSTRAINT IF EXISTS links_slug_key;

ALTER TABLE links
ADD CONSTRAINT unique_user_slug UNIQUE (user_id, slug);

CREATE INDEX IF NOT EXISTS idx_links_user_slug ON links(user_id, slug);

-- 1. Smart Device Targeting: per-device fallback URLs
ALTER TABLE links
ADD COLUMN device_targeting JSONB DEFAULT '{}'::JSONB;

-- 2. Dynamic UTM Injection
ALTER TABLE links
ADD COLUMN auto_append_utm BOOLEAN DEFAULT FALSE,
ADD COLUMN utm_defaults JSONB DEFAULT '{}'::JSONB;

-- 3. Real-Time Analytics Enhancements
ALTER TABLE clicks
ADD COLUMN device_type TEXT,
ADD COLUMN country TEXT,
ADD COLUMN conversion BOOLEAN DEFAULT FALSE;

-- 4. Retargeting Pixel Injection
ALTER TABLE links
ADD COLUMN pixels JSONB DEFAULT '[]'::JSONB;

-- 5. Custom Branding Support (CNAMEs)
CREATE TABLE IF NOT EXISTS custom_domains (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    domain TEXT NOT NULL UNIQUE,
    is_verified BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT now()
);

ALTER TABLE links
ADD COLUMN custom_domain_id UUID REFERENCES custom_domains(id);

-- 6. Smart Error Handling
ALTER TABLE links
ADD COLUMN fallback_url TEXT,
ADD COLUMN check_status_on_redirect BOOLEAN DEFAULT FALSE;

-- 7. Instant Link Editing (revision tracking)
CREATE TABLE IF NOT EXISTS link_revisions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    link_id UUID NOT NULL REFERENCES links(id) ON DELETE CASCADE,
    previous_destination TEXT,
    changed_at TIMESTAMPTZ DEFAULT now()
);

-- 8. Auto-Expire Links
ALTER TABLE links
ADD COLUMN expire_after_clicks INT,
ADD COLUMN expire_after_days INT,
ADD COLUMN deactivated_at TIMESTAMPTZ;

-- 9. IP Intelligence + High-Value Notification System
ALTER TABLE clicks
ADD COLUMN org_name TEXT,
ADD COLUMN is_high_value BOOLEAN DEFAULT FALSE;

CREATE TABLE IF NOT EXISTS notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    link_id UUID NOT NULL REFERENCES links(id),
    click_id UUID NOT NULL REFERENCES clicks(id),
    message TEXT,
    created_at TIMESTAMPTZ DEFAULT now(),
    sent_at TIMESTAMPTZ
);

-- Indexes for lookup performance
CREATE INDEX IF NOT EXISTS idx_clicks_org_name ON clicks(org_name);
CREATE INDEX IF NOT EXISTS idx_notifications_user_id ON notifications(user_id);

-- +migrate Down

-- Revert make_slug_user_scoped
ALTER TABLE links
DROP CONSTRAINT IF EXISTS unique_user_slug;

ALTER TABLE links
ADD CONSTRAINT links_slug_key UNIQUE (slug);
