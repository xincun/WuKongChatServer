-- +migrate Up


-- 拉取用户记录
CREATE TABLE IF NOT EXISTS `bench_pulluser_record`(
    id   integer       not null primary key AUTO_INCREMENT,
    pull_no VARCHAR(40)  not null default '' comment 'pull编号',
    client_no   VARCHAR(40)   not null default '' comment '客户端编号',
    start_id integer not null DEFAULT 0 comment '开始用户id',
    end_id integer not null DEFAULT 0 comment '结束用户id',
    error   VARCHAR(255)    not null default '' comment '拉取失败的错误日志',
    status smallint    not null DEFAULT 0 comment '拉取状态 0.未拉取 1.拉取成功 2.拉取失败 3.拉取中',
    created_at   timeStamp     not null DEFAULT CURRENT_TIMESTAMP comment '创建时间', 
    updated_at  timeStamp     not null DEFAULT CURRENT_TIMESTAMP comment '更新时间'
);