CREATE TABLE `contest` (
  `contest_id` INT PRIMARY KEY AUTO_INCREMENT COMMENT '赛事资讯ID',
  `title` VARCHAR(255) NOT NULL,
  `image_url` VARCHAR(500),
  `field` VARCHAR(255) COMMENT '竞赛所属类别，如：工科类',
  `format` VARCHAR(255) COMMENT '竞赛形式，如团体赛',
  `description` TEXT,
  `deadline` INT,
  `fee` VARCHAR(50),
  `team_size_min` INT,
  `team_size_max` INT,
  `participant_requirements` TEXT,
  `official_website` VARCHAR(255),
  `additional_info` TEXT,
  `created_time` DATETIME
);

CREATE TABLE `contact` (
  `contact_id` INT PRIMARY KEY AUTO_INCREMENT,
  `name` VARCHAR(255),
  `phone` VARCHAR(50),
  `email` VARCHAR(100)
);

CREATE TABLE `contest_contact_relationship` (
  `contest_contact_id` INT PRIMARY KEY AUTO_INCREMENT,
  `contact_id` INT,
  `contest_id` INT
);

CREATE TABLE `user_profile_info` (
  `user_id` INT PRIMARY KEY,
  `gender` INT,
  `enrollment_year` INT,
  `mobile_phone` VARCHAR(50),
  `college` VARCHAR(255),
  `nickname` VARCHAR(255),
  `realname` VARCHAR(255),
  `avatar_url` VARCHAR(255),
  `hasProfile` BOOLEAN,
  `introduction` TEXT,
  `qq_number` VARCHAR(50),
  `wechat_number` VARCHAR(50)
);

CREATE TABLE `honors` (
  `honor_id` INT PRIMARY KEY AUTO_INCREMENT,
  `user_id` INT,
  `honor` TEXT
);

CREATE TABLE `authentication` (
  `user_id` INT PRIMARY KEY AUTO_INCREMENT,
  `username` VARCHAR(255),
  `password` VARCHAR(255)
);

CREATE TABLE `team_info` (
  `team_id` INT PRIMARY KEY AUTO_INCREMENT,
  `contest_id` INT,
  `title` VARCHAR(255),
  `goal` VARCHAR(255),
  `cur_people_num` INT COMMENT '当前队伍人数',
  `created_time` INT,
  `leader_id` INT,
  `description` LONGTEXT
);

CREATE TABLE `team_application` (
  `application_id` INT PRIMARY KEY AUTO_INCREMENT,
  `user_id` INT,
  `team_id` INT,
  `reason` TEXT,
  `created_time` DATETIME,
  `application_type` INT COMMENT '申请类型，如退出申请，加入申请'
);

CREATE TABLE `article` (
  `article_id` INT PRIMARY KEY AUTO_INCREMENT,
  `title` VARCHAR(255),
  `author_id` INT,
  `author` VARCHAR(255),
  `created_time` DATETIME,
  `link` VARCHAR(255) COMMENT '文章在学校官网的链接',
  `contest_id` INT
);

CREATE TABLE `team_user_relationship` (
  `team_user_id` INT PRIMARY KEY AUTO_INCREMENT,
  `user_id` INT,
  `team_id` INT
);

CREATE TABLE `user_favorites` (
  `favorite_id` INT PRIMARY KEY AUTO_INCREMENT,
  `user_id` INT,
  `contest_id` INT,
  `created_time` DATETIME
);

-- 用户技能表
CREATE TABLE `user_skills` (
    `user_skill_id` INT PRIMARY KEY AUTO_INCREMENT,
    `user_id` INT,
    `skill` VARCHAR(255) NOT NULL,
    `category` VARCHAR(255) NOT NULL COMMENT '技能分类，如编程语言、设计工具',
    `proficiency` ENUM ('一般', '良好', '熟练', '精通') NOT NULL
);

-- 队伍所需技能表
CREATE TABLE `team_skills` (
    `team_skill_id` INT PRIMARY KEY AUTO_INCREMENT,
    `team_id` INT,
    `skill` VARCHAR(255) NOT NULL,
    `category` VARCHAR(255) NOT NULL COMMENT '技能分类，如编程语言、设计工具',
    `job` VARCHAR(255) NOT NULL COMMENT '什么岗位需要该skill'

);

ALTER TABLE `contest` COMMENT = '存储赛事板块';

ALTER TABLE `contact` COMMENT = '存储竞赛负责人, contest的子表';

-- ALTER TABLE `contest_contact_relationship` ADD FOREIGN KEY (`contact_id`) REFERENCES `contact` (`contact_id`);
--
-- ALTER TABLE `contest_contact_relationship` ADD FOREIGN KEY (`contest_id`) REFERENCES `contest` (`contest_id`);
--
-- ALTER TABLE `honors` ADD FOREIGN KEY (`user_id`) REFERENCES `user_profile_info` (`user_id`);
--
-- ALTER TABLE `user_profile_info` ADD FOREIGN KEY (`user_id`) REFERENCES `authentication` (`user_id`);

-- ALTER TABLE `team_info` ADD FOREIGN KEY (`contest_id`) REFERENCES `contest` (`contest_id`);
--
-- ALTER TABLE `team_info` ADD FOREIGN KEY (`leader_id`) REFERENCES `user_profile_info` (`user_id`);
--
-- ALTER TABLE `team_application` ADD FOREIGN KEY (`user_id`) REFERENCES `user_profile_info` (`user_id`);
--
-- ALTER TABLE `team_application` ADD FOREIGN KEY (`team_id`) REFERENCES `team_info` (`team_id`);
--
-- ALTER TABLE `team_user_relationship` ADD FOREIGN KEY (`user_id`) REFERENCES `user_profile_info` (`user_id`);
--
-- ALTER TABLE `team_user_relationship` ADD FOREIGN KEY (`team_id`) REFERENCES `team_info` (`team_id`);

ALTER TABLE `article` ADD FOREIGN KEY (`contest_id`) REFERENCES `contest` (`contest_id`);

ALTER TABLE `user_skills` ADD FOREIGN KEY (`user_id`) REFERENCES `user_profile_info` (`user_id`);

ALTER TABLE `team_skills` ADD FOREIGN KEY (`team_id`) REFERENCES `team_info` (`team_id`);
