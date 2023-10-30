CREATE TABLE sessions (
    `id` TEXT,
    -- i.e cookie name 'oauth', 'site'
    `name` TEXT NOT NULL,
    `value` TEXT NOT NULL,
    --
    PRIMARY KEY(id)
);