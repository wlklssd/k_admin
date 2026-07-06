CREATE TABLE[goadmin_dict_type] (
 [id] int identity(1,1),
 [name] varchar(100) NOT NULL,
 [code] varchar(100) NOT NULL,
 [description] varchar(3000) DEFAULT NULL,
 [sort] int NOT NULL DEFAULT 0,
 [status] tinyint NOT NULL DEFAULT 1,
 [created_at] datetime NULL DEFAULT GETDATE(),
 [updated_at] datetime NULL DEFAULT GETDATE(),
 PRIMARY KEY ([id]),
 UNIQUE ([code])
)

CREATE INDEX [admin_dict_type_status_index] ON [goadmin_dict_type] ([status])

CREATE TABLE[goadmin_dict_data] (
 [id] int identity(1,1),
 [dict_type] varchar(100) NOT NULL,
 [label] varchar(100) NOT NULL,
 [value] varchar(100) NOT NULL,
 [color] varchar(50) NOT NULL DEFAULT '',
 [css_class] varchar(100) NOT NULL DEFAULT '',
 [is_default] tinyint NOT NULL DEFAULT 0,
 [sort] int NOT NULL DEFAULT 0,
 [status] tinyint NOT NULL DEFAULT 1,
 [remark] varchar(3000) DEFAULT NULL,
 [created_at] datetime NULL DEFAULT GETDATE(),
 [updated_at] datetime NULL DEFAULT GETDATE(),
 PRIMARY KEY ([id]),
 UNIQUE ([dict_type],[value])
)

CREATE INDEX [admin_dict_data_type_index] ON [goadmin_dict_data] ([dict_type])
CREATE INDEX [admin_dict_data_status_index] ON [goadmin_dict_data] ([status])

set IDENTITY_INSERT [goadmin_dict_type] ON

INSERT INTO[goadmin_dict_type] ([id],[name],[code],[description],[sort],[status],[created_at],[updated_at])
VALUES
  (1,'Gender','sys_gender','Common gender dictionary',1,1,'2026-07-06 00:00:00','2026-07-06 00:00:00'),
  (2,'Status','sys_status','Common enable/disable status dictionary',2,1,'2026-07-06 00:00:00','2026-07-06 00:00:00');

set IDENTITY_INSERT [goadmin_dict_type] OFF

set IDENTITY_INSERT [goadmin_dict_data] ON

INSERT INTO[goadmin_dict_data] ([id],[dict_type],[label],[value],[color],[css_class],[is_default],[sort],[status],[remark],[created_at],[updated_at])
VALUES
  (1,'sys_gender','Male','male','blue','',1,1,1,'','2026-07-06 00:00:00','2026-07-06 00:00:00'),
  (2,'sys_gender','Female','female','magenta','',0,2,1,'','2026-07-06 00:00:00','2026-07-06 00:00:00'),
  (3,'sys_status','Enable','enable','green','',1,1,1,'','2026-07-06 00:00:00','2026-07-06 00:00:00'),
  (4,'sys_status','Disable','disable','red','',0,2,1,'','2026-07-06 00:00:00','2026-07-06 00:00:00');

set IDENTITY_INSERT [goadmin_dict_data] OFF
