CREATE TABLE IF NOT EXISTS links (
 id serial,
 url varchar,
 nick varchar,
 created_at timestamp without time zone,

 PRIMARY KEY (id)
);
