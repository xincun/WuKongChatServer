-- +migrate Up

-- 标签表
create table `label`(
    id                   integer                not null primary key AUTO_INCREMENT,
    uid                  VARCHAR(40)            not null default '',                     
    name                 VARCHAR(40)            not null default '',                     -- 标签名称
    member_uids          TEXT,                                                            -- 用户UIDs   
    created_at           timeStamp              not null DEFAULT CURRENT_TIMESTAMP,      -- 创建时间
    updated_at           timeStamp              not null DEFAULT CURRENT_TIMESTAMP       -- 更新时间
);


-- -- +migrate StatementBegin
-- CREATE TRIGGER label_updated_at
--   BEFORE UPDATE
--   ON `label` for each row 
--   BEGIN
--     set NEW.updated_at = NOW();
--   END;
-- -- +migrate StatementEnd