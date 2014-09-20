CREATE TABLE new_videos (
  id VARCHAR(255),
  title NVARCHAR(255),
  post_datetime CHAR(12),
  created_datetime TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_datetime TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);