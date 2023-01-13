module github.com/swaggo/http-swagger/example/go-chi

go 1.13

require (
	github.com/go-chi/chi v4.1.2+incompatible
	github.com/swaggo/http-swagger v1.2.6
	github.com/swaggo/swag v1.8.1
)

replace github.com/swaggo/http-swagger => ../..
