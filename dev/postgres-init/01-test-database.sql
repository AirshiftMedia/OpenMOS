-- Runs once, when the volume is first created. The store tests drop and
-- recreate the openmos schema, so they get a database of their own.
CREATE DATABASE openmos_test;
