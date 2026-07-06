ALTER TABLE [goadmin_users]
ADD [status] varchar(50) NOT NULL DEFAULT 'enable'

UPDATE [goadmin_users] SET [status] = 'enable' WHERE [status] = ''
