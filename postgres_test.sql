-- PostgreSQL Migration File for neormgo_test database
-- This file creates the database, tables, seeds data, and stored procedures

-- 1. Create the database
DROP DATABASE IF EXISTS neormgo_test;
CREATE DATABASE neormgo_test;

-- Connect to the database (you'll need to run \c neormgo_test; in psql after creating the database)
\c neormgo_test;

DROP TABLE IF EXISTS users CASCADE;
DROP PROCEDURE IF EXISTS get_mock_result;

-- 2. Create users table with name, age, height columns
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    age INTEGER NOT NULL,
    height DECIMAL(5,2) NOT NULL -- Height in centimeters
);

-- 3. Seed 5 users to the table
INSERT INTO users (name, age, height) VALUES
    ('John Doe', 25, 175.50),
    ('Jane Smith', 30, 162.30),
    ('Mike Johnson', 28, 180.75),
    ('Sarah Wilson', 22, 168.20),
    ('David Brown', 35, 172.80);

CREATE OR REPLACE FUNCTION get_mock_result()
RETURNS TABLE(result int) AS $$
BEGIN
    RETURN QUERY
    SELECT 1 AS result;
END;
$$ LANGUAGE plpgsql;