namespace go user

// 用户基本信息
struct UserInfo {
    1: i32 user_id,
    2: i32 gender,
    4: i32 enrollment_year,
    3: string mobile_phone
    5: string college,
    6: string nickname,
    7: string realname,
    8: bool has_profile,
    9: string avatar_url,
}
// 用户所掌握技能
struct UserSkill {
    1: i32 user_skill_id,
    2: i32 user_id,
    3: string skill,
    4: string category,
    5: string proficiency  // 一般/良好/熟练/精通
}

// 用户档案信息（全部）
struct UserProfileInfo {
    1: string introduction,
    2: string qq_number,
    3: string wechat_number,
    4: list<UserSkill> user_skills,
    5: list<string> honors,
    6: UserInfo user_info,
}

// 用户注册
struct UserRegisterRequest {
    1: string username
    2: string password
}

struct UserRegisterResponse {
    1: i32 status_code,
    2: string status_msg,
}

// 用户登录
struct UserLoginRequest {
    1: string username
    2: string password
}

struct UserLoginResponse {
    1: i32 status_code,
    2: string status_msg,
    3: i32 user_id,
}

// 获取用户信息
struct UserInfoRequest {
    1: i32 user_id
}

struct UserInfoResponse {
    1: i32 status_code,
    2: string status_msg,
    3: UserInfo user_info,
}

// 上传用户信息
struct UserInfoUploadRequest {
    1: UserInfo user_info
}

struct UserInfoUploadResponse {
    1: i32 status_code,
    2: string status_msg,
}

// 获取用户档案信息
struct UserProfileInfoRequest {
    1: i32 user_id
}

struct UserProfileInfoResponse {
    1: i32 status_code,
    2: string status_msg,
    3: UserProfileInfo user_profile_info,
}

// 上传用户档案信息

struct UserProfileUploadRequest {
    1: i32 user_id
    2: UserProfileInfo user_profile_info
}

struct UserProfileUploadResponse {
    1: i32 status_code,
    2: string status_msg,
}

service UserService {
    // 用户注册操作
    UserRegisterResponse UserRegister(1: UserRegisterRequest req)
    // 用户登录操作
    UserLoginResponse UserLogin(1: UserLoginRequest req)
    // 获取用户信息
    UserInfoResponse UserInfo(1: UserInfoRequest req)
    // 上传用户信息
    UserInfoUploadResponse UserInfoUpload(1: UserInfoUploadRequest req)
    // 获取用户档案信息
    UserProfileInfoResponse UserProfileInfo(1: UserProfileInfoRequest req)
    // 上传用户档案信息
    UserProfileUploadResponse UserProfileUpload(1: UserProfileUploadRequest req)
}