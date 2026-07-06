CREATE TABLE `goadmin_dict_type` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL,
  `code` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL,
  `description` varchar(3000) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `sort` int(11) unsigned NOT NULL DEFAULT '0',
  `status` tinyint(3) unsigned NOT NULL DEFAULT '1',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `admin_dict_type_code_unique` (`code`),
  KEY `admin_dict_type_status_index` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE `goadmin_dict_data` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `dict_type` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL,
  `label` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL,
  `value` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL,
  `color` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `css_class` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `is_default` tinyint(3) unsigned NOT NULL DEFAULT '0',
  `sort` int(11) unsigned NOT NULL DEFAULT '0',
  `status` tinyint(3) unsigned NOT NULL DEFAULT '1',
  `remark` varchar(3000) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `admin_dict_data_type_value_unique` (`dict_type`,`value`),
  KEY `admin_dict_data_type_index` (`dict_type`),
  KEY `admin_dict_data_status_index` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

INSERT INTO `goadmin_dict_type` (`id`, `name`, `code`, `description`, `sort`, `status`, `created_at`, `updated_at`)
VALUES
  (1,'Gender','sys_gender','Common gender dictionary',1,1,'2026-07-06 00:00:00','2026-07-06 00:00:00'),
  (2,'Status','sys_status','Common enable/disable status dictionary',2,1,'2026-07-06 00:00:00','2026-07-06 00:00:00');

INSERT INTO `goadmin_dict_data` (`id`, `dict_type`, `label`, `value`, `color`, `css_class`, `is_default`, `sort`, `status`, `remark`, `created_at`, `updated_at`)
VALUES
  (1,'sys_gender','Male','male','blue','',1,1,1,'','2026-07-06 00:00:00','2026-07-06 00:00:00'),
  (2,'sys_gender','Female','female','magenta','',0,2,1,'','2026-07-06 00:00:00','2026-07-06 00:00:00'),
  (3,'sys_status','Enable','enable','green','',1,1,1,'','2026-07-06 00:00:00','2026-07-06 00:00:00'),
  (4,'sys_status','Disable','disable','red','',0,2,1,'','2026-07-06 00:00:00','2026-07-06 00:00:00');
