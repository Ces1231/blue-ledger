-- AI assistant configuration per chapter
CREATE TABLE IF NOT EXISTS ai_configs (
  id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  chapter_id    UUID NOT NULL REFERENCES chapters(id) ON DELETE CASCADE,
  provider      TEXT NOT NULL DEFAULT 'claude' CHECK (provider IN ('claude','openai')),
  model         TEXT NOT NULL DEFAULT 'claude-sonnet-4-6',
  system_prompt TEXT NOT NULL DEFAULT 'You are a helpful assistant for a Phi Beta Sigma chapter. Be professional, supportive, and knowledgeable about fraternity operations.',
  enabled       BOOLEAN NOT NULL DEFAULT TRUE,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(chapter_id)
);

ALTER TABLE ai_configs ENABLE ROW LEVEL SECURITY;
CREATE POLICY chapter_isolation ON ai_configs
  USING (chapter_id = current_setting('app.chapter_id', TRUE)::UUID);

-- AI chat message history
CREATE TABLE IF NOT EXISTS ai_messages (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  chapter_id  UUID NOT NULL REFERENCES chapters(id) ON DELETE CASCADE,
  member_id   UUID NOT NULL REFERENCES members(id) ON DELETE CASCADE,
  role        TEXT NOT NULL CHECK (role IN ('user','assistant')),
  content     TEXT NOT NULL,
  provider    TEXT NOT NULL,
  model       TEXT NOT NULL,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_ai_messages_member ON ai_messages(member_id, created_at DESC);

ALTER TABLE ai_messages ENABLE ROW LEVEL SECURITY;
CREATE POLICY chapter_isolation ON ai_messages
  USING (chapter_id = current_setting('app.chapter_id', TRUE)::UUID);
