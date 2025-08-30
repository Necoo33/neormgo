-- MySQL Migration File for neormgo_test database
-- This file creates the database, tables, seeds data, and stored procedures

-- 1. Create the database
CREATE DATABASE IF NOT EXISTS `neormgo_test`;
USE `neormgo_test`;

-- 2. Create users table with name, age, height columns
CREATE TABLE IF NOT EXISTS `users` (
    `id` INT AUTO_INCREMENT PRIMARY KEY,
    `name` VARCHAR(255) NOT NULL,
    `age` INT NOT NULL,
    `height` DECIMAL(5,2) NOT NULL COMMENT 'Height in centimeters'
);

-- 3. Seed 5 users to the table
INSERT INTO `users` (`name`, `age`, `height`) VALUES
    ('John Doe', 25, 175.50),
    ('Jane Smith', 30, 162.30),
    ('Mike Johnson', 28, 180.75),
    ('Sarah Wilson', 22, 168.20),
    ('David Brown', 35, 172.80);

-- 4. Create a mock stored procedure that returns 1 with @result variable
DELIMITER //

CREATE PROCEDURE GetMockResult()
BEGIN
    SET @result = 1;
    SELECT @result AS result;
END //

DELIMITER ;
