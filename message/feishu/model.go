package feishu

// CommonRes ...
type CommonRes struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

// AppAccessToken ...
type AppAccessToken struct {
	CommonRes
	TenantAccessToken string `json:"tenant_access_token"`
	AppAccessToken    string `json:"app_access_token"`
	Expire            int64  `json:"expire"`
}

// AppUserIDData ...
type AppUserIDData struct {
	OpenID string `json:"open_id"`
	UserID string `json:"user_id"`
}

// AppUsersData ...
type AppUsersData struct {
	EmailUsers      map[string][]AppUserIDData `json:"email_users"`
	EmailsNotExist  []string                   `json:"emails_not_exist"`
	MobileUsers     map[string][]AppUserIDData `json:"mobile_users"`
	MobilesNotExist []string                   `json:"mobiles_not_exist"`
}

// AppUserInfo 用户信息
type AppUserInfo struct {
	CommonRes
	Data AppUsersData `json:"data"`
}

// AppContentData 消息内容
type AppContentData struct {
	Tag      string `json:"tag,omitempty"`
	Text     string `json:"text,omitempty"`
	ImageKey string `json:"image_key,omitempty"`
	Width    int    `json:"width,omitempty"`
	Height   int    `json:"height,omitempty"`
	Href     string `json:"href,omitempty"`
}

type AppKeyData struct {
	AppID     string
	AppSecret string
}
