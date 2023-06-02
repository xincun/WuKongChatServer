-- +migrate Up

-- 表情表
create table `sticker`(
    id                   integer                not null primary key AUTO_INCREMENT,
    title                VARCHAR(40)            not null default '',                   -- 表情名字
    path                 VARCHAR(100)           not null default '',                   -- 表情地址  
    category             VARCHAR(100)           not null default '',                   -- 分类 
    user_custom          int                    not null default 0,                    -- 用户自定义表情
    searchable_word      VARCHAR(40)            not null default '',                   -- 搜索关键字
    format               varchar(40)            not null default '',                   -- 表情格式
    width                int                    not null default 0,                    -- 宽
    height               int                    not null default 0,                    -- 高
    placeholder          blob                   not null ,                             -- 占位符 
    created_at           timeStamp              not null DEFAULT CURRENT_TIMESTAMP,    -- 创建时间
    updated_at           timeStamp              not null DEFAULT CURRENT_TIMESTAMP     -- 更新时间
);

-- -- +migrate StatementBegin
-- CREATE TRIGGER sticker_updated_at
--   BEFORE UPDATE
--   ON `sticker` for each row 
--   BEGIN
--     set NEW.updated_at = NOW();
--   END;
-- -- +migrate StatementEnd

-- 用户表情
create table `sticker_custom`(
    id                   integer                not null primary key AUTO_INCREMENT,
    uid                  VARCHAR(100)           not null default '',                   -- 用户id
    path                 VARCHAR(100)           not null default '',                   -- 表情地址 
    sort_num             integer                not null default 0,                    -- 序号
    width                int                    not null default 0,                    -- 宽
    height               int                    not null default 0,                    -- 高
    placeholder          blob                   not null,                              -- 占位符 
    category             VARCHAR(100)           not null default '',                   -- 分类 
    format               varchar(40)            not null default '',                   -- 表情格式
    created_at           timeStamp              not null DEFAULT CURRENT_TIMESTAMP,    -- 创建时间
    updated_at           timeStamp              not null DEFAULT CURRENT_TIMESTAMP     -- 更新时间
);

-- -- +migrate StatementBegin
-- CREATE TRIGGER sticker_custom_updated_at
--   BEFORE UPDATE
--   ON `sticker_custom` for each row 
--   BEGIN
--     set NEW.updated_at = NOW();
--   END;
-- -- +migrate StatementEnd

-- 用户表情分类
create table `sticker_user_category`(
    id                   integer                not null primary key AUTO_INCREMENT,
    uid                  VARCHAR(100)           not null default '', -- 用户id
    category             VARCHAR(100)           not null default '', -- 分类 
    sort_num             integer                not null default 0,  -- 序号
    created_at           timeStamp              not null DEFAULT CURRENT_TIMESTAMP,    -- 创建时间
    updated_at           timeStamp              not null DEFAULT CURRENT_TIMESTAMP     -- 更新时间
);

-- -- +migrate StatementBegin
-- CREATE TRIGGER sticker_user_category_updated_at
--   BEFORE UPDATE
--   ON `sticker_user_category` for each row 
--   BEGIN
--     set NEW.updated_at = NOW();
--   END;
-- -- +migrate StatementEnd


-- 表情商店
create table `sticker_store`(
    id                   integer                not null primary key AUTO_INCREMENT,
    category             VARCHAR(40)            not null default '', -- 表情分类
    title                VARCHAR(100)           not null default '', -- 分类名字
    `desc`                 VARCHAR(100)           not null default '', -- 说明
    cover                VARCHAR(100)            not null default '', -- 封面
    is_gone              smallint               not null default 0, -- 是否隐藏
    created_at           timeStamp              not null DEFAULT CURRENT_TIMESTAMP,    -- 创建时间
    updated_at           timeStamp              not null DEFAULT CURRENT_TIMESTAMP     -- 更新时间
);

-- -- +migrate StatementBegin
-- CREATE TRIGGER sticker_store_updated_at
--   BEFORE UPDATE
--   ON `sticker_store` for each row 
--   BEGIN
--     set NEW.updated_at = NOW();
--   END;
-- -- +migrate StatementEnd
