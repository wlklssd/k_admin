ALTER TABLE `goadmin_users`
ADD COLUMN `status` CHAR(50) COLLATE NOCASE NOT NULL DEFAULT 'enable';

UPDATE `goadmin_users` SET `status` = 'enable' WHERE `status` = '';
