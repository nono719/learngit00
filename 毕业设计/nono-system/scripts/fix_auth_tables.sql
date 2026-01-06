-- 修复认证相关表的列名
-- 如果表不存在，GORM 会自动创建正确的表结构
-- 如果表已存在，需要执行以下 SQL 来重命名列

-- 修复 auth_records 表
DO $$
BEGIN
    -- 检查列是否存在并重命名
    IF EXISTS (SELECT 1 FROM information_schema.columns 
               WHERE table_name='auth_records' AND column_name='device_d_id') THEN
        ALTER TABLE auth_records RENAME COLUMN device_d_id TO device_did;
    END IF;
    
    IF EXISTS (SELECT 1 FROM information_schema.columns 
               WHERE table_name='auth_records' AND column_name='source_domain' IS FALSE) THEN
        -- 如果 source_domain 列不存在，添加它
        ALTER TABLE auth_records ADD COLUMN IF NOT EXISTS source_domain VARCHAR(255);
    END IF;
    
    IF EXISTS (SELECT 1 FROM information_schema.columns 
               WHERE table_name='auth_records' AND column_name='target_domain' IS FALSE) THEN
        ALTER TABLE auth_records ADD COLUMN IF NOT EXISTS target_domain VARCHAR(255);
    END IF;
    
    IF EXISTS (SELECT 1 FROM information_schema.columns 
               WHERE table_name='auth_records' AND column_name='tx_hash' IS FALSE) THEN
        ALTER TABLE auth_records ADD COLUMN IF NOT EXISTS tx_hash VARCHAR(255);
    END IF;
END $$;

-- 修复 auth_logs 表
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.columns 
               WHERE table_name='auth_logs' AND column_name='device_d_id') THEN
        ALTER TABLE auth_logs RENAME COLUMN device_d_id TO device_did;
    END IF;
    
    IF EXISTS (SELECT 1 FROM information_schema.columns 
               WHERE table_name='auth_logs' AND column_name='source_domain' IS FALSE) THEN
        ALTER TABLE auth_logs ADD COLUMN IF NOT EXISTS source_domain VARCHAR(255);
    END IF;
    
    IF EXISTS (SELECT 1 FROM information_schema.columns 
               WHERE table_name='auth_logs' AND column_name='target_domain' IS FALSE) THEN
        ALTER TABLE auth_logs ADD COLUMN IF NOT EXISTS target_domain VARCHAR(255);
    END IF;
END $$;

