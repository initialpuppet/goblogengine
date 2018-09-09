package basehandler

import (
	"context"
	"fmt"
	"html/template"
	"net/http"

	"goblogengine/appenv"

	"google.golang.org/appengine"
	"google.golang.org/appengine/log"

	"strings"
)

// HTTPHandler is the type used to adapt blog handlers to Gorilla Mux.
// TODO: consider something like https://github.com/justinas/alice
type HTTPHandler func(context.Context, appenv.AppEnv, http.ResponseWriter, *http.Request) *AppError

// MakeHandler returns a function that can be passed to an HTTP router.
func MakeHandler(fn func(context.Context, appenv.AppEnv, http.ResponseWriter, *http.Request) *AppError) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := appengine.NewContext(r)
		env := appenv.GetEnv()

		if env.HostEnv == appenv.EnvLive {
			host := strings.ToLower(r.URL.Host)
			if host != env.Config.BaseDomainName {
				path := r.URL.Path
				redirectURL := fmt.Sprintf("https://%s%s",
					env.Config.BaseDomainName, path)
				http.Redirect(w, r, redirectURL, http.StatusMovedPermanently)
				log.Infof(ctx, "Redirecting to %s. Original Host:%s, Path:%s",
					redirectURL, host, path)
			}
		}

		if e := fn(ctx, env, w, r); e != nil {
			applicationError(ctx, w, r, e)
			return
		}
	}
}

// Returns an error message to the client. Self-contained to avoid infinite
// recursion. If template rendering fails, returns a basic error message
// instead.
// TODO: Return XML, JSON or HTML as error depending on the request
func applicationError(ctx context.Context, w http.ResponseWriter, r *http.Request, apperr *AppError) {
	log.Errorf(ctx, apperr.String())

	errorTemplate := "error/appdefault"
	if apperr.StatusCode == http.StatusNotFound {
		errorTemplate = "error/404"
	}

	if errtmpl, err := template.ParseFiles(fmt.Sprintf("templates/%s.html",
		errorTemplate)); err != nil {
		log.Errorf(ctx, fmt.Sprintf(
			"Error parsing template for application error page. [%s]",
			err.Error()))
	} else {
		w.WriteHeader(http.StatusInternalServerError)
		if err := errtmpl.Execute(w, apperr); err != nil {

			log.Errorf(ctx, fmt.Sprintf(
				"Error executing template for application error page. [%s]",
				err.Error()))
		} else {
			return
		}
	}
	http.Error(w, apperr.Message, http.StatusInternalServerError)
}

// AppError represents an error in an HTTP handler.
type AppError struct {
	Error      error
	Message    string
	StatusCode int
}

// String combines and returns all error information.
func (e AppError) String() string {
	var log string
	if e.Message != "" {
		log = e.Message
		if e.Error != nil {
			log += ":"
		}
	}
	if e.Error != nil {
		log += " " + e.Error.Error()
	}
	return log
}

// AppErrorf returns an AppError struct created with the supplied data.
func AppErrorf(message string, statusCode int, err error) *AppError {
	return &AppError{
		Message:    message,
		Error:      err,
		StatusCode: statusCode,
	}
}

// AppErrorDefault returns an AppError struct with default values.
func AppErrorDefault(err error) *AppError {
	return &AppError{
		StatusCode: http.StatusInternalServerError,
		Error:      err,
	}
}
