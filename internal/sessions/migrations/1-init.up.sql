-- used during the oauth process
CREATE TABLE oauth (
    `id` TEXT,
    `state` TEXT NOT NULL,
    PRIMARY KEY (id, state)
);
-- used for signed in users
CREATE TABLE site (
    `id` TEXT,
    `profile` TEXT NOT NULL,
    --
    PRIMARY KEY (id, profile)
);
