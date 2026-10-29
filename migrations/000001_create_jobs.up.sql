
CREATE TABLE jobs (
    id BIGINT    GENERATED ALWAYS AS IDENTITY     PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL,
    kind TEXT NOT NULL,
    args JSONB NOT NULL,
    
    
    
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
