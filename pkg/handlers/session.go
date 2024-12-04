package handlers

import (
	"context"
	"fmt"
	"net/http"
	"pr10/pkg/session"
	"time"

	"go.uber.org/zap"
)

var loginFormTmpl = []byte(`
<html>
	<body>
	<form action="/api/login" method="post">
		Login: <input type="text" name="login">
		Password: <input type="password" name="password">
		<input type="submit" value="Login">
	</form>
	</body>
</html>
`)

type SessionHandler struct {
	SessManager session.AuthCheckerClient
	Logger *zap.SugaredLogger
}

func (h *SessionHandler) checkSession(r *http.Request) (*session.Session, error) {
	cookieSessionID, err := r.Cookie("session_id")
	if err == http.ErrNoCookie {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	sess, err := h.SessManager.Check(
		context.Background(),
		&session.SessionID{
			ID: cookieSessionID.Value,
		})
	if err != nil {
		return nil, err
	}
	return sess, nil
}

func (h *SessionHandler) InnerPage(w http.ResponseWriter, r *http.Request) {
	sess, err := h.checkSession(r)
	if err != nil {
		w.Write(loginFormTmpl)
		return
	}
	if sess == nil {
		w.Write(loginFormTmpl)
		return
	}

	h.Logger.Infof("Inner page for %s", sess.Login)

	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintln(w, "Welcome, "+sess.Login+" <br />")
	fmt.Fprintln(w, "Session ua: "+sess.Useragent+" <br />")
	fmt.Fprintln(w, `<a href="/api/logout">logout</a>`)
}

func (h *SessionHandler) LoginPage(w http.ResponseWriter, r *http.Request) {
	inputLogin := r.FormValue("login")
	expiration := time.Now().Add(365 * 24 * time.Hour)

	sess, err := h.SessManager.Create(
		context.Background(),
		&session.Session{
			Login:     inputLogin,
			Useragent: r.UserAgent(),
		})
	if err != nil {
		h.Logger.Error("Cant create session")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	cookie := http.Cookie{
		Name:    "session_id",
		Value:   sess.ID,
		Expires: expiration,
	}
	http.SetCookie(w, &cookie)
	http.Redirect(w, r, "/api/", http.StatusFound)
}

func (h *SessionHandler) LogoutPage(w http.ResponseWriter, r *http.Request) {
	cookieSessionID, err := r.Cookie("session_id")
	if err == http.ErrNoCookie {
		http.Redirect(w, r, "/api/", http.StatusFound)
		return
	} else if err != nil {
		h.Logger.Error("Cookie not found")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	h.SessManager.Delete(
		context.Background(),
		&session.SessionID{
			ID: cookieSessionID.Value,
		})

	cookieSessionID.Expires = time.Now().AddDate(0, 0, -1)
	http.SetCookie(w, cookieSessionID)

	http.Redirect(w, r, "/api/", http.StatusFound)
}
