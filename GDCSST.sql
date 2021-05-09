DROP TABLE IF EXISTS `scenarios`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `scenarios` (
  `instance_id` varchar(21) DEFAULT NULL,
  `players` int(3) DEFAULT NULL,
  `mission_name` varchar(255) DEFAULT NULL,
  `mission_time` int(11) DEFAULT NULL,
  `mission_time_formatted` varchar(21) DEFAULT NULL,
  `start` int(11) DEFAULT '0',
  `end` int(11) DEFAULT '0',
  KEY `index_instance_id` (`instance_id`,`start`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;


DROP TABLE IF EXISTS `scenarios_log`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `scenarios_log` (
  `instance_id` varchar(21) DEFAULT NULL,
  `timestamp` int(11) DEFAULT NULL,
  `players` int(3) DEFAULT NULL,
  `players_max` int(3) DEFAULT NULL,
  `mission_name` varchar(255) DEFAULT NULL,
  `mission_time` int(11) DEFAULT NULL,
  `mission_time_formatted` varchar(21) DEFAULT NULL,
  KEY `index_instance_id` (`instance_id`),
  KEY `index_players` (`players`),
  KEY `index_mission_name` (`mission_name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;


DROP TABLE IF EXISTS `servers`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `servers` (
  `INSTANCE_ID` varchar(21) DEFAULT NULL,
  `timestamp` int(11) DEFAULT NULL,
  `status` varchar(4) DEFAULT NULL,
  `NAME` varchar(255) DEFAULT NULL,
  `DESCRIPTION` text,
  `IP_ADDRESS` varchar(16) DEFAULT NULL,
  `PORT` varchar(5) DEFAULT NULL,
  `PASSWORD` varchar(3) DEFAULT NULL,
  `PLAYERS` int(3) DEFAULT NULL,
  `PLAYERS_MAX` int(3) DEFAULT NULL,
  `MISSION_NAME` varchar(255) DEFAULT NULL,
  `MISSION_TIME` int(11) DEFAULT NULL,
  `MISSION_TIME_FORMATTED` varchar(21) DEFAULT NULL,
  `country_iso` varchar(7) DEFAULT NULL,
  `country_name` varchar(100) DEFAULT NULL,
  KEY `index_status` (`status`),
  KEY `index_players` (`PLAYERS`),
  KEY `index_name` (`NAME`),
  KEY `index_mission_name` (`MISSION_NAME`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
