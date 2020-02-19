alter table `job` add column if not exists `params` longtext not null after `max_runs`;
