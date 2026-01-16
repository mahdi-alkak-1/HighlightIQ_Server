CREATE TABLE youtube_publishes (
  id INT NOT NULL AUTO_INCREMENT,
  clip_id INT NOT NULL,

  youtube_video_id VARCHAR(32) NOT NULL,
  youtube_url VARCHAR(255) NOT NULL,

  status ENUM('queued','uploaded','failed') NOT NULL DEFAULT 'queued',

  published_at DATETIME NULL,
  last_synced_at DATETIME NULL,

  views INT NOT NULL DEFAULT 0,
  likes INT NOT NULL DEFAULT 0,
  comments INT NOT NULL DEFAULT 0,

  analytics JSON NULL,

  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

  PRIMARY KEY (id),

  UNIQUE KEY uq_youtube_video_id (youtube_video_id),
  KEY idx_youtube_clip_id (clip_id),
  KEY idx_youtube_status (status),

  CONSTRAINT fk_youtube_clip
    FOREIGN KEY (clip_id) REFERENCES clips(id)
    ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
