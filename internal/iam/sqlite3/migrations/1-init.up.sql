CREATE TABLE profiles (
    `id` TEXT,
    `login` TEXT NOT NULL UNIQUE,
    `photo_url` TEXT,
    `name` TEXT,
    --
    PRIMARY KEY (id)
);

CREATE TABLE accounts (
    `oauth` TEXT NOT NULL,
    `profile` TEXT NOT NULL,
    --
    FOREIGN KEY (profile) REFERENCES profiles (id),
    PRIMARY KEY (oauth)
);
