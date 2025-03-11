package service

import (
	"context"
	"sync"

	"github.com/Yra-A/Fusion_Go/cmd/user/dal/db"
	"github.com/Yra-A/Fusion_Go/kitex_gen/user"
)

type QueryUserProfileService struct {
	ctx context.Context
}

func NewQueryUserProfileService(ctx context.Context) *QueryUserProfileService {
	return &QueryUserProfileService{ctx: ctx}
}

func (s *QueryUserProfileService) QueryUserProfile(user_id int32) (*user.UserProfileInfo, error) {
	u := &user.UserProfileInfo{}
	tasks := []TaskFunc{
		func() error { return s.FetchUserProfileInfo(user_id, u) },
		func() error { return s.FetchUserSkills(user_id, u) },
		func() error { return s.FetchUserHonors(user_id, u) },
		func() error { return s.FetchUserInfo(user_id, u) },
	}

	errChan := make(chan error, len(tasks))
	defer close(errChan)
	var wg sync.WaitGroup
	for _, task := range tasks {
		wg.Add(1)
		go func(t TaskFunc) {
			defer wg.Done()
			if err := t(); err != nil {
				errChan <- err
			}
		}(task)
	}
	wg.Wait()
	select {
	case err := <-errChan:
		return nil, err
	default:
	}
	return u, nil
}

func (s *QueryUserProfileService) FetchUserProfileInfo(user_id int32, u *user.UserProfileInfo) error {
	dbUserProfileInfo, err := db.QueryUserProfileByUserId(db.DB, user_id)
	if err != nil {
		return err
	}
	u.Introduction = dbUserProfileInfo.Introduction
	u.QqNumber = dbUserProfileInfo.QQNumber
	u.WechatNumber = dbUserProfileInfo.WeChatNumber
	return nil
}

func (s *QueryUserProfileService) FetchUserSkills(user_id int32, u *user.UserProfileInfo) error {
	dbSkills, err := db.QueryUserSkillsByUserId(user_id)
	if err != nil {
		return err
	}
	if len(dbSkills) == 0 {
		u.UserSkills = []*user.UserSkill{}
		return nil
	}
	// 转换技能列表
	userSkills := make([]*user.UserSkill, len(dbSkills))
	for i, skill := range dbSkills {
		if skill == nil {
			continue
		}
		userSkills[i] = &user.UserSkill{
			UserSkillId: skill.UserSkillID,
			UserId:     skill.UserID,
			Skill:      skill.Skill,
			Category:   skill.Category,
			Proficiency: skill.Proficiency,
		}
	}
	u.UserSkills = userSkills
	return nil
}

func (s *QueryUserProfileService) FetchUserHonors(user_id int32, u *user.UserProfileInfo) error {
	dbHonors, err := db.QueryHonorsByUserId(user_id)
	if err != nil {
		return err
	}
	u.Honors = dbHonors
	return nil
}

func (s *QueryUserProfileService) FetchUserInfo(user_id int32, u *user.UserProfileInfo) error {
	dbUserInfo, err := db.QueryUserByUserId(user_id)
	if err != nil {
		return err
	}
	u.UserInfo = &user.UserInfo{
		UserId:         dbUserInfo.UserID,
		Gender:         dbUserInfo.Gender,
		EnrollmentYear: dbUserInfo.EnrollmentYear,
		MobilePhone:    dbUserInfo.MobilePhone,
		College:        dbUserInfo.College,
		Nickname:       dbUserInfo.Nickname,
		Realname:       dbUserInfo.Realname,
		HasProfile:     dbUserInfo.HasProfile,
		AvatarUrl:      dbUserInfo.AvatarURL,
	}
	return nil
}
