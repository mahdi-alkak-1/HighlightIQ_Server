CREATE TABLE recordings (
  id INT NOT NULL AUTO_INCREMENT,
  user_id INT NOT NULL,

  title VARCHAR(120) NOT NULL,
  original_filename VARCHAR(255) NOT NULL,
  storage_path VARCHAR(255) NOT NULL,

  duration_seconds INT NOT NULL DEFAULT 0,

  status ENUM('uploaded','processing','ready','failed') NOT NULL DEFAULT 'uploaded',

  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

  PRIMARY KEY (id),
  KEY idx_recordings_user_id (user_id),

  CONSTRAINT fk_recordings_user
    FOREIGN KEY (user_id) REFERENCES users(id)
    ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
