CREATE EXTENSION IF NOT EXISTS CITEXT;

DROP TABLE IF EXISTS parkmaildb."User" CASCADE;
DROP TABLE IF EXISTS parkmaildb."Post" CASCADE;
DROP TABLE IF EXISTS parkmaildb."Thread" CASCADE;
DROP TABLE IF EXISTS parkmaildb."Forum" CASCADE;
DROP TABLE IF EXISTS parkmaildb."Vote" CASCADE;


CREATE TABLE parkmaildb."User"
(
    Id SERIAL PRIMARY KEY,
    NickName CITEXT UNIQUE NOT NULL,
    FullName TEXT NOT NULL,
    About TEXT,
    Email CITEXT UNIQUE NOT NULL
);

CREATE TABLE parkmaildb."Forum"
(
    Id SERIAL PRIMARY KEY,
    Title TEXT NOT NULL,
    "user" CITEXT REFERENCES parkmaildb."User"(NickName) NOT NULL,
    Slug TEXT UNIQUE NOT NULL,
    Posts INT,
    Threads INT
);

CREATE TABLE parkmaildb."Thread"
(
    Id SERIAL PRIMARY KEY,
    Title TEXT NOT NULL,
    Author CITEXT REFERENCES parkmaildb."User"(NickName) NOT NULL,
    Forum TEXT REFERENCES parkmaildb."Forum"(Slug) NOT NULL,
    Message TEXT NOT NULL,
    Votes INT,
    Slug TEXT,
    Created TIMESTAMP,
    CONSTRAINT uniq UNIQUE (Title, Forum)
);

CREATE TABLE parkmaildb."Post"
(
    Id SERIAL PRIMARY KEY,
    Parent INT DEFAULT 0,
    Author CITEXT REFERENCES parkmaildb."User"(NickName) NOT NULL,
    Message TEXT NOT NULL,
    IsEdited bool NOT NULL DEFAULT FALSE,
    Forum TEXT REFERENCES parkmaildb."Forum"(Slug) NOT NULL,
    Thread INT REFERENCES parkmaildb."Thread"(Id) NOT NULL,
    Created TIMESTAMP DEFAULT now(),
    Path INT[]
);

CREATE TABLE parkmaildb."Users_by_Forum"
(
    Id SERIAL PRIMARY KEY,
    Forum TEXT NOT NULL,
    "user" CITEXT REFERENCES parkmaildb."User"(NickName) NOT NULL,
    CONSTRAINT onlyOneUser UNIQUE (Forum, "user")
);

CREATE TABLE parkmaildb."Vote"
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

-- добавление нового форума
CREATE OR REPLACE FUNCTION add_user_to_tmp() RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO parkmaildb."Users_by_Forum" (forum, "user") VALUES (NEW.slug, NEW."user")
    ON CONFLICT DO NOTHING;
    RETURN NULL;
END
$$ LANGUAGE 'plpgsql';

CREATE TRIGGER create_forum_trigger
    AFTER INSERT ON parkmaildb."Forum"
    FOR EACH ROW EXECUTE PROCEDURE add_user_to_tmp();

-- добавление нового голоса
CREATE OR REPLACE FUNCTION add_new_voice() RETURNS TRIGGER AS $$
BEGIN
    UPDATE parkmaildb."Thread" t SET votes = t.votes + NEW.Value WHERE t.Id = New.threadid;
    RETURN;
END
$$ LANGUAGE 'plpgsql';

CREATE TRIGGER voice_trigger
    AFTER INSERT ON parkmaildb."Vote"
    FOR EACH ROW EXECUTE PROCEDURE add_new_voice();

-- Изменение голоса
CREATE OR REPLACE FUNCTION change_voice() RETURNS TRIGGER AS $$
DECLARE
    voice INT;
BEGIN
    IF old.value = new.value
    THEN RETURN NULL;
    END IF;

    IF old.value = -1
    THEN voice = 2;
    ELSE voice = -2;
    END IF;

    UPDATE parkmaildb."Thread" t SET votes = t.votes + voice WHERE t.Id = New.threadid;
    RETURN NULL;
END
$$ LANGUAGE 'plpgsql';

CREATE TRIGGER voice_update_trigger
    AFTER UPDATE ON parkmaildb."Vote"
    FOR EACH ROW EXECUTE PROCEDURE change_voice();

-- Добавление поста
CREATE OR REPLACE FUNCTION add_post() RETURNS TRIGGER AS $$
DECLARE
    voice INT;
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