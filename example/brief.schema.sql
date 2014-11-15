BEGIN;

CREATE TABLE person (
    id text NOT NULL PRIMARY KEY,
    created_at datetime,
    updated_at datetime,
    deleted_at datetime,
    external_id text,
    name text,
    email text
);

CREATE TABLE tenant (
    id text NOT NULL PRIMARY KEY,
    created_at datetime,
    updated_at datetime,
    deleted_at datetime,
    creator_id text REFERENCES person(id),
    name text,
    title text,
    domain text
);

CREATE TABLE client (
    id text NOT NULL PRIMARY KEY,
    updated_at datetime,
    deleted_at datetime,
    tenant_id text REFERENCES tenant(id),
    name text,
    created_at datetime,
    creator_id text REFERENCES person(id)
);

CREATE TABLE comment (
    id text NOT NULL PRIMARY KEY,
    created_at datetime,
    updated_at datetime,
    creator_id text REFERENCES person(id),
    review_file_id text REFERENCES review_file(id),
    deleted_at datetime,
    comment text
);

CREATE TABLE event (
    id integer NOT NULL PRIMARY KEY,
    review_id text REFERENCES review(id),
    channel text,
    event text,
    payload json,
    created_at datetime,
    tenant_id text REFERENCES tenant(id),
    client_id text REFERENCES client(id),
    project_id text REFERENCES project(id)
);

CREATE TABLE file_version (
    file_id text REFERENCES file(id),
    content_type text,
    created_at datetime,
    updated_at datetime,
    deleted_at datetime,
    creator_id text REFERENCES person(id),
    name text,
    complete boolean,
    id text NOT NULL PRIMARY KEY
);

CREATE TABLE file (
    name text,
    deleted_at datetime,
    creator_id text REFERENCES person(id),
    project_id text REFERENCES project(id),
    id text NOT NULL PRIMARY KEY,
    created_at datetime,
    updated_at datetime
);

CREATE TABLE invite (
    accepted_at datetime,
    creator_id text REFERENCES person(id),
    membership_id text REFERENCES membership(id),
    email text,
    token text,
    id text NOT NULL PRIMARY KEY,
    created_at datetime,
    updated_at datetime,
    deleted_at datetime,
    expires_at datetime
);

CREATE TABLE membership (
    client_id text REFERENCES client(id),
    type text,
    project_id text REFERENCES project(id),
    id text NOT NULL PRIMARY KEY,
    created_at datetime,
    deleted_at datetime,
    creator_id text REFERENCES person(id),
    updated_at datetime,
    review_id text REFERENCES review(id),
    person_id text REFERENCES person(id),
    tenant_id text REFERENCES tenant(id)
);

CREATE TABLE project (
    updated_at datetime,
    deleted_at datetime,
    creator_id text REFERENCES person(id),
    name text,
    id text NOT NULL PRIMARY KEY,
    client_id text REFERENCES client(id),
    created_at datetime
);

CREATE TABLE review_file (
    created_at datetime,
    creator_id text REFERENCES person(id),
    review_id text REFERENCES review(id),
    id text NOT NULL PRIMARY KEY,
    updated_at datetime,
    deleted_at datetime,
    file_version_id text REFERENCES file_version(id)
);

CREATE TABLE review (
    created_at datetime,
    deleted_at datetime,
    project_id text REFERENCES project(id),
    message text,
    status text,
    expires_at datetime,
    id text NOT NULL PRIMARY KEY,
    updated_at datetime,
    creator_id text REFERENCES person(id),
    name text
);

END;
