-- +migrate Up

ALTER TABLE `sticker_store` ADD COLUMN cover_lim VARCHAR(100) NOT NULL DEFAULT '' COMMENT 'json格式封面';
