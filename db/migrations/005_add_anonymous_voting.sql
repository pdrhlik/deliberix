ALTER TABLE survey
  ADD COLUMN allow_anonymous TINYINT(1) NOT NULL DEFAULT 0 AFTER moderation_enabled;

ALTER TABLE survey_participant
  MODIFY user_id INT UNSIGNED DEFAULT NULL,
  ADD COLUMN anon_session_id CHAR(36) DEFAULT NULL AFTER user_id,
  ADD CONSTRAINT chk_participant_identity
    CHECK ((user_id IS NULL) <> (anon_session_id IS NULL)),
  ADD UNIQUE KEY uq_survey_anon (survey_id, anon_session_id);

ALTER TABLE response
  MODIFY user_id INT UNSIGNED DEFAULT NULL,
  ADD COLUMN anon_session_id CHAR(36) DEFAULT NULL AFTER user_id,
  ADD CONSTRAINT chk_response_identity
    CHECK ((user_id IS NULL) <> (anon_session_id IS NULL)),
  ADD UNIQUE KEY uq_statement_anon (statement_id, anon_session_id);
