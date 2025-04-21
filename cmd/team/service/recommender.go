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
	"github.com/Yra-A/Fusion_Go/kitex_gen/user/userservice"
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
func (s *RecommenderService) GenerateUserEmbedding(ctx context.Context, userID int32) ([]float64, error) {
	klog.CtxInfof(ctx, "开始生成用户embedding, userID=%v", userID)
	
	// 从用户服务获取完整的用户信息
	userClient, err := userservice.NewClient("user")
	if err != nil {
		klog.CtxErrorf(ctx, "创建用户服务客户端失败: %v", err)
		return nil, fmt.Errorf("failed to create user client: %v", err)
	}

	// 获取用户信息
	userProfileResp, err := userClient.UserProfileInfo(ctx, &user.UserProfileInfoRequest{
		UserId: userID,
	})
	if err != nil {
		klog.CtxErrorf(ctx, "获取用户信息失败: %v", err)
		return nil, err
	}
	if userProfileResp.StatusCode != 0 {
		klog.CtxErrorf(ctx, "获取用户信息返回错误状态码: %v, 消息: %v", userProfileResp.StatusCode, userProfileResp.StatusMsg)
		return nil, fmt.Errorf("获取用户信息失败: %s", userProfileResp.StatusMsg)
	}

	// 构建用户描述
	description := fmt.Sprintf("性别: %s, 入学年份: %d, 学院: %s, 昵称: %s, 自我介绍: %s, 荣誉: %v, 技能: %v", 
		getGenderText(userProfileResp.UserProfileInfo.UserInfo.Gender),
		userProfileResp.UserProfileInfo.UserInfo.EnrollmentYear,
		userProfileResp.UserProfileInfo.UserInfo.College,
		userProfileResp.UserProfileInfo.UserInfo.Nickname,
		userProfileResp.UserProfileInfo.Introduction,
		userProfileResp.UserProfileInfo.Honors,
		userProfileResp.UserProfileInfo.UserSkills)
	
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
func (s *RecommenderService) RecommendTeams(ctx context.Context, userID int32, contestID int32) ([]*team.TeamInfo, error) {
	klog.CtxInfof(ctx, "开始推荐队伍, userID=%v, contestID=%v", userID, contestID)
	
	// 获取用户嵌入向量
	userEmbedding, err := s.GenerateUserEmbedding(ctx, userID)
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

	for _, team := range teams {
		// 解析队伍的嵌入向量
		var teamEmbedding []float64
		if err := json.Unmarshal([]byte(team.Embedding), &teamEmbedding); err != nil {
			klog.CtxErrorf(ctx, "解析队伍embedding失败, teamID=%v: %v", team.TeamBriefInfo.TeamId, err)
			continue // 跳过解析失败的队伍
		}

		// 计算余弦相似度
		score := cosineSimilarity(userEmbedding, teamEmbedding)
		timeBoost := calculateTimeBoost(team.TeamBriefInfo.CreatedTime)
		score *= timeBoost
		scores = append(scores, teamScore{team: team, score: score})
		klog.CtxInfof(ctx, "队伍 %v 的推荐分数: %v (时间衰减: %v)", team.TeamBriefInfo.TeamId, score, timeBoost)
	}

	// 按相似度排序
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].score > scores[j].score
	})

	// 返回前10个推荐队伍
	result := make([]*team.TeamInfo, 0, 10)
	for i := 0; i < len(scores) && i < 10; i++ {
		result = append(result, scores[i].team)
		klog.CtxInfof(ctx, "最终推荐队伍 %v, 分数: %v", scores[i].team.TeamBriefInfo.TeamId, scores[i].score)
	}

	klog.CtxInfof(ctx, "推荐完成, 共推荐 %d 个队伍", len(result))
	return result, nil
}

// cosineSimilarity 计算余弦相似度
func cosineSimilarity(a, b []float64) float64 {
	if len(a) != len(b) {
		return 0
	}

	var dotProduct, normA, normB float64
	for i := 0; i < len(a); i++ {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}

// calculateTimeBoost 计算时间衰减
func calculateTimeBoost(createdTime int64) float64 {
	now := time.Now().Unix()
	days := float64(now-createdTime) / (24 * 60 * 60)
	
	// 使用指数衰减函数，半衰期为7天
	halfLife := 7.0
	return math.Exp(-math.Ln2 * days / halfLife)
} 