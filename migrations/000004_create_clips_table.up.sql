CREATE TABLE clips (
  id INT NOT NULL AUTO_INCREMENT,

  user_id INT NOT NULL,
  recording_id INT NOT NULL,
  candidate_id INT NULL,

  title VARCHAR(120) NOT NULL,
  caption TEXT NULL,

  start_ms INT NOT NULL,
  end_ms INT NOT NULL,
  duration_seconds INT NOT NULL DEFAULT 0,

  status ENUM('draft','ready','published','failed') NOT NULL DEFAULT 'draft',

  export_path VARCHAR(255) NULL,

  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

  PRIMARY KEY (id),

  KEY idx_clips_user_id (user_id),
  KEY idx_clips_recording_id (recording_id),
  KEY idx_clips_candidate_id (candidate_id),
  KEY idx_clips_status (status),

  CONSTRAINT fk_clips_user
    FOREIGN KEY (user_id) REFERENCES users(id)
    ON DELETE CASCADE,

  CONSTRAINT fk_clips_recording
    FOREIGN KEY (recording_id) REFERENCES recordings(id)
    ON DELETE CASCADE,

  CONSTRAINT fk_clips_candidate
    FOREIGN KEY (candidate_id) REFERENCES clip_candidates(id)
    ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
