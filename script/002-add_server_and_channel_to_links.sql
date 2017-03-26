DO $$
  BEGIN
    BEGIN
      ALTER TABLE IF EXISTS links ADD COLUMN server varchar;
    EXCEPTION
      WHEN duplicate_column THEN RAISE NOTICE 'column "server" already exists in "links"; skipping';
    END;
  END;
$$;

DO $$
  BEGIN
    BEGIN
      ALTER TABLE IF EXISTS links ADD COLUMN channel varchar;
    EXCEPTION
      WHEN duplicate_column THEN RAISE NOTICE 'column "channel" already exists in "links"; skipping';
    END;
  END;
$$;
