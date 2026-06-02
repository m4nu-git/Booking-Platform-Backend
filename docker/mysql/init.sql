-- Creates all four service databases on first MySQL startup.
-- This file is automatically executed by the MySQL Docker image
-- when the data directory is empty (fresh container).

CREATE DATABASE IF NOT EXISTS `auth_dev`          CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE DATABASE IF NOT EXISTS `airbnb_dev`         CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE DATABASE IF NOT EXISTS `airbnb_booking_dev` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE DATABASE IF NOT EXISTS `review_dev`         CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- Grant the root user remote access to all four databases
-- (the MySQL image creates root with the password set in MYSQL_ROOT_PASSWORD)
GRANT ALL PRIVILEGES ON `auth_dev`.*          TO 'root'@'%';
GRANT ALL PRIVILEGES ON `airbnb_dev`.*         TO 'root'@'%';
GRANT ALL PRIVILEGES ON `airbnb_booking_dev`.* TO 'root'@'%';
GRANT ALL PRIVILEGES ON `review_dev`.*         TO 'root'@'%';
FLUSH PRIVILEGES;
