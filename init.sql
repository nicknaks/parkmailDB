DROP SCHEMA IF EXISTS parkmaildb CASCADE;
CREATE EXTENSION IF NOT EXISTS citext;
CREATE SCHEMA parkmaildb;

DROP TABLE IF EXISTS parkmaildb."User" CASCADE;
DROP TABLE IF EXISTS parkmaildb."Post" CASCADE;
DROP TABLE IF EXISTS parkmaildb."Thread" CASCADE;
DROP TABLE IF EXISTS parkmaildb."Forum" CASCADE;
DROP TABLE IF EXISTS parkmaildb."Vote" CASCADE;
DROP TABLE IF EXISTS parkmaildb."Users_by_Forum" CASCADE;


CREATE UNLOGGED TABLE parkmaildb."User"
(
    Id SERIAL PRIMARY KEY,
    NickName CITEXT UNIQUE NOT NULL,
    FullName TEXT NOT NULL,
    About TEXT,
    Email CITEXT UNIQUE NOT NULL
);

CREATE UNLOGGED TABLE parkmaildb."Forum"
(
    Id SERIAL PRIMARY KEY,
    Title TEXT NOT NULL,
    "user" CITEXT REFERENCES parkmaildb."User"(NickName) NOT NULL,
    Slug CITEXT UNIQUE NOT NULL,
    Posts INT,
    Threads INT
);

CREATE UNLOGGED TABLE parkmaildb."Thread"
(
    Id SERIAL PRIMARY KEY,
    Title TEXT NOT NULL,
    Author CITEXT REFERENCES parkmaildb."User"(NickName) NOT NULL,
    Forum CITEXT REFERENCES parkmaildb."Forum"(Slug) NOT NULL,
    Message TEXT NOT NULL,
    Votes INT,
    Slug CITEXT UNIQUE DEFAULT citext(1),
    Created TIMESTAMP WITH TIME ZONE
);

CREATE UNLOGGED TABLE parkmaildb."Post"
(
    Id SERIAL PRIMARY KEY,
    Parent INT DEFAULT 0,
    Author CITEXT REFERENCES parkmaildb."User"(NickName) NOT NULL,
    Message TEXT NOT NULL,
    IsEdited bool NOT NULL DEFAULT FALSE,
    Forum CITEXT REFERENCES parkmaildb."Forum"(Slug) NOT NULL,
    Thread INT REFERENCES parkmaildb."Thread"(Id) NOT NULL,
    Created TIMESTAMP WITH TIME ZONE DEFAULT now(),
    Path INT[] DEFAULT ARRAY []::INTEGER[]
);

CREATE UNLOGGED TABLE parkmaildb."Users_by_Forum"
(
    Id SERIAL PRIMARY KEY,
    Forum CITEXT NOT NULL,
    "user" CITEXT REFERENCES parkmaildb."User"(NickName) NOT NULL,
    CONSTRAINT onlyOneUser UNIQUE (Forum, "user")
);

CREATE UNLOGGED TABLE parkmaildb."Vote"
(
    Id SERIAL PRIMARY KEY,
    ThreadId INT REFERENCES parkmaildb."Thread"(id) NOT NULL,
    "user" CITEXT REFERENCES parkmaildb."User"(NickName) NOT NULL,
    Value INT NOT NULL,
    CONSTRAINT onlyOneVote UNIQUE (ThreadId, "user")
);

-- добавление новой ветки
CREATE OR REPLACE FUNCTION inc_threads_of_forum() RETURNS TRIGGER AS $$
BEGIN
    UPDATE parkmaildb."Forum" SET threads = threads + 1 WHERE NEW.Forum = slug;
    INSERT INTO parkmaildb."Users_by_Forum" (forum, "user") VALUES (NEW.Forum, NEW.Author)
    ON CONFLICT DO NOTHING;
    RETURN NULL;
END
$$ LANGUAGE 'plpgsql';

CREATE TRIGGER create_thread_trigger
    AFTER INSERT ON parkmaildb."Thread"
    FOR EACH ROW EXECUTE PROCEDURE inc_threads_of_forum();

-- добавление нового голоса
CREATE OR REPLACE FUNCTION add_new_voice() RETURNS TRIGGER AS $$
BEGIN
    UPDATE parkmaildb."Thread" t SET votes = t.votes + NEW.Value WHERE t.Id = New.threadid;
    RETURN NULL;
END
$$ LANGUAGE 'plpgsql';

CREATE TRIGGER voice_trigger
    AFTER INSERT ON parkmaildb."Vote"
    FOR EACH ROW EXECUTE PROCEDURE add_new_voice();

-- Изменение голоса
CREATE OR REPLACE FUNCTION change_voice() RETURNS TRIGGER AS $$
BEGIN

    IF old.value <> new.value
    THEN UPDATE parkmaildb."Thread" t SET votes = (t.votes + new.value * 2) WHERE t.Id = New.threadid;
    END IF;
    RETURN new;
END
$$ LANGUAGE 'plpgsql';

CREATE TRIGGER voice_update_trigger
    AFTER UPDATE ON parkmaildb."Vote"
    FOR EACH ROW EXECUTE PROCEDURE change_voice();

-- Добавление поста
CREATE OR REPLACE FUNCTION add_post() RETURNS TRIGGER AS $$
BEGIN
    --     увеличить счетчик постов в форуме
    UPDATE parkmaildb."Forum" SET posts = posts + 1 WHERE Slug = NEW.forum;
--     добавить пользователя в таблицу форум-user
    INSERT INTO parkmaildb."Users_by_Forum" (forum, "user") VALUES (NEW.forum, NEW.author)
    ON CONFLICT DO NOTHING;
--     прописать путь
    NEW.path = (SELECT P.path FROM parkmaildb."Post" P WHERE P.id = NEW.parent LIMIT 1) || NEW.id;
    RETURN NEW;
END
$$ LANGUAGE 'plpgsql';

CREATE TRIGGER add_post
    BEFORE INSERT ON parkmaildb."Post"
    FOR EACH ROW EXECUTE PROCEDURE add_post();

CREATE INDEX IF NOT EXISTS user_nick ON parkmaildb."User" (nickname);
CREATE INDEX IF NOT EXISTS user_email ON parkmaildb."User" (email);

CREATE INDEX IF NOT EXISTS forum_slug ON parkmaildb."Forum" (slug);

CREATE INDEX IF NOT EXISTS thread_slug ON parkmaildb."Thread" (slug);
CREATE INDEX IF NOT EXISTS post_path_1 ON parkmaildb."Post" ((path[1]));
