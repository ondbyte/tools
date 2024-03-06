package parser

import "net/http"

func main() {
	server := http.Server{}
	mux := http.NewServeMux()
	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		user_id = r.PathValue("user_id")
		order_id = r.PathValue("order_id")
		HandleUsers(userId, orderId)
	})
}
