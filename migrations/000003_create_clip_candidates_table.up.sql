CREATE TABLE clip_candidates (
  id INT NOT NULL AUTO_INCREMENT,
  recording_id INT NOT NULL,

  start_ms INT NOT NULL,
  end_ms INT NOT NULL,

  score FLOAT NOT NULL DEFAULT 0,

  detected_signals JSON NULL,

  status ENUM('new','rejected','approved') NOT NULL DEFAULT 'new',

  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

  PRIMARY KEY (id),
  KEY idx_candidates_recording_id (recording_id),
  KEY idx_candidates_status (status),

  CONSTRAINT fk_candidates_recording
    FOREIGN KEY (recording_id) REFERENCES recordings(id)
    ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
