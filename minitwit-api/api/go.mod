module minitwit.com/api

go 1.22.0

require minitwit.com/db v0.0.0
require minitwit.com/model v0.0.0

require github.com/mattn/go-sqlite3 v1.14.22 // indirect

replace minitwit.com/db => ../db
replace minitwit.com/model => ../model

