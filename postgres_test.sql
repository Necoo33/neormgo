-- PostgreSQL Migration File for neormgo_test database
-- This file creates the database, tables, seeds data, and stored procedures

-- 1. Create the database
CREATE DATABASE neormgo_test;

-- Connect to the database (you'll need to run \c neormgo_test; in psql after creating the database)
-- \c neormgo_test;

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

-- 4. Create a mock stored procedure (function in PostgreSQL) that returns 1
CREATE OR REPLACE FUNCTION get_mock_result()
RETURNS TABLE(result INTEGER) AS $$
DECLARE
    mock_result INTEGER := 1;
BEGIN
    RETURN QUERY SELECT mock_result;
END;
$$ LANGUAGE plpgsql;

-- Alternative version that sets a session variable (closest to MySQL @result)
CREATE OR REPLACE FUNCTION set_mock_result()
RETURNS INTEGER AS $$
BEGIN
    -- Set a custom setting (session variable equivalent)
    PERFORM set_config('custom.result', '1', false);
    RETURN 1;
END;
$$ LANGUAGE plpgsql;

-- Function to get the session variable
CREATE OR REPLACE FUNCTION get_session_result()
RETURNS INTEGER AS $$
BEGIN
    RETURN COALESCE(current_setting('custom.result', true)::INTEGER, 0);
END;
$$ LANGUAGE plpgsql;
