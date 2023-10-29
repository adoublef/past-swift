CREATE TABLE profiles (
    `id` TEXT NOT NULL,
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
