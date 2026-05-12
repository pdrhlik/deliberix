ALTER TABLE survey ADD COLUMN moderation_enabled TINYINT(1) NOT NULL DEFAULT 1 AFTER statement_char_max;
