BEGIN;

CREATE TABLE "users"
(
    id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name VARCHAR (32) UNIQUE NOT NULL
);

CREATE TABLE "groups"
(
    id INT GENERATED ALWAYS AS IDENTITY (INCREMENT -1) PRIMARY KEY,
    name VARCHAR (32) UNIQUE NOT NULL
);

CREATE TABLE "user_groups"
(
    id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    group_id INT,
    user_id  INT,
    CONSTRAINT user_groups_group_id_fkey FOREIGN KEY (group_id) REFERENCES groups (id),
    CONSTRAINT user_groups_user_id_fkey  FOREIGN KEY (user_id)  REFERENCES users  (id)
);

CREATE TABLE "messages"
(
    id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    response_to INT,
    sent_at TIMESTAMP NOT NULL DEFAULT NOW(),
    sender INT,
    receiver INT,
    subject VARCHAR (255),
    body VARCHAR (2000)
);

COMMIT;
