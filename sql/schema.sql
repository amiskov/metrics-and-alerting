DROP TYPE IF EXISTS metric_type CASCADE;
CREATE TYPE metric_type AS ENUM ('gauge', 'counter');

DROP TABLE IF EXISTS metrics CASCADE;
CREATE TABLE metrics (
  id SERIAL,
  type metric_type NOT NULL, 
  name VARCHAR(128) UNIQUE NOT NULL,
  value DOUBLE PRECISION,
  delta BIGINT,
  -- make sure we store only 1 number (add more fields if necessary)
  CHECK ((value IS NOT NULL)::INTEGER + (delta IS NOT NULL)::INTEGER = 1)
);