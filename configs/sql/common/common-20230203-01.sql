-- +migrate Up

-- app 模块管理
create table `app_module`(
    id                   integer                not null primary key AUTO_INCREMENT,
    sid                  varchar(40)            not null default '',                  -- 模块ID    
    name                 VARCHAR(40)            not null default '',                  -- 模块名称
    `desc`               varchar(100)           not null default '',                  -- 模块介绍
    status               smallint               not null default 0,                   -- 模块状态 1.可用 0.不可用
    created_at           timeStamp              not null DEFAULT CURRENT_TIMESTAMP,   -- 创建时间
    updated_at           timeStamp              not null DEFAULT CURRENT_TIMESTAMP    -- 更新时间
);
CREATE  INDEX app_module_sid_idx on `app_module` (sid);