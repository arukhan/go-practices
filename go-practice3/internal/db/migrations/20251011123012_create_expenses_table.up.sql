CREATE TABLE IF NOT EXISTS expenses (
                                        id           SERIAL PRIMARY KEY,
                                        user_id      INTEGER NOT NULL,
                                        category_id  INTEGER NOT NULL,
                                        amount       NUMERIC(12,2) NOT NULL CHECK (amount > 0),
    currency     CHAR(3) NOT NULL,
    spent_at     TIMESTAMPTZ NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    note         TEXT NULL,
    CONSTRAINT fk_expenses_user
    FOREIGN KEY (user_id) REFERENCES users(id)
    ON UPDATE CASCADE ON DELETE RESTRICT,
    CONSTRAINT fk_expenses_category
    FOREIGN KEY (category_id) REFERENCES categories(id)
    ON UPDATE CASCADE ON DELETE RESTRICT
    );

CREATE INDEX IF NOT EXISTS idx_expenses_user_id ON expenses(user_id);
CREATE INDEX IF NOT EXISTS idx_expenses_user_spent_at ON expenses(user_id, spent_at);
