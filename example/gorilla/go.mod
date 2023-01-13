module github.com/swaggo/http-swagger/example/gorilla

go 1.13

require (
	github.com/gorilla/mux v1.8.0
	github.com/swaggo/http-swagger v1.2.6
	github.com/swaggo/swag v1.8.1
)

replace github.com/swaggo/http-swagger => ../..
