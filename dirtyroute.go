/*
* ===== DIRTY ROUTE =====
* Simple Controller->ActionMethod pattern routing
* @atuhor github.com/markomr
* @version 0.1
*/

package dirtyroute

import (
	"net/http"
	"errors"
	"strings"
	"fmt"
	"strconv"
)

type Options struct {
	ContentTypes	[]string
}

type Router struct {
	Controllers 	[]*Controller 		// Slice of pointers to controlers
	Options 		*Options			// Pointer to options
	ErrorHandler	ActionHandler		// error handler action
	AuthHandler		AuthHandler			// auth layer
}

func NewRouter(options *Options) *Router {
	router := Router{}
	router.Options = options
	router.ErrorHandler = defaultErrorHandler
	router.AuthHandler = defaultAuthHandler
	return &router
}

// Register a new controller
type Controller struct {
	Name 		string
	Actions 	[]*Action
}

func (router *Router) RegisterController(c *Controller) {
	router.Controllers = append(router.Controllers, c)
}

// Register Controller Actions
type Action struct {
	Name       		string
	Pattern 		[]string
	Method			string
	Private 		bool
	Handler 		ActionHandler
}

type ActionHandler func(http.ResponseWriter, *http.Request, []string)

// Load the controller action and create a reference in the controller actions slice
func (c *Controller) RegisterAction(a *Action) {
	c.Actions = append(c.Actions, a)
}

// Route To Controller
func (router *Router) Route(w http.ResponseWriter, r *http.Request) {
	var err error

	// Check the content type
	contentType := r.Header.Get("Content-type")
	if contentType == "" { contentType = "Text/Plain" }

	// Check it's enabled
	var cType bool
	for _, t := range router.Options.ContentTypes {
		if strings.ToLower(contentType) == strings.ToLower(t) { cType = true }
	}

	// Return an error if not found
	if !cType {
		err = errors.New("Unsupported content type")
		router.ErrorHandler(w, r, []string{string(http.StatusUnprocessableEntity), err.Error()})
		return
	}

	// Evlauate the url
	params := router.GetParams(r.URL.Path)

	// Get the controller, go through it's actions looking for a pattern match
	var controller *Controller
	controller, err = router.GetController(params.Controller)
	if err != nil {
		router.ErrorHandler(w, r, []string{string(http.StatusNotFound), err.Error()})
		return
	}

	// Check for a pattern match with the controller actions
	// Call the action handler if pattern matched
	for _, a := range controller.Actions {
		mErr := a.Matches(params.Pattern, r.Method)
		if mErr == nil {
			token, err := router.AuthHandler(a, r)
			if token.StatusCode == 0 && err == nil { // 0 indicates continue
				a.Handler(w, r, params.Pattern)
				return
			}
			if token.HandleError && err != nil {
				router.ErrorHandler(w, r, []string{string(token.StatusCode), err.Error()})
				return
			}
		}
	}

	router.ErrorHandler(w, r, []string{string(http.StatusNotFound), err.Error()})
}

// Check the action pattern matches path params and method
func (a *Action) Matches(pattern []string, method string) error {
	var err error
	var matches int
	if len(a.Pattern) == len(pattern) && a.Method == method {
		for i, p := range a.Pattern {
			// Check if an int can be parsed, i.e no characters
			var isint bool
			if _, err := strconv.Atoi(pattern[i]); err == nil { isint = true }
			// Check for a match
			if p == pattern[i] { matches++; continue }							// direct match
			if strings.Contains(p, "{i||s}") { matches++; continue } 			// int or string arg
			if strings.Contains(p, "{i}") && isint { matches++; continue } 		// int arg
			if strings.Contains(p, "{s}") && !isint { matches++; } 				// string arg
		}
	}
	if matches != len(a.Pattern) { err = errors.New("Pattern did not match"); }
	return err
}

// The action params
type Params struct {
	Controller 	string
	Pattern 	[]string
}

// Build Router Params
func (router *Router) GetParams(pstr string) Params {
	params := Params{}
	path := strings.Split(pstr, "/")
	for i, p := range path {
		if i == 0 { continue }									// First part of the split is blank
		if p == "" { p = "{/}" } 								// Index Param
		if i == 1 { params.Controller = p						// Setting the controller
		} else { params.Pattern = append(params.Pattern, p) }	// Setting Action Pattern Params
	}
	// Give empty pattern an index default
	if len(params.Pattern) == 0 { params.Pattern = append(params.Pattern, "{/}") }
	return params
}

// Get Controller
func (router *Router) GetController(name string) (*Controller, error) {
	var err error
	// Look for registered controller
	for _, c := range router.Controllers {
		if c.Name == name { return c, err }
	}
	// Not found
	err = errors.New("Controller not found")
	return nil, err
}

/* Router Error Handler Interface*/
func defaultErrorHandler(w http.ResponseWriter, r *http.Request, args []string) {
	fmt.Fprint(w, "Error: STATUS ", args[1], " : ERROR ", args[1])
}

/* Router Auth Interface*/
type AuthHandler func(*Action, *http.Request) (AuthToken, error)

type AuthToken struct {
	HandleError 	bool
	StatusCode		int
}

func defaultAuthHandler(a *Action, r *http.Request) (AuthToken, error) {
	var err error
	token := AuthToken {
		HandleError: true,
	}
	return token, err
}
