package service

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/Yra-A/Fusion_Go/cmd/team/dal/db"
	"github.com/Yra-A/Fusion_Go/kitex_gen/team"
	"github.com/Yra-A/Fusion_Go/kitex_gen/user"
	"github.com/Yra-A/Fusion_Go/pkg/configs/openai"
	"github.com/cloudwego/kitex/pkg/klog"
	openaiClient "github.com/sashabaranov/go-openai"
)

// RecommenderService 推荐服务
type RecommenderService struct {
	openaiClient *openaiClient.Client
	db           *db.TeamDB
}

// NewRecommenderService 创建推荐服务实例
func NewRecommenderService(db *db.TeamDB) *RecommenderService {
	client, err := openai.NewClient()
	if err != nil {
		panic(fmt.Sprintf("Failed to create OpenAI client: %v", err))
	}

	return &RecommenderService{
		openaiClient: client,
		db:           db,
	}
}

// getGenderText 将性别数字转换为文字描述
func getGenderText(gender int32) string {
	switch gender {
	case 1:
		return "男"
	case 2:
		return "女"
	default:
		return "未知"
	}
}

// GenerateUserEmbedding 生成用户embedding
func (s *RecommenderService) GenerateUserEmbedding(ctx context.Context, userProfile *user.UserProfileInfo) ([]float64, error) {
	klog.CtxInfof(ctx, "开始生成用户embedding, userID=%v", userProfile.UserInfo.UserId)
	
	// 构建用户描述
	description := fmt.Sprintf("性别: %s, 入学年份: %d, 学院: %s, 昵称: %s, 自我介绍: %s, 荣誉: %v, 技能: %v", 
		getGenderText(userProfile.UserInfo.Gender),
		userProfile.UserInfo.EnrollmentYear,
		userProfile.UserInfo.College,
		userProfile.UserInfo.Nickname,
		userProfile.Introduction,
		userProfile.Honors,
		userProfile.UserSkills)
	
	klog.CtxInfof(ctx, "用户描述构建完成: %s", description)

	// 调用OpenAI API生成embedding
	resp, err := s.openaiClient.CreateEmbeddings(ctx, openaiClient.EmbeddingRequest{
		Input: description,
		Model: openai.EmbeddingModel,
	})
	if err != nil {
		klog.CtxErrorf(ctx, "OpenAI API调用失败: %v", err)
		return nil, fmt.Errorf("failed to create embedding: %v", err)
	}

	if len(resp.Data) == 0 {
		klog.CtxErrorf(ctx, "OpenAI API返回空数据")
		return nil, fmt.Errorf("no embedding data returned")
	}

	klog.CtxInfof(ctx, "成功生成用户embedding, 维度: %d", len(resp.Data[0].Embedding))

	// 将float32转换为float64
	embedding := make([]float64, len(resp.Data[0].Embedding))
	for i, v := range resp.Data[0].Embedding {
		embedding[i] = float64(v)
	}

	return embedding, nil
}

// RecommendTeams 推荐队伍
func (s *RecommenderService) RecommendTeams(ctx context.Context, userProfile *user.UserProfileInfo, contestID int32) ([]*team.TeamInfo, error) {
	klog.CtxInfof(ctx, "开始推荐队伍, userID=%v, contestID=%v", userProfile.UserInfo.UserId, contestID)
	
	// 获取用户嵌入向量
	userEmbedding, err := s.GenerateUserEmbedding(ctx, userProfile)
	if err != nil {
		return nil, err
	}

	// 获取所有队伍信息
	teams, err := s.db.GetContestTeamsWithEmbedding(contestID)
	if err != nil {
		klog.CtxErrorf(ctx, "获取队伍信息失败: %v", err)
		return nil, err
	}
	klog.CtxInfof(ctx, "获取到 %d 个队伍", len(teams))

	// 计算相似度并排序
	type teamScore struct {
		team  *team.TeamInfo
		score float64
	}
	var scores []teamScore
	var failedTeams []*team.TeamInfo

	for _, team := range teams {
		// 解析队伍的嵌入向量
		var teamEmbedding db.TeamEmbedding
		if err := json.Unmarshal([]byte(team.Embedding), &teamEmbedding); err != nil {
			klog.CtxErrorf(ctx, "解析队伍embedding失败, teamID=%v: %v", team.TeamBriefInfo.TeamId, err)
			// 将解析失败的队伍添加到失败列表
			failedTeams = append(failedTeams, team)
			continue
		}

		// 计算用户与队伍中每个岗位的匹配度，取最大值
		var maxScore float64
		for _, position := range teamEmbedding.Positions {
			score := db.CalculatePositionMatchScore(userEmbedding, position)
			klog.CtxInfof(ctx, "队伍 %v 的岗位 %s 与用户的匹配度: %v", 
				team.TeamBriefInfo.TeamId, position.Job, score)
			if score > maxScore {
				maxScore = score
			}
		}

		// 应用时间衰减
		timeBoost := calculateTimeBoost(team.TeamBriefInfo.CreatedTime)
		maxScore *= timeBoost

		scores = append(scores, teamScore{team: team, score: maxScore})
		klog.CtxInfof(ctx, "队伍 %v 的推荐分数: %v (时间衰减: %v)", team.TeamBriefInfo.TeamId, maxScore, timeBoost)
	}

	// 按相似度排序
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].score > scores[j].score
	})

	// 返回所有队伍,解析失败的队伍放在最后
	result := make([]*team.TeamInfo, 0, len(teams))
	for i := 0; i < len(scores); i++ {
		result = append(result, scores[i].team)
		klog.CtxInfof(ctx, "最终推荐队伍 %v, 分数: %v", scores[i].team.TeamBriefInfo.TeamId, scores[i].score)
	}
	
	// 将解析失败的队伍添加到结果末尾
	result = append(result, failedTeams...)
	for _, team := range failedTeams {
		klog.CtxInfof(ctx, "解析失败的队伍 %v 被添加到末尾", team.TeamBriefInfo.TeamId)
	}

	klog.CtxInfof(ctx, "推荐完成, 共推荐 %d 个队伍", len(result))
	return result, nil
}

// calculateTimeBoost 计算时间衰减
func calculateTimeBoost(createdTime int64) float64 {
	now := time.Now().Unix()
	days := float64(now-createdTime) / (24 * 60 * 60)
	
	// 使用指数衰减函数，半衰期为30天，并设置最小衰减值为0.5
	halfLife := 30.0
	decay := math.Exp(-math.Ln2 * days / halfLife)
	
	// 设置最小衰减值
	if decay < 0.5 {
		return 0.5
	}
	return decay
} 