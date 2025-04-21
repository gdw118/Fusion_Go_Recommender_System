package embedding

import (
	"time"

	"github.com/Yra-A/Fusion_Go/kitex_gen/team"
)

// TeamInfoProvider 定义获取队伍信息的接口
type TeamInfoProvider interface {
	GetTeamInfo(teamID int32) (*team.TeamInfo, error)
	GetContestTeamsWithEmbedding(contestID int32) ([]*team.TeamInfo, error)
	UpdateTeamEmbedding(teamID int32, embedding []float64, updatedTime time.Time) error
} 