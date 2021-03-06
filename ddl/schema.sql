DROP TABLE IF EXISTS `job`;
CREATE TABLE `job` (
    `id`               int(11)           NOT NULL AUTO_INCREMENT,
    `status_id`        int(11)           NOT NULL,
    `name`             varchar(255)      NOT NULL,
    `task`             varchar(255)      NOT NULL,
    `percent_complete` int(11)           NOT NULL,
    `log_message`      mediumtext        NOT NULL,
    `token`            varchar(255)      NOT NULL,
    `email`            varchar(255)      NOT NULL,
    `file_type`        char(3)           NOT NULL,
    `input_data`       longblob          NOT NULL,
    `density_map`      mediumblob        NULL,
    `fsc_chart`        mediumblob        NULL,
    `summary_chart`    mediumblob        NULL,
    `raw_data`         longblob          NULL,
    `dmax`             float             NOT NULL,
    `num_samples`      int(11)           NOT NULL,
    `oversampling`     float             NOT NULL,
    `voxel_size`       float             NOT NULL,
    `electrons`        int(11)           NOT NULL,
    `max_steps`        int(11)           NOT NULL,
    `max_runs`         int(11)           NOT NULL,
    `params`           longtext          NOT NULL,
    `submitted`        datetime          NULL,
    `started`          datetime          NULL,
    `completed`        datetime          NULL,
    PRIMARY KEY        (`id`),
    UNIQUE             (`token`)
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
