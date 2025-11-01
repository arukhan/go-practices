
CREATE TABLE IF NOT EXISTS categories (
                                          id   SERIAL PRIMARY KEY,
                                          name TEXT NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS products (
                                        id          SERIAL PRIMARY KEY,
                                        name        TEXT NOT NULL,
                                        category_id INT  NOT NULL REFERENCES categories(id),
    price       INT  NOT NULL
    );

INSERT INTO categories (name) VALUES
                                  ('phones'),
                                  ('laptops'),
                                  ('audio')
    ON CONFLICT (name) DO NOTHING;

INSERT INTO products (name, category_id, price) VALUES
                                                    ('iPhone', (SELECT id FROM categories WHERE name='phones'), 400000),
                                                    ('Galaxy', (SELECT id FROM categories WHERE name='phones'), 320000),
                                                    ('MacBook Air', (SELECT id FROM categories WHERE name='laptops'), 650000),
                                                    ('ThinkPad', (SELECT id FROM categories WHERE name='laptops'), 580000),
                                                    ('AirPods', (SELECT id FROM categories WHERE name='audio'), 90000),
                                                    ('WH-1000XM5', (SELECT id FROM categories WHERE name='audio'), 180000);
