CREATE TABLE IF NOT EXISTS ignored_nicks (
  id serial,
  server varchar,
  nick varchar,
  created_at timestamp without time zone,

  PRIMARY KEY (id)
);
