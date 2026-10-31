-- 1. Create queues table
CREATE TABLE queues (
    id INT    GENERATED ALWAYS AS IDENTITY     PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    worker_amount INT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL

    
);

-- 2. Create queue_kinds mapping table
CREATE TABLE queue_kinds (
    kind TEXT PRIMARY KEY,

    queue_id INT NOT NULL
        REFERENCES queues(id) ON DELETE CASCADE -- 必须引用 queues 表中存在的 id；联级删除
);



-- 4. Add queue_id to existing jobs
ALTER TABLE jobs
ADD COLUMN queue_id INT;

