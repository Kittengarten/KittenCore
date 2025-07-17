package fanqie

import "net/url"

const (
	// SNSSDK ...
	SNSSDK = `https://novel.snssdk.com/`
	// API ...
	SNSSDKAPI = `api`
	// Novel ...
	Novel = `novel`
	// Book ...
	Book = `book`
	// Directory ...
	Directory = `directory`
	// V ...
	V = `v`
	// AID ...
	AID = `aid`
	// AIDValue ...
	AIDValue = `1319`
	// ItemIDs ...
	ItemIDs = `item_ids`
)

// SNSSDKDetail 从 SNSSDK 获取详情的 API
func SNSSDKDetail(cpURL string) (*url.URL, error) {
	u, err := url.Parse(SNSSDK)
	if err != nil {
		return nil, err
	}
	u = u.JoinPath(SNSSDKAPI, Novel, Book, Directory, Detail, V)
	v := u.Query()
	v.Add(AID, AIDValue)
	v.Add(ItemIDs, API.ChapterID(cpURL))
	u.RawQuery = v.Encode()
	return u, nil
}
