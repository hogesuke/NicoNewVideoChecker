CREATE TABLE new_videos (
  serial_no INT AUTO_INCREMENT,
  id INT,
  title NVARCHAR(255),
  post_datetime CHAR(12),
  status CHAR(1), -- 0: 未処理, 1: 処理済, 9: 削除済
  created_datetime TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_datetime TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (serial_no, id)
);

CREATE TABLE videos (
  serial_no INT AUTO_INCREMENT,
  id INT,
  prefix VARCHAR(10),
  title NVARCHAR(255),
  description NVARCHAR(2047),
  contributor_id INT,
  contributor_name NVARCHAR(32),
  thumbnail_url VARCHAR(511),
  post_datetime CHAR(12),
  length VARCHAR(10),
  view_counter INT,
  comment_counter INT,
  mylist_counter INT,
  created_datetime TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_datetime TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (serial_no, id)
);

CREATE TABLE contributors (
  id INT,
  name NVARCHAR(32),
  icon_url NVARCHAR(511),
  created_datetime TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_datetime TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id)
);

CREATE TABLE videos_contributors (
  video_id INT,
  contributor_id VARCHAR(12),
  created_datetime TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_datetime TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (video_id, contributor_id)
);

CREATE TABLE categories (
  id INT AUTO_INCREMENT,
  name NVARCHAR(32),
  created_datetime TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_datetime TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id)
);

CREATE TABLE videos_categories (
  video_id INT,
  category_id VARCHAR(12),
  created_datetime TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_datetime TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (video_id, category_id)
);

CREATE TABLE tags (
  id INT NOT NULL AUTO_INCREMENT,
  tag NVARCHAR(127),
  created_datetime TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_datetime TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id)
);

CREATE TABLE videos_tags (
  video_id INT,
  tag_id INT,
  created_datetime TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_datetime TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (video_id, tag_id)
);

CREATE TABLE users (
  id INT NOT NULL AUTO_INCREMENT,
  provider_id BIGINT,
  provider_name VARCHAR(32),
  raw_name NVARCHAR(32),
  name VARCHAR(32),
  created_datetime TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_datetime TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id, provider_id)
);

CREATE TABLE users_contributors (
  user_id INT,
  contributor_id INT,
  created_datetime TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_datetime TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (user_id, contributor_id)
);

CREATE TABLE completions (
  video_id INT,
  created_datetime TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_datetime TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (video_id)
);

CREATE TABLE users_completions (
  user_id INT,
  video_id INT,
  created_datetime TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_datetime TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (user_id, video_id)
);
