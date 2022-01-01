CREATE TABLE `access_tokens` (
  `access_token` varchar(255) PRIMARY KEY NOT NULL,
  `user_id` bigint NOT NULL,
  `user_role` varchar(255) NOT NULL,
  `expires` bigint NOT NULL
);

CREATE INDEX `access_tokens_index_0` ON `access_tokens` (`access_token`);

CREATE INDEX `access_tokens_index_1` ON `access_tokens` (`user_id`);

CREATE TABLE `users` (
  `id` bigint NOT NULL PRIMARY KEY,
  `email` varchar(255) NOT NULL,
  `password` varchar(255) NOT NULL, 
  `role` varchar(255) NOT NULL
)