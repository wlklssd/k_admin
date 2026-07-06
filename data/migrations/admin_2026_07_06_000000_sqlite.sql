CREATE TABLE IF NOT EXISTS "goadmin_dict_type" (
`id` integer PRIMARY KEY autoincrement,
`name` CHAR(100) COLLATE NOCASE NOT NULL,
`code` CHAR(100) COLLATE NOCASE NOT NULL,
`description` CHAR(3000) COLLATE NOCASE,
`sort` INT NOT NULL DEFAULT '0',
`status` INT NOT NULL DEFAULT '1',
`created_at` TIMESTAMP default CURRENT_TIMESTAMP,
`updated_at` TIMESTAMP default CURRENT_TIMESTAMP,
UNIQUE (`code`)
);

CREATE INDEX IF NOT EXISTS `admin_dict_type_status_index` ON `goadmin_dict_type` (`status`);

CREATE TABLE IF NOT EXISTS "goadmin_dict_data" (
`id` integer PRIMARY KEY autoincrement,
`dict_type` CHAR(100) COLLATE NOCASE NOT NULL,
`label` CHAR(100) COLLATE NOCASE NOT NULL,
`value` CHAR(100) COLLATE NOCASE NOT NULL,
`color` CHAR(50) COLLATE NOCASE NOT NULL DEFAULT '',
`css_class` CHAR(100) COLLATE NOCASE NOT NULL DEFAULT '',
`is_default` INT NOT NULL DEFAULT '0',
`sort` INT NOT NULL DEFAULT '0',
`status` INT NOT NULL DEFAULT '1',
`remark` CHAR(3000) COLLATE NOCASE,
`created_at` TIMESTAMP default CURRENT_TIMESTAMP,
`updated_at` TIMESTAMP default CURRENT_TIMESTAMP,
UNIQUE (`dict_type`, `value`)
);

CREATE INDEX IF NOT EXISTS `admin_dict_data_type_index` ON `goadmin_dict_data` (`dict_type`);
CREATE INDEX IF NOT EXISTS `admin_dict_data_status_index` ON `goadmin_dict_data` (`status`);

INSERT OR IGNORE INTO `goadmin_dict_type` (`id`, `name`, `code`, `description`, `sort`, `status`, `created_at`, `updated_at`)
VALUES
  (1,'Gender','sys_gender','Common gender dictionary',1,1,'2026-07-06 00:00:00','2026-07-06 00:00:00'),
  (2,'Status','sys_status','Common enable/disable status dictionary',2,1,'2026-07-06 00:00:00','2026-07-06 00:00:00');

INSERT OR IGNORE INTO `goadmin_dict_data` (`id`, `dict_type`, `label`, `value`, `color`, `css_class`, `is_default`, `sort`, `status`, `remark`, `created_at`, `updated_at`)
VALUES
  (1,'sys_gender','Male','male','blue','',1,1,1,'','2026-07-06 00:00:00','2026-07-06 00:00:00'),
  (2,'sys_gender','Female','female','magenta','',0,2,1,'','2026-07-06 00:00:00','2026-07-06 00:00:00'),
  (3,'sys_status','Enable','enable','green','',1,1,1,'','2026-07-06 00:00:00','2026-07-06 00:00:00'),
  (4,'sys_status','Disable','disable','red','',0,2,1,'','2026-07-06 00:00:00','2026-07-06 00:00:00');
