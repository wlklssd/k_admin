ALTER TABLE `goadmin_users`
  ADD COLUMN `status` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'enable' AFTER `avatar`;

UPDATE `goadmin_users` SET `status` = 'enable' WHERE `status` = '';
