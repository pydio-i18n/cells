-- MySQL dump 10.13  Distrib 8.0.21, for osx10.13 (x86_64)
--
-- Host: localhost    Database: cells
-- ------------------------------------------------------
-- Server version	8.0.21

/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!50503 SET NAMES utf8mb4 */;
/*!40103 SET @OLD_TIME_ZONE=@@TIME_ZONE */;
/*!40103 SET TIME_ZONE='+00:00' */;
/*!40014 SET @OLD_UNIQUE_CHECKS=@@UNIQUE_CHECKS, UNIQUE_CHECKS=0 */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;

--
-- Table structure for table `gorp_migrations`
--

DROP TABLE IF EXISTS `gorp_migrations`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `gorp_migrations` (
  `id` varchar(255) NOT NULL,
  `applied_at` datetime DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `hydra_client`
--

DROP TABLE IF EXISTS `hydra_client`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `hydra_client` (
  `id` varchar(255) NOT NULL,
  `client_name` text NOT NULL,
  `client_secret` text NOT NULL,
  `redirect_uris` text NOT NULL,
  `grant_types` text NOT NULL,
  `response_types` text NOT NULL,
  `scope` text NOT NULL,
  `owner` text NOT NULL,
  `policy_uri` text NOT NULL,
  `tos_uri` text NOT NULL,
  `client_uri` text NOT NULL,
  `logo_uri` text NOT NULL,
  `contacts` text NOT NULL,
  `client_secret_expires_at` int NOT NULL DEFAULT '0',
  `sector_identifier_uri` text NOT NULL,
  `jwks` text NOT NULL,
  `jwks_uri` text NOT NULL,
  `request_uris` text NOT NULL,
  `token_endpoint_auth_method` varchar(25) NOT NULL DEFAULT '',
  `request_object_signing_alg` varchar(10) NOT NULL DEFAULT '',
  `userinfo_signed_response_alg` varchar(10) NOT NULL DEFAULT '',
  `subject_type` varchar(15) NOT NULL DEFAULT '',
  `allowed_cors_origins` text NOT NULL,
  `pk` int unsigned NOT NULL AUTO_INCREMENT,
  `audience` text NOT NULL,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `frontchannel_logout_uri` text NOT NULL,
  `frontchannel_logout_session_required` tinyint(1) NOT NULL DEFAULT '0',
  `post_logout_redirect_uris` text NOT NULL,
  `backchannel_logout_uri` text NOT NULL,
  `backchannel_logout_session_required` tinyint(1) NOT NULL DEFAULT '0',
  `metadata` text NOT NULL,
  `token_endpoint_auth_signing_alg` varchar(10) NOT NULL DEFAULT '',
  PRIMARY KEY (`pk`),
  UNIQUE KEY `hydra_client_idx_id_uq` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=5 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `hydra_jwk`
--

DROP TABLE IF EXISTS `hydra_jwk`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `hydra_jwk` (
  `sid` varchar(255) NOT NULL,
  `kid` varchar(255) NOT NULL,
  `version` int NOT NULL DEFAULT '0',
  `keydata` text NOT NULL,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `pk` int unsigned NOT NULL AUTO_INCREMENT,
  PRIMARY KEY (`pk`),
  UNIQUE KEY `hydra_jwk_idx_id_uq` (`sid`,`kid`)
) ENGINE=InnoDB AUTO_INCREMENT=3 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `hydra_oauth2_access`
--

DROP TABLE IF EXISTS `hydra_oauth2_access`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `hydra_oauth2_access` (
  `signature` varchar(255) NOT NULL,
  `request_id` varchar(40) NOT NULL DEFAULT '',
  `requested_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `client_id` varchar(255) NOT NULL DEFAULT '',
  `scope` text NOT NULL,
  `granted_scope` text NOT NULL,
  `form_data` text NOT NULL,
  `session_data` text NOT NULL,
  `subject` varchar(255) NOT NULL DEFAULT '',
  `active` tinyint(1) NOT NULL DEFAULT '1',
  `requested_audience` text NOT NULL,
  `granted_audience` text NOT NULL,
  `challenge_id` varchar(40) DEFAULT NULL,
  PRIMARY KEY (`signature`),
  KEY `hydra_oauth2_access_requested_at_idx` (`requested_at`),
  KEY `hydra_oauth2_access_client_id_idx` (`client_id`),
  KEY `hydra_oauth2_access_challenge_id_idx` (`challenge_id`),
  KEY `hydra_oauth2_access_client_id_subject_idx` (`client_id`,`subject`),
  KEY `hydra_oauth2_access_request_id_idx` (`request_id`),
  CONSTRAINT `hydra_oauth2_access_challenge_id_fk` FOREIGN KEY (`challenge_id`) REFERENCES `hydra_oauth2_consent_request_handled` (`challenge`) ON DELETE CASCADE,
  CONSTRAINT `hydra_oauth2_access_client_id_fk` FOREIGN KEY (`client_id`) REFERENCES `hydra_client` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `hydra_oauth2_authentication_request`
--

DROP TABLE IF EXISTS `hydra_oauth2_authentication_request`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `hydra_oauth2_authentication_request` (
  `challenge` varchar(40) NOT NULL,
  `requested_scope` text NOT NULL,
  `verifier` varchar(40) NOT NULL,
  `csrf` varchar(40) NOT NULL,
  `subject` varchar(255) NOT NULL,
  `request_url` text NOT NULL,
  `skip` tinyint(1) NOT NULL,
  `client_id` varchar(255) NOT NULL,
  `requested_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `authenticated_at` timestamp NULL DEFAULT NULL,
  `oidc_context` text NOT NULL,
  `login_session_id` varchar(40),
  `requested_at_audience` text NOT NULL,
  PRIMARY KEY (`challenge`),
  UNIQUE KEY `hydra_oauth2_authentication_request_veri_idx` (`verifier`),
  KEY `hydra_oauth2_authentication_request_cid_idx` (`client_id`),
  KEY `hydra_oauth2_authentication_request_sub_idx` (`subject`),
  KEY `hydra_oauth2_authentication_request_login_session_id_idx` (`login_session_id`),
  CONSTRAINT `hydra_oauth2_authentication_request_client_id_fk` FOREIGN KEY (`client_id`) REFERENCES `hydra_client` (`id`) ON DELETE CASCADE,
  CONSTRAINT `hydra_oauth2_authentication_request_login_session_id_fk` FOREIGN KEY (`login_session_id`) REFERENCES `hydra_oauth2_authentication_session` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `hydra_oauth2_authentication_request_handled`
--

DROP TABLE IF EXISTS `hydra_oauth2_authentication_request_handled`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `hydra_oauth2_authentication_request_handled` (
  `challenge` varchar(40) NOT NULL,
  `subject` varchar(255) NOT NULL,
  `remember` tinyint(1) NOT NULL,
  `remember_for` int NOT NULL,
  `error` text NOT NULL,
  `acr` text NOT NULL,
  `requested_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `authenticated_at` timestamp NULL DEFAULT NULL,
  `was_used` tinyint(1) NOT NULL,
  `forced_subject_identifier` varchar(255) DEFAULT '',
  `context` text NOT NULL,
  `amr` text NOT NULL,
  PRIMARY KEY (`challenge`),
  CONSTRAINT `hydra_oauth2_authentication_request_handled_challenge_fk` FOREIGN KEY (`challenge`) REFERENCES `hydra_oauth2_authentication_request` (`challenge`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `hydra_oauth2_authentication_session`
--

DROP TABLE IF EXISTS `hydra_oauth2_authentication_session`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `hydra_oauth2_authentication_session` (
  `id` varchar(40) NOT NULL,
  `authenticated_at` timestamp NULL DEFAULT NULL,
  `subject` varchar(255) NOT NULL,
  `remember` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  KEY `hydra_oauth2_authentication_session_sub_idx` (`subject`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `hydra_oauth2_code`
--

DROP TABLE IF EXISTS `hydra_oauth2_code`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `hydra_oauth2_code` (
  `signature` varchar(255) NOT NULL,
  `request_id` varchar(40) NOT NULL DEFAULT '',
  `requested_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `client_id` varchar(255) NOT NULL DEFAULT '',
  `scope` text NOT NULL,
  `granted_scope` text NOT NULL,
  `form_data` text NOT NULL,
  `session_data` text NOT NULL,
  `subject` varchar(255) NOT NULL DEFAULT '',
  `active` tinyint(1) NOT NULL DEFAULT '1',
  `requested_audience` text NOT NULL,
  `granted_audience` text NOT NULL,
  `challenge_id` varchar(40) DEFAULT NULL,
  PRIMARY KEY (`signature`),
  KEY `hydra_oauth2_code_client_id_idx` (`client_id`),
  KEY `hydra_oauth2_code_challenge_id_idx` (`challenge_id`),
  KEY `hydra_oauth2_code_request_id_idx` (`request_id`),
  CONSTRAINT `hydra_oauth2_code_challenge_id_fk` FOREIGN KEY (`challenge_id`) REFERENCES `hydra_oauth2_consent_request_handled` (`challenge`) ON DELETE CASCADE,
  CONSTRAINT `hydra_oauth2_code_client_id_fk` FOREIGN KEY (`client_id`) REFERENCES `hydra_client` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `hydra_oauth2_consent_request`
--

DROP TABLE IF EXISTS `hydra_oauth2_consent_request`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `hydra_oauth2_consent_request` (
  `challenge` varchar(40) NOT NULL,
  `verifier` varchar(40) NOT NULL,
  `client_id` varchar(255) NOT NULL,
  `subject` varchar(255) NOT NULL,
  `request_url` text NOT NULL,
  `skip` tinyint(1) NOT NULL,
  `requested_scope` text NOT NULL,
  `csrf` varchar(40) NOT NULL,
  `authenticated_at` timestamp NULL DEFAULT NULL,
  `requested_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `oidc_context` text NOT NULL,
  `forced_subject_identifier` varchar(255) DEFAULT '',
  `login_session_id` varchar(40),
  `login_challenge` varchar(40),
  `requested_at_audience` text NOT NULL,
  `acr` text NOT NULL,
  `context` text NOT NULL,
  `amr` text NOT NULL,
  PRIMARY KEY (`challenge`),
  UNIQUE KEY `hydra_oauth2_consent_request_veri_idx` (`verifier`),
  KEY `hydra_oauth2_consent_request_cid_idx` (`client_id`),
  KEY `hydra_oauth2_consent_request_sub_idx` (`subject`),
  KEY `hydra_oauth2_consent_request_login_session_id_idx` (`login_session_id`),
  KEY `hydra_oauth2_consent_request_login_challenge_idx` (`login_challenge`),
  KEY `hydra_oauth2_consent_request_client_id_subject_idx` (`client_id`,`subject`),
  CONSTRAINT `hydra_oauth2_consent_request_client_id_fk` FOREIGN KEY (`client_id`) REFERENCES `hydra_client` (`id`) ON DELETE CASCADE,
  CONSTRAINT `hydra_oauth2_consent_request_login_challenge_fk` FOREIGN KEY (`login_challenge`) REFERENCES `hydra_oauth2_authentication_request` (`challenge`) ON DELETE SET NULL,
  CONSTRAINT `hydra_oauth2_consent_request_login_session_id_fk` FOREIGN KEY (`login_session_id`) REFERENCES `hydra_oauth2_authentication_session` (`id`) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `hydra_oauth2_consent_request_handled`
--

DROP TABLE IF EXISTS `hydra_oauth2_consent_request_handled`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `hydra_oauth2_consent_request_handled` (
  `challenge` varchar(40) NOT NULL,
  `granted_scope` text NOT NULL,
  `remember` tinyint(1) NOT NULL,
  `remember_for` int NOT NULL,
  `error` text NOT NULL,
  `requested_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `session_access_token` text NOT NULL,
  `session_id_token` text NOT NULL,
  `authenticated_at` timestamp NULL DEFAULT NULL,
  `was_used` tinyint(1) NOT NULL,
  `granted_at_audience` text NOT NULL,
  `handled_at` timestamp NULL DEFAULT NULL,
  PRIMARY KEY (`challenge`),
  CONSTRAINT `hydra_oauth2_consent_request_handled_challenge_fk` FOREIGN KEY (`challenge`) REFERENCES `hydra_oauth2_consent_request` (`challenge`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `hydra_oauth2_jti_blacklist`
--

DROP TABLE IF EXISTS `hydra_oauth2_jti_blacklist`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `hydra_oauth2_jti_blacklist` (
  `signature` varchar(64) NOT NULL,
  `expires_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`signature`),
  KEY `hydra_oauth2_jti_blacklist_expiry` (`expires_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `hydra_oauth2_logout_request`
--

DROP TABLE IF EXISTS `hydra_oauth2_logout_request`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `hydra_oauth2_logout_request` (
  `challenge` varchar(36) NOT NULL,
  `verifier` varchar(36) NOT NULL,
  `subject` varchar(255) NOT NULL,
  `sid` varchar(36) NOT NULL,
  `client_id` varchar(255) DEFAULT NULL,
  `request_url` text NOT NULL,
  `redir_url` text NOT NULL,
  `was_used` tinyint(1) NOT NULL DEFAULT '0',
  `accepted` tinyint(1) NOT NULL DEFAULT '0',
  `rejected` tinyint(1) NOT NULL DEFAULT '0',
  `rp_initiated` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`challenge`),
  UNIQUE KEY `hydra_oauth2_logout_request_veri_idx` (`verifier`),
  KEY `hydra_oauth2_logout_request_client_id_idx` (`client_id`),
  CONSTRAINT `hydra_oauth2_logout_request_client_id_fk` FOREIGN KEY (`client_id`) REFERENCES `hydra_client` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `hydra_oauth2_obfuscated_authentication_session`
--

DROP TABLE IF EXISTS `hydra_oauth2_obfuscated_authentication_session`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `hydra_oauth2_obfuscated_authentication_session` (
  `subject` varchar(255) NOT NULL,
  `client_id` varchar(255) NOT NULL,
  `subject_obfuscated` varchar(255) NOT NULL,
  PRIMARY KEY (`subject`,`client_id`),
  UNIQUE KEY `hydra_oauth2_obfuscated_authentication_session_so_idx` (`client_id`,`subject_obfuscated`),
  CONSTRAINT `hydra_oauth2_obfuscated_authentication_session_client_id_fk` FOREIGN KEY (`client_id`) REFERENCES `hydra_client` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `hydra_oauth2_oidc`
--

DROP TABLE IF EXISTS `hydra_oauth2_oidc`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `hydra_oauth2_oidc` (
  `signature` varchar(255) NOT NULL,
  `request_id` varchar(40) NOT NULL DEFAULT '',
  `requested_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `client_id` varchar(255) NOT NULL DEFAULT '',
  `scope` text NOT NULL,
  `granted_scope` text NOT NULL,
  `form_data` text NOT NULL,
  `session_data` text NOT NULL,
  `subject` varchar(255) NOT NULL DEFAULT '',
  `active` tinyint(1) NOT NULL DEFAULT '1',
  `requested_audience` text NOT NULL,
  `granted_audience` text NOT NULL,
  `challenge_id` varchar(40) DEFAULT NULL,
  PRIMARY KEY (`signature`),
  KEY `hydra_oauth2_oidc_client_id_idx` (`client_id`),
  KEY `hydra_oauth2_oidc_challenge_id_idx` (`challenge_id`),
  KEY `hydra_oauth2_oidc_request_id_idx` (`request_id`),
  CONSTRAINT `hydra_oauth2_oidc_challenge_id_fk` FOREIGN KEY (`challenge_id`) REFERENCES `hydra_oauth2_consent_request_handled` (`challenge`) ON DELETE CASCADE,
  CONSTRAINT `hydra_oauth2_oidc_client_id_fk` FOREIGN KEY (`client_id`) REFERENCES `hydra_client` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `hydra_oauth2_pkce`
--

DROP TABLE IF EXISTS `hydra_oauth2_pkce`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `hydra_oauth2_pkce` (
  `signature` varchar(255) NOT NULL,
  `request_id` varchar(40) NOT NULL DEFAULT '',
  `requested_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `client_id` varchar(255) NOT NULL DEFAULT '',
  `scope` text NOT NULL,
  `granted_scope` text NOT NULL,
  `form_data` text NOT NULL,
  `session_data` text NOT NULL,
  `subject` varchar(255) NOT NULL,
  `active` tinyint(1) NOT NULL DEFAULT '1',
  `requested_audience` text NOT NULL,
  `granted_audience` text NOT NULL,
  `challenge_id` varchar(40) DEFAULT NULL,
  PRIMARY KEY (`signature`),
  KEY `hydra_oauth2_pkce_client_id_idx` (`client_id`),
  KEY `hydra_oauth2_pkce_challenge_id_idx` (`challenge_id`),
  KEY `hydra_oauth2_pkce_request_id_idx` (`request_id`),
  CONSTRAINT `hydra_oauth2_pkce_challenge_id_fk` FOREIGN KEY (`challenge_id`) REFERENCES `hydra_oauth2_consent_request_handled` (`challenge`) ON DELETE CASCADE,
  CONSTRAINT `hydra_oauth2_pkce_client_id_fk` FOREIGN KEY (`client_id`) REFERENCES `hydra_client` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `hydra_oauth2_refresh`
--

DROP TABLE IF EXISTS `hydra_oauth2_refresh`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `hydra_oauth2_refresh` (
  `signature` varchar(255) NOT NULL,
  `request_id` varchar(40) NOT NULL DEFAULT '',
  `requested_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `client_id` varchar(255) NOT NULL DEFAULT '',
  `scope` text NOT NULL,
  `granted_scope` text NOT NULL,
  `form_data` text NOT NULL,
  `session_data` text NOT NULL,
  `subject` varchar(255) NOT NULL DEFAULT '',
  `active` tinyint(1) NOT NULL DEFAULT '1',
  `requested_audience` text NOT NULL,
  `granted_audience` text NOT NULL,
  `challenge_id` varchar(40) DEFAULT NULL,
  PRIMARY KEY (`signature`),
  KEY `hydra_oauth2_refresh_client_id_idx` (`client_id`),
  KEY `hydra_oauth2_refresh_challenge_id_idx` (`challenge_id`),
  KEY `hydra_oauth2_refresh_client_id_subject_idx` (`client_id`,`subject`),
  KEY `hydra_oauth2_refresh_request_id_idx` (`request_id`),
  CONSTRAINT `hydra_oauth2_refresh_challenge_id_fk` FOREIGN KEY (`challenge_id`) REFERENCES `hydra_oauth2_consent_request_handled` (`challenge`) ON DELETE CASCADE,
  CONSTRAINT `hydra_oauth2_refresh_client_id_fk` FOREIGN KEY (`client_id`) REFERENCES `hydra_client` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `idm_personal_tokens`
--

DROP TABLE IF EXISTS `idm_personal_tokens`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `idm_personal_tokens` (
  `uuid` varchar(36) NOT NULL,
  `access_token` varchar(128) NOT NULL,
  `pat_type` int DEFAULT NULL,
  `label` varchar(255) DEFAULT NULL,
  `user_uuid` varchar(255) NOT NULL,
  `user_login` varchar(255) NOT NULL,
  `auto_refresh` int DEFAULT '0',
  `expire_at` int DEFAULT NULL,
  `created_at` int DEFAULT NULL,
  `created_by` varchar(128) DEFAULT NULL,
  `updated_at` int DEFAULT NULL,
  `scopes` longtext,
  PRIMARY KEY (`uuid`),
  UNIQUE KEY `pat_unique_access_token_key` (`access_token`),
  KEY `pat_user_uuid_key` (`user_uuid`),
  KEY `pat_user_login_key` (`user_login`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `idm_usr_meta`
--

DROP TABLE IF EXISTS `idm_usr_meta`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `idm_usr_meta` (
  `uuid` varchar(255) NOT NULL,
  `node_uuid` varchar(255) NOT NULL,
  `namespace` varchar(255) NOT NULL,
  `owner` varchar(255) DEFAULT NULL,
  `timestamp` int DEFAULT NULL,
  `format` varchar(50) DEFAULT NULL,
  `data` blob,
  PRIMARY KEY (`uuid`),
  UNIQUE KEY `namespace` (`namespace`,`node_uuid`,`owner`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `idm_usr_meta_ns`
--

DROP TABLE IF EXISTS `idm_usr_meta_ns`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `idm_usr_meta_ns` (
  `namespace` varchar(255) NOT NULL,
  `label` varchar(255) NOT NULL,
  `ns_order` int NOT NULL,
  `indexable` tinyint(1) DEFAULT NULL,
  `definition` blob,
  PRIMARY KEY (`namespace`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `idm_usr_meta_policies`
--

DROP TABLE IF EXISTS `idm_usr_meta_policies`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `idm_usr_meta_policies` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `resource` varchar(255) NOT NULL,
  `action` varchar(255) NOT NULL,
  `subject` varchar(255) NOT NULL,
  `effect` enum('allow','deny') DEFAULT 'deny',
  `conditions` varchar(500) NOT NULL DEFAULT '{}',
  PRIMARY KEY (`id`),
  KEY `resource` (`resource`),
  KEY `action` (`action`),
  KEY `subject` (`subject`)
) ENGINE=InnoDB AUTO_INCREMENT=3 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `pydio_grpc_data_index_cellsdata_idx_tree`
--

DROP TABLE IF EXISTS `pydio_grpc_data_index_cellsdata_idx_tree`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `pydio_grpc_data_index_cellsdata_idx_tree` (
  `uuid` varchar(128) NOT NULL,
  `level` smallint NOT NULL,
  `name` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NOT NULL,
  `leaf` tinyint(1) NOT NULL DEFAULT '0',
  `mtime` int NOT NULL,
  `etag` varchar(255) NOT NULL DEFAULT '',
  `size` bigint NOT NULL DEFAULT '0',
  `mode` varchar(10) NOT NULL DEFAULT '',
  `mpath1` varchar(255) NOT NULL,
  `mpath2` varchar(255) NOT NULL,
  `mpath3` varchar(255) NOT NULL,
  `mpath4` varchar(255) NOT NULL,
  `hash` varchar(40) NOT NULL,
  `hash2` varchar(50) NOT NULL,
  PRIMARY KEY (`uuid`),
  UNIQUE KEY `pydio_grpc_data_index_cellsdata_idx_tree_u1` (`hash`),
  UNIQUE KEY `pydio_grpc_data_index_cellsdata_idx_tree_u2` (`hash2`),
  KEY `pydio_grpc_data_index_cellsdata_idx_tree_mpath1_idx` (`mpath1`),
  KEY `pydio_grpc_data_index_cellsdata_idx_tree_mpath2_idx` (`mpath2`),
  KEY `pydio_grpc_data_index_cellsdata_idx_tree_mpath3_idx` (`mpath3`),
  KEY `pydio_grpc_data_index_cellsdata_idx_tree_mpath4_idx` (`mpath4`),
  KEY `pydio_grpc_data_index_cellsdata_idx_tree_name_idx` (`name`(128)),
  KEY `pydio_grpc_data_index_cellsdata_idx_tree_level_idx` (`level`)
) ENGINE=InnoDB DEFAULT CHARSET=ascii;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `pydio_grpc_data_index_personal_idx_tree`
--

DROP TABLE IF EXISTS `pydio_grpc_data_index_personal_idx_tree`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `pydio_grpc_data_index_personal_idx_tree` (
  `uuid` varchar(128) NOT NULL,
  `level` smallint NOT NULL,
  `name` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NOT NULL,
  `leaf` tinyint(1) NOT NULL DEFAULT '0',
  `mtime` int NOT NULL,
  `etag` varchar(255) NOT NULL DEFAULT '',
  `size` bigint NOT NULL DEFAULT '0',
  `mode` varchar(10) NOT NULL DEFAULT '',
  `mpath1` varchar(255) NOT NULL,
  `mpath2` varchar(255) NOT NULL,
  `mpath3` varchar(255) NOT NULL,
  `mpath4` varchar(255) NOT NULL,
  `hash` varchar(40) NOT NULL,
  `hash2` varchar(50) NOT NULL,
  PRIMARY KEY (`uuid`),
  UNIQUE KEY `pydio_grpc_data_index_personal_idx_tree_u1` (`hash`),
  UNIQUE KEY `pydio_grpc_data_index_personal_idx_tree_u2` (`hash2`),
  KEY `pydio_grpc_data_index_personal_idx_tree_mpath1_idx` (`mpath1`),
  KEY `pydio_grpc_data_index_personal_idx_tree_mpath2_idx` (`mpath2`),
  KEY `pydio_grpc_data_index_personal_idx_tree_mpath3_idx` (`mpath3`),
  KEY `pydio_grpc_data_index_personal_idx_tree_mpath4_idx` (`mpath4`),
  KEY `pydio_grpc_data_index_personal_idx_tree_name_idx` (`name`(128)),
  KEY `pydio_grpc_data_index_personal_idx_tree_level_idx` (`level`)
) ENGINE=InnoDB DEFAULT CHARSET=ascii;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `pydio_grpc_data_index_pydiods1_idx_tree`
--

DROP TABLE IF EXISTS `pydio_grpc_data_index_pydiods1_idx_tree`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `pydio_grpc_data_index_pydiods1_idx_tree` (
  `uuid` varchar(128) NOT NULL,
  `level` smallint NOT NULL,
  `name` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NOT NULL,
  `leaf` tinyint(1) NOT NULL DEFAULT '0',
  `mtime` int NOT NULL,
  `etag` varchar(255) NOT NULL DEFAULT '',
  `size` bigint NOT NULL DEFAULT '0',
  `mode` varchar(10) NOT NULL DEFAULT '',
  `mpath1` varchar(255) NOT NULL,
  `mpath2` varchar(255) NOT NULL,
  `mpath3` varchar(255) NOT NULL,
  `mpath4` varchar(255) NOT NULL,
  `hash` varchar(40) NOT NULL,
  `hash2` varchar(50) NOT NULL,
  PRIMARY KEY (`uuid`),
  UNIQUE KEY `pydio_grpc_data_index_pydiods1_idx_tree_u1` (`hash`),
  UNIQUE KEY `pydio_grpc_data_index_pydiods1_idx_tree_u2` (`hash2`),
  KEY `pydio_grpc_data_index_pydiods1_idx_tree_mpath1_idx` (`mpath1`),
  KEY `pydio_grpc_data_index_pydiods1_idx_tree_mpath2_idx` (`mpath2`),
  KEY `pydio_grpc_data_index_pydiods1_idx_tree_mpath3_idx` (`mpath3`),
  KEY `pydio_grpc_data_index_pydiods1_idx_tree_mpath4_idx` (`mpath4`),
  KEY `pydio_grpc_data_index_pydiods1_idx_tree_name_idx` (`name`(128)),
  KEY `pydio_grpc_data_index_pydiods1_idx_tree_level_idx` (`level`)
) ENGINE=InnoDB DEFAULT CHARSET=ascii;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `pydio_grpc_data_index_thumbnails_idx_tree`
--

DROP TABLE IF EXISTS `pydio_grpc_data_index_thumbnails_idx_tree`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `pydio_grpc_data_index_thumbnails_idx_tree` (
  `uuid` varchar(128) NOT NULL,
  `level` smallint NOT NULL,
  `name` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NOT NULL,
  `leaf` tinyint(1) NOT NULL DEFAULT '0',
  `mtime` int NOT NULL,
  `etag` varchar(255) NOT NULL DEFAULT '',
  `size` bigint NOT NULL DEFAULT '0',
  `mode` varchar(10) NOT NULL DEFAULT '',
  `mpath1` varchar(255) NOT NULL,
  `mpath2` varchar(255) NOT NULL,
  `mpath3` varchar(255) NOT NULL,
  `mpath4` varchar(255) NOT NULL,
  `hash` varchar(40) NOT NULL,
  `hash2` varchar(50) NOT NULL,
  PRIMARY KEY (`uuid`),
  UNIQUE KEY `pydio_grpc_data_index_thumbnails_idx_tree_u1` (`hash`),
  UNIQUE KEY `pydio_grpc_data_index_thumbnails_idx_tree_u2` (`hash2`),
  KEY `pydio_grpc_data_index_thumbnails_idx_tree_mpath1_idx` (`mpath1`),
  KEY `pydio_grpc_data_index_thumbnails_idx_tree_mpath2_idx` (`mpath2`),
  KEY `pydio_grpc_data_index_thumbnails_idx_tree_mpath3_idx` (`mpath3`),
  KEY `pydio_grpc_data_index_thumbnails_idx_tree_mpath4_idx` (`mpath4`),
  KEY `pydio_grpc_data_index_thumbnails_idx_tree_name_idx` (`name`(128)),
  KEY `pydio_grpc_data_index_thumbnails_idx_tree_level_idx` (`level`)
) ENGINE=InnoDB DEFAULT CHARSET=ascii;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `pydio_grpc_data_index_versions_idx_tree`
--

DROP TABLE IF EXISTS `pydio_grpc_data_index_versions_idx_tree`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `pydio_grpc_data_index_versions_idx_tree` (
  `uuid` varchar(128) NOT NULL,
  `level` smallint NOT NULL,
  `name` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NOT NULL,
  `leaf` tinyint(1) NOT NULL DEFAULT '0',
  `mtime` int NOT NULL,
  `etag` varchar(255) NOT NULL DEFAULT '',
  `size` bigint NOT NULL DEFAULT '0',
  `mode` varchar(10) NOT NULL DEFAULT '',
  `mpath1` varchar(255) NOT NULL,
  `mpath2` varchar(255) NOT NULL,
  `mpath3` varchar(255) NOT NULL,
  `mpath4` varchar(255) NOT NULL,
  `hash` varchar(40) NOT NULL,
  `hash2` varchar(50) NOT NULL,
  PRIMARY KEY (`uuid`),
  UNIQUE KEY `pydio_grpc_data_index_versions_idx_tree_u1` (`hash`),
  UNIQUE KEY `pydio_grpc_data_index_versions_idx_tree_u2` (`hash2`),
  KEY `pydio_grpc_data_index_versions_idx_tree_mpath1_idx` (`mpath1`),
  KEY `pydio_grpc_data_index_versions_idx_tree_mpath2_idx` (`mpath2`),
  KEY `pydio_grpc_data_index_versions_idx_tree_mpath3_idx` (`mpath3`),
  KEY `pydio_grpc_data_index_versions_idx_tree_mpath4_idx` (`mpath4`),
  KEY `pydio_grpc_data_index_versions_idx_tree_name_idx` (`name`(128)),
  KEY `pydio_grpc_data_index_versions_idx_tree_level_idx` (`level`)
) ENGINE=InnoDB DEFAULT CHARSET=ascii;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `schema_migration`
--

DROP TABLE IF EXISTS `schema_migration`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `schema_migration` (
  `version` varchar(48) NOT NULL,
  `version_self` int NOT NULL DEFAULT '0',
  UNIQUE KEY `schema_migration_version_idx` (`version`),
  KEY `schema_migration_version_self_idx` (`version_self`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40103 SET TIME_ZONE=@OLD_TIME_ZONE */;

/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40014 SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;

-- Dump completed on 2021-12-03  8:52:53
