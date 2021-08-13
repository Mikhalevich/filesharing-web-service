package wrapper

import (
	"net/http"
	"time"

	"github.com/Mikhalevich/filesharing-web-service/handler"
)

type CookieSession struct {
	expirePeriod int64
}

func NewCookieSession(period int64) *CookieSession {
	return &CookieSession{
		expirePeriod: period,
	}
}

func (cs *CookieSession) GetToken(name string, r *http.Request) (*handler.Token, error) {
	if name == "" {
		return nil, nil
	}

	for _, cook := range r.Cookies() {
		if cook.Name != name {
			continue
		}

		return &handler.Token{
			Value: cook.Value,
		}, nil
	}

	return nil, handler.ErrNotExist
}

func (cs *CookieSession) SetToken(w http.ResponseWriter, token *handler.Token, name string) {
	cookie := http.Cookie{Name: name, Value: token.Value, Path: "/", Expires: time.Now().Add(time.Duration(cs.expirePeriod) * time.Second), HttpOnly: true}
	http.SetCookie(w, &cookie)
}

func (cs *CookieSession) Remove(w http.ResponseWriter, name string) {
	http.SetCookie(w, &http.Cookie{Name: name, Value: "", Path: "/", Expires: time.Unix(0, 0), HttpOnly: true})
}

// func (cs *CookieSession) Create() goauth.Session {
// 	bytes := make([]byte, 32)
// 	rand.Read(bytes)
// 	return *goauth.NewSession("session", base64.URLEncoding.EncodeToString(bytes), cs.expirePeriod)
// }
