CREATE TABLE todos (
    id          uuid        PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id     uuid        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title       text        NOT NULL,
    description text,
    completed   boolean     NOT NULL DEFAULT false,
    due_date    timestamptz,
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_todos_user_id       ON todos(user_id);
CREATE INDEX idx_todos_user_completed ON todos(user_id, completed);
CREATE INDEX idx_todos_user_created  ON todos(user_id, created_at);
