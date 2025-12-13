CREATE TABLE jit_requests (
                              id TEXT PRIMARY KEY,
                              user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                              role TEXT NOT NULL,
                              resource_id TEXT NULL,  -- project_id or other resource
                              duration_minutes INTEGER NOT NULL,
                              reason TEXT,
                              status TEXT NOT NULL DEFAULT 'pending', -- pending, approved, rejected
                              approved_by TEXT NULL REFERENCES users(id),
                              created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                              updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

                              CONSTRAINT fk_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX idx_jit_requests_user_id ON jit_requests(user_id);
CREATE INDEX idx_jit_requests_status ON jit_requests(status);