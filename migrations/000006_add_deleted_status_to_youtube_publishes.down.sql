UPDATE youtube_publishes
  SET status = 'failed'
  WHERE status = 'deleted';

ALTER TABLE youtube_publishes
  MODIFY status ENUM('queued','uploaded','failed') NOT NULL DEFAULT 'queued';
