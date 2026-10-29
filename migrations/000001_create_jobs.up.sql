
CREATE TABLE jobs (
    id BIGINT    GENERATED ALWAYS AS IDENTITY     PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL,
    kind TEXT NOT NULL,
    args JSONB, -- 这里为了方便测试等原因暂时允许留空！
    
    
    
    state TEXT NOT NULL
        CHECK (state IN (
            'available',
            'running',
            'retryable',
            'completed',
            'discarded'
        )),
    available_at TIMESTAMPTZ,
    
    
    
    attempt INTEGER NOT NULL DEFAULT 0,
    last_attempt TIMESTAMPTZ,
    last_error TEXT
);
