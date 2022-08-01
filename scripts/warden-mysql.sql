CREATE TABLE IF NOT EXISTS `attachments` (
  `id` char(36) NOT NULL,
  `cipher_uuid` char(36) NOT NULL,
  `file_name` text NOT NULL,
  `file_size` int(11) NOT NULL,
  `akey` text,
  PRIMARY KEY (`id`)
);

CREATE TABLE IF NOT EXISTS `ciphers` (
  `uuid` char(36) NOT NULL,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  `user_uuid` char(36) DEFAULT NULL,
  `organization_uuid` char(36) DEFAULT NULL,
  `atype` int(11) NOT NULL,
  `name` text NOT NULL,
  `notes` text,
  `fields` text,
  `data` text NOT NULL,
  `password_history` text,
  `deleted_at` datetime DEFAULT NULL,
  `reprompt` int(11) DEFAULT NULL,
  PRIMARY KEY (`uuid`),
  KEY `user_atype_updated` (`user_uuid`,`atype`,`updated_at`)
);

CREATE TABLE IF NOT EXISTS `ciphers_collections` (
  `cipher_uuid` char(36) NOT NULL,
  `collection_uuid` char(36) NOT NULL,
  PRIMARY KEY (`cipher_uuid`,`collection_uuid`)
);

CREATE TABLE IF NOT EXISTS `collections` (
  `uuid` varchar(40) NOT NULL,
  `org_uuid` varchar(40) NOT NULL,
  `name` text NOT NULL,
  PRIMARY KEY (`uuid`)
);

CREATE TABLE IF NOT EXISTS `devices` (
  `uuid` char(36) NOT NULL,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  `user_uuid` char(36) NOT NULL,
  `name` text NOT NULL,
  `atype` int(11) NOT NULL,
  `push_token` text,
  `refresh_token` text NOT NULL,
  `twofactor_remember` text,
  PRIMARY KEY (`uuid`)
);

CREATE TABLE IF NOT EXISTS `emergency_access` (
  `uuid` char(36) NOT NULL,
  `grantor_uuid` char(36) DEFAULT NULL,
  `grantee_uuid` char(36) DEFAULT NULL,
  `email` varchar(255) DEFAULT NULL,
  `key_encrypted` text,
  `atype` int(11) NOT NULL,
  `status` int(11) NOT NULL,
  `wait_time_days` int(11) NOT NULL,
  `recovery_initiated_at` datetime DEFAULT NULL,
  `last_notification_at` datetime DEFAULT NULL,
  `updated_at` datetime NOT NULL,
  `created_at` datetime NOT NULL,
  PRIMARY KEY (`uuid`)
);

CREATE TABLE IF NOT EXISTS `favorites` (
  `user_uuid` char(36) NOT NULL,
  `cipher_uuid` char(36) NOT NULL,
  PRIMARY KEY (`user_uuid`,`cipher_uuid`)
);

CREATE TABLE IF NOT EXISTS `folders` (
  `uuid` char(36) NOT NULL,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  `user_uuid` char(36) NOT NULL,
  `name` text NOT NULL,
  PRIMARY KEY (`uuid`)
);

CREATE TABLE IF NOT EXISTS `folders_ciphers` (
  `cipher_uuid` char(36) NOT NULL,
  `folder_uuid` char(36) NOT NULL,
  PRIMARY KEY (`cipher_uuid`,`folder_uuid`)
);

CREATE TABLE IF NOT EXISTS `invitations` (
  `email` varchar(255) NOT NULL,
  PRIMARY KEY (`email`)
);

CREATE TABLE IF NOT EXISTS `organizations` (
  `uuid` varchar(40) NOT NULL,
  `name` text NOT NULL,
  `billing_email` text NOT NULL,
  `private_key` text,
  `public_key` text,
  PRIMARY KEY (`uuid`)
);

CREATE TABLE IF NOT EXISTS `org_policies` (
  `uuid` char(36) NOT NULL,
  `org_uuid` char(36) NOT NULL,
  `atype` int(11) NOT NULL,
  `enabled` tinyint(1) NOT NULL,
  `data` text NOT NULL,
  PRIMARY KEY (`uuid`),
  UNIQUE KEY `org_uuid` (`org_uuid`,`atype`)
);

CREATE TABLE IF NOT EXISTS `sends` (
  `uuid` char(36) NOT NULL,
  `user_uuid` char(36) DEFAULT NULL,
  `organization_uuid` char(36) DEFAULT NULL,
  `name` text NOT NULL,
  `notes` text,
  `atype` int(11) NOT NULL,
  `data` text NOT NULL,
  `akey` text NOT NULL,
  `password_hash` blob,
  `password_salt` blob,
  `password_iter` int(11) DEFAULT NULL,
  `max_access_count` int(11) DEFAULT NULL,
  `access_count` int(11) NOT NULL,
  `creation_date` datetime NOT NULL,
  `revision_date` datetime NOT NULL,
  `expiration_date` datetime DEFAULT NULL,
  `deletion_date` datetime NOT NULL,
  `disabled` tinyint(1) NOT NULL,
  `hide_email` tinyint(1) DEFAULT NULL,
  PRIMARY KEY (`uuid`)
);

CREATE TABLE IF NOT EXISTS `twofactor` (
  `uuid` char(36) NOT NULL,
  `user_uuid` char(36) NOT NULL,
  `atype` int(11) NOT NULL,
  `enabled` tinyint(1) NOT NULL,
  `data` text NOT NULL,
  `last_used` int(11) NOT NULL DEFAULT '0',
  PRIMARY KEY (`uuid`),
  UNIQUE KEY `user_uuid` (`user_uuid`,`atype`)
);

CREATE TABLE IF NOT EXISTS `twofactor_incomplete` (
  `user_uuid` char(36) NOT NULL,
  `device_uuid` char(36) NOT NULL,
  `device_name` text NOT NULL,
  `login_time` datetime NOT NULL,
  `ip_address` text NOT NULL,
  PRIMARY KEY (`user_uuid`,`device_uuid`)
);

CREATE TABLE IF NOT EXISTS `users` (
  `uuid` char(36) NOT NULL,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  `email` varchar(255) NOT NULL,
  `name` text NOT NULL,
  `password_hash` blob NOT NULL,
  `salt` blob NOT NULL,
  `password_iterations` int(11) NOT NULL,
  `password_hint` text,
  `akey` text,
  `private_key` text,
  `public_key` text,
  `totp_secret` text,
  `totp_recover` text,
  `security_stamp` text NOT NULL,
  `equivalent_domains` text NOT NULL,
  `excluded_globals` text NOT NULL,
  `client_kdf_type` int(11) NOT NULL DEFAULT '0',
  `client_kdf_iter` int(11) NOT NULL DEFAULT '100000',
  `verified_at` datetime DEFAULT NULL,
  `last_verifying_at` datetime DEFAULT NULL,
  `login_verify_count` int(11) NOT NULL DEFAULT '0',
  `email_new` varchar(255) DEFAULT NULL,
  `email_new_token` varchar(16) DEFAULT NULL,
  `enabled` tinyint(1) NOT NULL DEFAULT '1',
  `stamp_exception` text,
  `api_key` varchar(255) DEFAULT NULL,
  PRIMARY KEY (`uuid`),
  UNIQUE KEY `email` (`email`)
);

CREATE TABLE IF NOT EXISTS `users_collections` (
  `user_uuid` char(36) NOT NULL,
  `collection_uuid` char(36) NOT NULL,
  `read_only` tinyint(1) NOT NULL DEFAULT '0',
  `hide_passwords` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`user_uuid`,`collection_uuid`)
);

CREATE TABLE IF NOT EXISTS `users_organizations` (
  `uuid` char(36) NOT NULL,
  `user_uuid` char(36) NOT NULL,
  `org_uuid` char(36) NOT NULL,
  `access_all` tinyint(1) NOT NULL,
  `akey` text,
  `status` int(11) NOT NULL,
  `atype` int(11) NOT NULL,
  PRIMARY KEY (`uuid`),
  UNIQUE KEY `user_uuid` (`user_uuid`,`org_uuid`)
);
