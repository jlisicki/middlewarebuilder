package httphandler

import (
	"github.com/jlisicki/middlewarebuilder"
	"log"
	"net/http"
)

// GetUserHandler returns user entry.
type GetUserHandler struct{}

func (g GetUserHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	_, _ = writer.Write([]byte("user"))
}

var _ http.Handler = GetUserHandler{}

// CreateGetUserHandler creates user handler with all required middlewares.
func CreateGetUserHandler() (http.Handler, error) {
	return middlewarebuilder.NewBuilder[http.Handler]().
		Add(middlewarebuilder.FactoryFunc[http.Handler](func(next http.Handler) (http.Handler, error) {
			return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				log.Printf("%s %s", request.Method, request.URL)
				next.ServeHTTP(writer, request)
			}), nil
		})).
		WithHandler(GetUserHandler{}).
		Build()
}
