ALTER TABLE youtube_publishes
  MODIFY status ENUM('queued','uploaded','failed','deleted') NOT NULL DEFAULT 'queued';
