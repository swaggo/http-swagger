module github.com/swaggo/http-swagger/example/go-chi

go 1.13

replace github.com/swaggo/http-swagger => ./../../

require (
	github.com/go-chi/chi v4.1.2+incompatible
	github.com/swaggo/http-swagger v0.0.0-00010101000000-000000000000
	github.com/swaggo/swag v1.8.1
)
