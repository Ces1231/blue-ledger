-- 023_challenges.up.sql
-- Member challenge/game engine

CREATE TABLE IF NOT EXISTS challenges (
  id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  chapter_id      UUID        NOT NULL REFERENCES chapters(id)  ON DELETE CASCADE,
  challenger_id   UUID        NOT NULL REFERENCES members(id)   ON DELETE CASCADE,
  challenged_id   UUID        NOT NULL REFERENCES members(id)   ON DELETE CASCADE,
  type            TEXT        NOT NULL CHECK (type IN ('xp_duel','service_race','trivia','streak_showdown')),
  status          TEXT        NOT NULL DEFAULT 'pending'
                              CHECK (status IN ('pending','accepted','declined','active','completed','expired')),
  xp_stake        INT         NOT NULL DEFAULT 50 CHECK (xp_stake >= 10 AND xp_stake <= 500),
  game_data       JSONB       NOT NULL DEFAULT '{}',
  winner_id       UUID        REFERENCES members(id) ON DELETE SET NULL,
  expires_at      TIMESTAMPTZ NOT NULL DEFAULT NOW() + INTERVAL '24 hours',
  accepted_at     TIMESTAMPTZ,
  completed_at    TIMESTAMPTZ,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_challenges_chapter    ON challenges(chapter_id);
CREATE INDEX idx_challenges_challenger ON challenges(challenger_id);
CREATE INDEX idx_challenges_challenged ON challenges(challenged_id);
CREATE INDEX idx_challenges_status     ON challenges(status);
