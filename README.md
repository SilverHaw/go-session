## Quick view

```
package main

import (
	"io"
	"net/http"
	sessions "test/g64-session"
)

var sessionMgr *sessions.SessionMgr

func login(w http.ResponseWriter, r *http.Request) {
	sessionID := sessionMgr.InitSession(w, r)
	sessionMgr.Set(sessionID, "key", "success")
	io.WriteString(w, "hello, world!\n")
}

func main() {
	session := sessions.NewSessionMgr("sessionid", 3600)
	sessionMgr = session
	http.HandleFunc("/", login)
	http.ListenAndServe(":8000", nil)
}
```





# Install 

```
git clone https://e.coding.net/g64/gpm/g64-session.git
```



