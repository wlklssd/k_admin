CREATE TABLE[goadmin_department] (
 [id] int identity(1,1),
 [name] varchar(100) NOT NULL,
 [code] varchar(100) NOT NULL DEFAULT '',
 [description] varchar(3000) DEFAULT NULL,
 [sort] int NOT NULL DEFAULT 0,
 [status] tinyint NOT NULL DEFAULT 1,
 [created_at] datetime NULL DEFAULT GETDATE(),
 [updated_at] datetime NULL DEFAULT GETDATE(),
 PRIMARY KEY ([id])
)

CREATE INDEX [admin_department_code_index] ON [goadmin_department] ([code])

CREATE TABLE[goadmin_department_roles] (
 [department_id] int NOT NULL,
 [role_id] int NOT NULL,
 [created_at] datetime NULL DEFAULT GETDATE(),
 [updated_at] datetime NULL DEFAULT GETDATE(),
 PRIMARY KEY ([department_id],[role_id])
)

CREATE INDEX [admin_department_roles_role_id_index] ON [goadmin_department_roles] ([role_id])
