CREATE TABLE IF NOT EXISTS "goadmin_department" (
`id` integer PRIMARY KEY autoincrement,
`name` CHAR(100) COLLATE NOCASE NOT NULL,
`code` CHAR(100) COLLATE NOCASE NOT NULL DEFAULT '',
`description` CHAR(3000) COLLATE NOCASE,
`sort` INT NOT NULL DEFAULT '0',
`status` INT NOT NULL DEFAULT '1',
`created_at` TIMESTAMP default CURRENT_TIMESTAMP,
`updated_at` TIMESTAMP default CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS `admin_department_code_index` ON `goadmin_department` (`code`);

CREATE TABLE IF NOT EXISTS "goadmin_department_roles" (
`department_id` INT NOT NULL,
`role_id` INT NOT NULL,
`created_at` TIMESTAMP default CURRENT_TIMESTAMP,
`updated_at` TIMESTAMP default CURRENT_TIMESTAMP,
UNIQUE (`department_id`, `role_id`)
);

CREATE INDEX IF NOT EXISTS `admin_department_roles_role_id_index` ON `goadmin_department_roles` (`role_id`);
