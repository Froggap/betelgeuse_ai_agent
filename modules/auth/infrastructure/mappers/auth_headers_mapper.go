package authinframappers

import "net/http"

func AuthHeadersMapper(w http.ResponseWriter, token string, code string) {
	w.Header().Set("X-Refresh-Code", code)
	w.Header().Set("X-Refresh-Token", token)
}
