-- 简单版本：直接重命名列（如果列存在）
-- 如果列不存在，GORM 会在下次启动时自动创建

-- 修复 auth_records 表
ALTER TABLE auth_records RENAME COLUMN device_d_id TO device_did;

-- 修复 auth_logs 表  
ALTER TABLE auth_logs RENAME COLUMN device_d_id TO device_did;

