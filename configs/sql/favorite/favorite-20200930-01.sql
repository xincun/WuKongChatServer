-- +migrate Up

-- 收藏表
create table `favorite`
(
  id            integer                not null primary key AUTO_INCREMENT,
  type          integer                not null default 0,                      -- 收藏类型
  uid           VARCHAR(40)            not null default '',                     -- 作者uid
  unique_key    VARCHAR(40)            not null default '',                     -- 唯一key
  author_uid    VARCHAR(40)            not null default '',                     -- 作者uid
  author_name   VARCHAR(40)            not null default '',                     -- 作者名字
  payload       text                   ,                                        -- 收藏内容
  created_at    timeStamp              not null DEFAULT CURRENT_TIMESTAMP,      -- 创建时间
  updated_at    timeStamp              not null DEFAULT CURRENT_TIMESTAMP       -- 更新时间
);

-- -- +migrate StatementBegin
-- CREATE TRIGGER favorite_updated_at
--   BEFORE UPDATE
--   ON `favorite` for each row 
--   BEGIN
--     set NEW.updated_at = NOW();
--   END;
-- -- +migrate StatementEnd