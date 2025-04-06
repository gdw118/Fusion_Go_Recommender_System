package oss

import "os"

var (
	AccessKeyID     = os.Getenv("ALIBABA_CLOUD_ACCESS_KEY_ID")
	AccessKeySecret = os.Getenv("ALIBABA_CLOUD_ACCESS_KEY_SECRET")
)

const (
	EndPoint = "oss-cn-shanghai.aliyuncs.com"

	OssBucketName = "gdw-fusion"
	BaseURL       = "public/fusion/"
	PublicURL     = "https://gdw-fusion.oss-cn-shanghai.aliyuncs.com/" + BaseURL
)