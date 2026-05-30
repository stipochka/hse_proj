-- заменяем stored_name на s3_key (путь к объекту в S3-совместимом хранилище)
ALTER TABLE activity_files RENAME COLUMN stored_name TO s3_key;
