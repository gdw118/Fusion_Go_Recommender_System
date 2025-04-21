package db

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Yra-A/Fusion_Go/cmd/team/embedding"
	"github.com/Yra-A/Fusion_Go/kitex_gen/team"
	"github.com/Yra-A/Fusion_Go/pkg/configs/openai"
	"github.com/Yra-A/Fusion_Go/pkg/errno"
)

type TeamInfo struct {
	TeamID              int32     `gorm:"primary_key;column:team_id"`
	ContestID           int32     `gorm:"column:contest_id"`
	Title               string    `gorm:"column:title"`
	Goal                string    `gorm:"column:goal"`
	CurPeopleNum        int32     `gorm:"column:cur_people_num"`
	CreatedTime         time.Time `gorm:"column:created_time"`
	LeaderID            int32     `gorm:"column:leader_id"`
	Description         string    `gorm:"column:description"`
	Embedding           string    `gorm:"column:embedding;type:json"` // 存储为 JSON 字符串
	EmbeddingUpdatedTime time.Time `gorm:"column:embedding_updated_time"`
}

func (TeamInfo) TableName() string {
	return "team_info"
}

// TeamSkills 队伍招募岗位所需要的技能
type TeamSkills struct {
	TeamSkillID int32  `gorm:"primary_key;column:team_skill_id;autoIncrement"`
	TeamID      int32  `gorm:"column:team_id"`
	Skill       string `gorm:"column:skill"`
	Category    string `gorm:"column:category"`
	Job         string `gorm:"column:job"`
}

func (TeamSkills) TableName() string {
	return "team_skills"
}

type TeamApplication struct {
	ApplicationID   int32     `gorm:"primary_key;column:application_id"`
	UserID          int32     `gorm:"column:user_id"`
	TeamID          int32     `gorm:"column:team_id"`
	Reason          string    `gorm:"column:reason"`
	CreatedTime     time.Time `gorm:"column:created_time"`
	ApplicationType int32     `gorm:"column:application_type"`
}

func (TeamApplication) TableName() string {
	return "team_application"
}

type TeamUserRelationship struct {
	TeamUserID int32 `gorm:"primary_key;column:team_user_id"`
	UserID     int32 `gorm:"column:user_id"`
	TeamID     int32 `gorm:"column:team_id"`
}

func (TeamUserRelationship) TableName() string {
	return "team_user_relationship"
}

// TeamDB 实现TeamInfoProvider接口
type TeamDB struct{}

func NewTeamDB() *TeamDB {
	return &TeamDB{}
}

// GetTeamInfo 实现TeamInfoProvider接口
func (t *TeamDB) GetTeamInfo(teamID int32) (*team.TeamInfo, error) {
	return QueryTeamInfo(teamID)
}

// GetContestTeamsWithEmbedding 实现TeamInfoProvider接口
func (t *TeamDB) GetContestTeamsWithEmbedding(contestID int32) ([]*team.TeamInfo, error) {
	return GetContestTeamsWithEmbedding(contestID)
}

// UpdateTeamEmbedding 实现TeamInfoProvider接口
func (t *TeamDB) UpdateTeamEmbedding(teamID int32, embedding []float64, updatedTime time.Time) error {
	fmt.Printf("开始更新数据库中的embedding，teamID: %d, embedding长度: %d\n", teamID, len(embedding))
	
	// 将 embedding 数组序列化为 JSON 字符串
	embeddingJSON, err := json.Marshal(embedding)
	if err != nil {
		fmt.Printf("序列化 embedding 失败: %v\n", err)
		return err
	}
	
	if err := DB.Model(&TeamInfo{}).Where("team_id = ?", teamID).Updates(map[string]interface{}{
		"embedding": string(embeddingJSON),
		"embedding_updated_time": updatedTime,
	}).Error; err != nil {
		fmt.Printf("更新数据库失败: %v\n", err)
		return err
	}
	
	fmt.Printf("数据库更新成功\n")
	return nil
}

// CreateTeamSkills 创建团队技能需求
func CreateTeamSkills(teamID int32, skills []*team.TeamSkill) error {
	for _, skill := range skills {
		if err := DB.Create(&TeamSkills{
			TeamID:   teamID,
			Skill:    skill.Skill,
			Category: skill.Category,
			Job:      skill.Job,
		}).Error; err != nil {
			return err
		}
	}
	return nil
}

// GetTeamSkills 获取团队技能需求
func GetTeamSkills(teamID int32) ([]*team.TeamSkill, error) {
	var teamSkills []*TeamSkills
	if err := DB.Where("team_id = ?", teamID).Find(&teamSkills).Error; err != nil {
		return nil, err
	}
	var skills []*team.TeamSkill
	for _, ts := range teamSkills {
		skills = append(skills, &team.TeamSkill{
			TeamSkillId: ts.TeamSkillID,
			TeamId:      ts.TeamID,
			Skill:       ts.Skill,
			Category:    ts.Category,
			Job:         ts.Job,
		})
	}
	return skills, nil
}

// CreateTeam 创建团队
func CreateTeam(user_id int32, contest_id int32, title string, goal string, description string, skills []*team.TeamSkill) (int32, error) {
	team := &TeamInfo{
		Title:        title,
		ContestID:    contest_id,
		Goal:         goal,
		CurPeopleNum: 1,
		CreatedTime:  time.Now(),
		LeaderID:     user_id,
		Description:  description,
		EmbeddingUpdatedTime: time.Now(), // 初始化 embedding 更新时间
	}
	if err := DB.Create(team).Error; err != nil {
		return 0, err
	}
	var teamInfo TeamInfo
	// 获取最新创建的 team 的 team_id，Last 会返回按主键排序的最后一个满足条件的记录
	if err := DB.Select("team_id").Where("leader_id = ?", user_id).Last(&teamInfo).Error; err != nil {
		return 0, err
	}
	
	// 创建团队技能需求
	if err := CreateTeamSkills(teamInfo.TeamID, skills); err != nil {
		return 0, err
	}
	
	TeamAddUser(teamInfo.TeamID, user_id)

	// 生成并更新embedding
	client, err := openai.NewClient()
	if err != nil {
		fmt.Printf("Warning: %v, skipping embedding generation for team %d\n", err, teamInfo.TeamID)
		return teamInfo.TeamID, nil
	}
	embeddingService := embedding.NewService(context.Background(), NewTeamDB(), client)
	if err := embeddingService.UpdateTeamEmbedding(teamInfo.TeamID); err != nil {
		// 如果embedding生成失败，记录错误但不影响队伍创建
		fmt.Printf("Failed to generate embedding for team %d: %v\n", teamInfo.TeamID, err)
	}

	return teamInfo.TeamID, nil
}

// ModifyTeam 修改团队信息
func ModifyTeam(team_id int32, title string, goal string, description string, skills []*team.TeamSkill) error {
	// 修改团队基本信息
	team := &TeamInfo{
		TeamID: team_id,
	}
	if err := DB.Model(&team).Updates(map[string]interface{}{"title": title, "goal": goal, "description": description}).Error; err != nil {
		return err
	}

	// 删除原有的技能需求
	if err := DB.Where("team_id = ?", team_id).Delete(&TeamSkills{}).Error; err != nil {
		return err
	}

	// 创建新的技能需求
	if err := CreateTeamSkills(team_id, skills); err != nil {
		return err
	}

	// 更新embedding
	client, err := openai.NewClient()
	if err != nil {
		fmt.Printf("Warning: %v, skipping embedding update for team %d\n", err, team_id)
		return nil
	}
	embeddingService := embedding.NewService(context.Background(), NewTeamDB(), client)
	if err := embeddingService.UpdateTeamEmbedding(team_id); err != nil {
		// 如果embedding生成失败，记录错误但不影响队伍修改
		fmt.Printf("Failed to update embedding for team %d: %v\n", team_id, err)
	}

	return nil
}

func QueryTeamList(contest_id int32) ([]*team.TeamBriefInfo, error) {
	var teamList []*TeamInfo
	if err := DB.Where("contest_id = ?", contest_id).Find(&teamList).Error; err != nil {
		return nil, err
	}
	var teamBriefInfoList []*team.TeamBriefInfo
	for _, t := range teamList {
		teamBriefInfoList = append(teamBriefInfoList, &team.TeamBriefInfo{
			TeamId:       t.TeamID,
			Title:        t.Title,
			Goal:         t.Goal,
			CurPeopleNum: t.CurPeopleNum,
			CreatedTime:  t.CreatedTime.Unix(),
			ContestId:    contest_id,
			LeaderInfo: &team.MemberInfo{
				UserId: t.LeaderID,
			},
		})
	}
	return teamBriefInfoList, nil
}

// QueryTeamInfo 查询队伍信息
func QueryTeamInfo(team_id int32) (*team.TeamInfo, error) {
	var teamInfo TeamInfo
	if err := DB.Where("team_id = ?", team_id).First(&teamInfo).Error; err != nil {
		return nil, err
	}
	
	// 获取团队技能需求
	teamSkills, err := GetTeamSkills(team_id)
	if err != nil {
		return nil, err
	}
	
	var teamUserRelationship []*TeamUserRelationship
	if err := DB.Where("team_id = ?", team_id).Find(&teamUserRelationship).Error; err != nil {
		return nil, err
	}
	var memberList []*team.MemberInfo
	for _, t := range teamUserRelationship {
		memberList = append(memberList, &team.MemberInfo{
			UserId: t.UserID,
		})
	}
	return &team.TeamInfo{
		TeamBriefInfo: &team.TeamBriefInfo{
			TeamId:       teamInfo.TeamID,
			Title:        teamInfo.Title,
			Goal:         teamInfo.Goal,
			CurPeopleNum: teamInfo.CurPeopleNum,
			CreatedTime:  teamInfo.CreatedTime.Unix(),
			ContestId:    teamInfo.ContestID,
			LeaderInfo: &team.MemberInfo{
				UserId: teamInfo.LeaderID,
			},
		},
		Description: teamInfo.Description,
		TeamSkills:  teamSkills,
		Members:     memberList,
	}, nil
}

func CreateTeamApplication(user_id int32, team_id int32, reason string, created_time int64, application_type int32) error {
	if err := DB.Create(&TeamApplication{
		UserID:          user_id,
		TeamID:          team_id,
		Reason:          reason,
		CreatedTime:     time.Unix(created_time, 0),
		ApplicationType: application_type,
	}).Error; err != nil {
		return err
	}
	return nil
}

func GetTeamApplicationList(user_id int32, team_id int32) ([]*team.TeamApplication, error) {
	var teamInfo TeamInfo
	if err := DB.Select("leader_id").Where("team_id = ?", team_id).First(&teamInfo).Error; err != nil {
		return nil, err
	}
	// 只有队长才能处理申请
	if teamInfo.LeaderID != user_id {
		return nil, errno.AuthorizationFailedErr
	}
	var teamApplicationList []*TeamApplication
	if err := DB.Where("team_id = ?", team_id).Find(&teamApplicationList).Error; err != nil {
		return nil, err
	}
	var teamApplications []*team.TeamApplication
	for _, t := range teamApplicationList {
		if t.ApplicationType != 0 {
			teamApplications = append(teamApplications, &team.TeamApplication{
				ApplicationId:   t.ApplicationID,
				TeamId:          t.TeamID,
				Reason:          t.Reason,
				CreatedTime:     t.CreatedTime.Unix(),
				ApplicationType: t.ApplicationType,
				MemberInfo: &team.MemberInfo{
					UserId: t.UserID,
				},
			})
		}
	}
	return teamApplications, nil
}
func TeamAddUser(team_id int32, member_id int32) error {
	var record TeamUserRelationship
	DB.Where("team_id = ? AND user_id = ?", team_id, member_id).First(&record)
	if record.TeamUserID != 0 {
		return nil
	}
	if err := DB.Create(&TeamUserRelationship{
		UserID: member_id,
		TeamID: team_id,
	}).Error; err != nil {
		return err
	}
	// 更新 cur number
	var count int64
	if err := DB.Model(&TeamUserRelationship{}).Where("team_id = ?", team_id).Count(&count).Error; err != nil {
		return err
	}

	if err := DB.Model(&TeamInfo{}).Where("team_id = ?", team_id).Update("cur_people_num", count).Error; err != nil {
		return err
	}
	return nil
}

func TeamManageAction(user_id int32, application_id int32, action_type int32) error {
	var teamApplication TeamApplication
	if err := DB.Where("application_id = ?", application_id).First(&teamApplication).Error; err != nil {
		return err
	}
	var teamInfo TeamInfo
	if err := DB.Select("leader_id").Where("team_id = ?", teamApplication.TeamID).First(&teamInfo).Error; err != nil {
		return err
	}
	// 只有队长才能处理申请
	if teamInfo.LeaderID != user_id {
		return errno.AuthorizationFailedErr
	}

	// 接受申请
	if action_type == 1 {
		TeamAddUser(teamApplication.TeamID, teamApplication.UserID)
	}

	if err := DB.Model(&teamApplication).Update("application_type", 0).Error; err != nil {
		return err
	}
	return nil
}

// GetContestTeamsWithEmbedding 获取竞赛下的队伍及其 embedding
func GetContestTeamsWithEmbedding(contestID int32) ([]*team.TeamInfo, error) {
	var teamInfos []TeamInfo
	if err := DB.Where("contest_id = ?", contestID).Find(&teamInfos).Error; err != nil {
		return nil, err
	}

	var teams []*team.TeamInfo
	for _, t := range teamInfos {
		// 获取团队技能需求
		teamSkills, err := GetTeamSkills(t.TeamID)
		if err != nil {
			return nil, err
		}

		// 获取团队成员
		var teamUserRelationship []*TeamUserRelationship
		if err := DB.Where("team_id = ?", t.TeamID).Find(&teamUserRelationship).Error; err != nil {
			return nil, err
		}
		var memberList []*team.MemberInfo
		for _, tur := range teamUserRelationship {
			memberList = append(memberList, &team.MemberInfo{
				UserId: tur.UserID,
			})
		}

		// 解析 embedding
		var embedding []float64
		if t.Embedding != "" {
			if err := json.Unmarshal([]byte(t.Embedding), &embedding); err != nil {
				return nil, err
			}
		}

		teams = append(teams, &team.TeamInfo{
			TeamBriefInfo: &team.TeamBriefInfo{
				TeamId:       t.TeamID,
				Title:        t.Title,
				Goal:         t.Goal,
				CurPeopleNum: t.CurPeopleNum,
				CreatedTime:  t.CreatedTime.Unix(),
				ContestId:    t.ContestID,
				LeaderInfo: &team.MemberInfo{
					UserId: t.LeaderID,
				},
			},
			Description: t.Description,
			TeamSkills:  teamSkills,
			Members:     memberList,
			Embedding:   t.Embedding, // 直接使用原始的 JSON 字符串
		})
	}

	return teams, nil
}
