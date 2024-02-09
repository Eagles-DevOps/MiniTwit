package main


import "fmt"
	

func main() {
    r := mux.NewRouter()
    r.HandleFunc("/", Home)
    r.HandleFunc("/timeline", TimeLine)
    r.HandleFunc("/public", Public)
    http.Handle("/", r)
}

func Home(w http.ResponseWriter, r *http.Request){
	vars := mux.Vars(r)
    w.WriteHeader(http.StatusOK)
    fmt.Fprintf(w, "Category: %v\n", vars["category"])
	fmt.Fprintf("Hello world")
}



