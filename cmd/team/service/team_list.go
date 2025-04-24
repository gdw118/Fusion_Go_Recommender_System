package service

import (
	"context"

	"github.com/Yra-A/Fusion_Go/cmd/team/dal/db"
	"github.com/Yra-A/Fusion_Go/cmd/team/rpc"
	"github.com/Yra-A/Fusion_Go/kitex_gen/team"
	"github.com/Yra-A/Fusion_Go/kitex_gen/user"
	"github.com/cloudwego/kitex/pkg/klog"
)

type TeamListService struct {
	ctx context.Context
}

func NewTeamListService(ctx context.Context) *TeamListService {
	return &TeamListService{ctx: ctx}
}

func (s *TeamListService) TeamList(contest_id int32, limit int32, offset int32, user_id int32) ([]*team.TeamBriefInfo, int32, error) {
	// 获取基础队伍列表
	teamList, err := db.QueryTeamList(contest_id, user_id)
	if err != nil {
		return nil, 0, err
	}

	var userProfile *user.UserProfileInfo
	// 如果提供了 user_id，使用推荐系统进行个性化排序
	if user_id != 0 {
		klog.CtxInfof(s.ctx, "使用推荐系统进行个性化排序, userID=%v", user_id)
		
		// 获取用户信息
		kresp, err := rpc.UserProfileInfo(s.ctx, &user.UserProfileInfoRequest{UserId: user_id})
		if err != nil {
			klog.CtxErrorf(s.ctx, "获取用户信息失败: %v", err)
			// 如果获取用户信息失败，继续使用原有逻辑
		} else {
			userProfile = kresp.UserProfileInfo
			recommender := NewRecommenderService(db.NewTeamDB())
			recommendedTeams, err := recommender.RecommendTeams(s.ctx, userProfile, contest_id)
			if err != nil {
				klog.CtxErrorf(s.ctx, "推荐系统调用失败: %v", err)
				// 如果推荐失败，继续使用原有逻辑
			} else {
				// 将推荐结果转换为 TeamBriefInfo
				recommendedBriefInfos := make([]*team.TeamBriefInfo, 0, len(recommendedTeams))
				for _, t := range recommendedTeams {
					recommendedBriefInfos = append(recommendedBriefInfos, t.TeamBriefInfo)
				}
				teamList = recommendedBriefInfos
			}
		}
	}

	total := len(teamList)
	if offset < int32(len(teamList)) {
		if offset+limit >= int32(len(teamList)) {
			teamList = teamList[offset:]
		} else {
			teamList = teamList[offset : offset+limit]
		}
	} else {
		teamList = nil
	}

	// 补充队长信息
	for _, t := range teamList {
		// 如果队长是当前用户且已获取用户信息,直接使用已有的用户信息
		if t.LeaderInfo.UserId == user_id && userProfile != nil {
			t.LeaderInfo = &team.MemberInfo{
				UserId:         userProfile.UserInfo.UserId,
				Nickname:       userProfile.UserInfo.Nickname,
				AvatarUrl:      userProfile.UserInfo.AvatarUrl,
				College:        userProfile.UserInfo.College,
				EnrollmentYear: userProfile.UserInfo.EnrollmentYear,
				Gender:         userProfile.UserInfo.Gender,
				Honors:         userProfile.Honors,
			}
		} else {
			// 否则获取队长信息
			kresp, err := rpc.UserProfileInfo(context.Background(), &user.UserProfileInfoRequest{UserId: t.LeaderInfo.UserId})
			if err != nil {
				return nil, 0, err
			}
			u := kresp.UserProfileInfo
			t.LeaderInfo = &team.MemberInfo{
				UserId:         u.UserInfo.UserId,
				Nickname:       u.UserInfo.Nickname,
				AvatarUrl:      u.UserInfo.AvatarUrl,
				College:        u.UserInfo.College,
				EnrollmentYear: u.UserInfo.EnrollmentYear,
				Gender:         u.UserInfo.Gender,
				Honors:         u.Honors,
			}
		}
	}
	return teamList, int32(total), nil
}
