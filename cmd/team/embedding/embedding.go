package embedding

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Yra-A/Fusion_Go/kitex_gen/team"
	"github.com/Yra-A/Fusion_Go/pkg/configs/openai"
	openaiClient "github.com/sashabaranov/go-openai"
)

// Service 负责生成和管理队伍的embedding
type Service struct {
	ctx            context.Context
	teamInfoProvider TeamInfoProvider
	openaiClient   *openaiClient.Client
}

func NewService(ctx context.Context, provider TeamInfoProvider, client *openaiClient.Client) *Service {
	return &Service{
		ctx:            ctx,
		teamInfoProvider: provider,
		openaiClient:   client,
	}
}

// GenerateTeamEmbedding 生成队伍的embedding
func (s *Service) GenerateTeamEmbedding(teamInfo *team.TeamInfo) ([]float32, error) {
	// 1. 构建队伍的文本描述
	description := s.buildTeamDescription(teamInfo)

	// 2. 调用LLM API生成embedding
	embedding, err := s.callLLMAPI(description)
	if err != nil {
		return nil, fmt.Errorf("failed to generate embedding: %v", err)
	}

	return embedding, nil
}

// buildTeamDescription 构建队伍的文本描述
func (s *Service) buildTeamDescription(teamInfo *team.TeamInfo) string {
	var sb strings.Builder

	// 添加队伍基本信息
	sb.WriteString(fmt.Sprintf("队伍名称：%s\n", teamInfo.TeamBriefInfo.Title))
	sb.WriteString(fmt.Sprintf("队伍目标：%s\n", teamInfo.TeamBriefInfo.Goal))
	sb.WriteString(fmt.Sprintf("队伍描述：%s\n", teamInfo.Description))

	// 添加技能需求
	if len(teamInfo.TeamSkills) > 0 {
		sb.WriteString("技能需求：\n")
		for _, skill := range teamInfo.TeamSkills {
			sb.WriteString(fmt.Sprintf("- %s（%s）：%s\n", skill.Job, skill.Skill, skill.Category))
		}
	}

	return sb.String()
}

// callLLMAPI 调用LLM API生成embedding
func (s *Service) callLLMAPI(text string) ([]float32, error) {
	// 调用OpenAI API生成embedding
	resp, err := s.openaiClient.CreateEmbeddings(
		context.Background(),
		openaiClient.EmbeddingRequest{
			Input: text,
			Model: openai.EmbeddingModel,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create embedding: %v", err)
	}

	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("no embedding data returned")
	}

	return resp.Data[0].Embedding, nil
}

// UpdateTeamEmbedding 更新队伍的embedding
func (s *Service) UpdateTeamEmbedding(teamID int32) error {
	fmt.Printf("开始更新队伍 %d 的embedding\n", teamID)
	
	// 1. 获取队伍信息
	teamInfo, err := s.teamInfoProvider.GetTeamInfo(teamID)
	if err != nil {
		return fmt.Errorf("failed to get team info: %v", err)
	}
	fmt.Printf("成功获取队伍信息: %+v\n", teamInfo)

	// 2. 生成embedding
	embedding32, err := s.GenerateTeamEmbedding(teamInfo)
	if err != nil {
		return fmt.Errorf("生成embedding失败: %v", err)
	}
	fmt.Printf("成功生成embedding，长度: %d\n", len(embedding32))

	// 3. 检查embedding长度
	if len(embedding32) == 0 {
		return fmt.Errorf("empty embedding vector")
	}

	// 4. 将 float32 转换为 float64
	embedding64 := make([]float64, len(embedding32))
	for i, v := range embedding32 {
		embedding64[i] = float64(v)
	}
	fmt.Printf("转换后的embedding64长度: %d\n", len(embedding64))

	// 5. 更新数据库
	if err := s.teamInfoProvider.UpdateTeamEmbedding(teamID, embedding64, time.Now()); err != nil {
		return fmt.Errorf("更新数据库失败: %v", err)
	}
	fmt.Printf("成功更新数据库中的embedding\n")

	return nil
} 