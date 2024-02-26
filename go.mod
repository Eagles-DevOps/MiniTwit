module minitwit-api

go 1.21

replace minitwit.com/api => ./api

replace minitwit.com/model => ./model

replace minitwit.com/db => ./db

replace minitwit.com/sim => ./sim

require (
	github.com/gorilla/mux v1.8.1
	minitwit.com/api v0.0.0
	minitwit.com/db v0.0.0
	minitwit.com/sim v0.0.0
)

require (
	github.com/mattn/go-sqlite3 v1.14.22 // indirect
	golang.org/x/crypto v0.19.0 // indirect
	minitwit.com/model v0.0.0 // indirect
)
