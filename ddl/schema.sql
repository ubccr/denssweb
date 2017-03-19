DROP TABLE IF EXISTS `job`;
CREATE TABLE `job` (
    `id`             int(11)           NOT NULL AUTO_INCREMENT,
    `status_id`      int(11)           NOT NULL,
    `input_data`     blob              NOT NULL,
    `dmax`           tinyint unsigned  NOT NULL,
    `submitted`      datetime          NULL,
    `started`        datetime          NULL,
    `completed`      datetime          NULL,
    PRIMARY KEY      (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

DROP TABLE IF EXISTS `job_status`;
CREATE TABLE `job_status` (
    `id`             int(11)           NOT NULL AUTO_INCREMENT,
    `status`         varchar(255)      NOT NULL,
    PRIMARY KEY      (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

INSERT INTO job_status SET id = 1, status = "Pending";
INSERT INTO job_status SET id = 2, status = "Running";
INSERT INTO job_status SET id = 3, status = "Complete";
INSERT INTO job_status SET id = 4, status = "Error";
