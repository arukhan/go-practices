CREATE TABLE IF NOT EXISTS categories (
                                          id       SERIAL PRIMARY KEY,
                                          name     TEXT NOT NULL,
                                          user_id  INTEGER NULL,
                                          CONSTRAINT fk_categories_user
                                          FOREIGN KEY (user_id) REFERENCES users(id)
    ON UPDATE CASCADE ON DELETE SET NULL,
    CONSTRAINT uq_categories_user_name UNIQUE (user_id, name)
    );

CREATE INDEX IF NOT EXISTS idx_categories_user_id ON categories(user_id);
