CREATE TABLE IF NOT EXISTS todos(
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title       TEXT NOT NULL,
    completed   BOOLEAN NOT NULL DEFAULT false,
    description TEXT NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
   
);

 CREATE INDEX idx_todos_user_id ON todos(user_id);