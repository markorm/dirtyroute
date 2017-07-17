# Dirtyroute
A simple Controller -> Action/Method router for Go.

## Usage
`````javascript
func init() {
    // Router Config
    options := dirtyroute.Options {
        ContentTypes: []string {
            "text/plain",
            "application/json",
        },
    }

    // Get a router
    router = dirtyroute.NewRouter(&options)

    // Middleware for error handling
    router.ErrorHandler = myErrorHandler

    // Middleware for authorization
    router.AuthHandler = myAuthHandler

    // Register Controllers
    router.RegisterController(controllers.IndexController())
    router.RegisterController(controllers.PostsController())
}

func main() {
    route := http.HandlerFunc(router.Route)
    http.Handle("/", myHttpMiddleware(route))
    http.ListenAndServe(":9090", nil)
}
`````

## Controller and Action Configuration

Define your controllers as shown below

`````javascript

// Controller
func PostsController() *dirtyroute.Controller {
	c := dirtyroute.Controller{}
	c.Name = "posts"
	c.RegisterAction(&GetPosts)
	c.RegisterAction(&GetPost)
	c.RegisterAction(&CreatePost)
	c.RegisterAction(&RemovePost)
	c.RegisterAction(&UpdatePost)
	return &c
}


// Action
var GetPosts = dirtyroute.Action {
	Pattern: []string{"{/}"},
	Method:  "GET",
	Handler: func(w http.ResponseWriter, r *http.Request, args []string) {
	    // Your Action Logic
	},
}

// Action
var GetAuthor = dirtyroute.Action {
	Pattern: []string{"{s}","{i}"},
	Method:  "POST",
	Handler: func(w http.ResponseWriter, r *http.Request, args []string) {
	    // Your Action Logic
	},
}
`````

The controller name is represented by the first part of the url path,
i.e mydomain.com/posts, the action patterns match the following:

The values of these matching path paramters will be provided to the action
in the args []string argument

{/} : Matches the index
{i} : Match any int
{s} : Match any string
{i||s} : Match int or string
{word} : Match any word

using these you can achieve the following:

// Eg view details of the author of a post
controller/id/author = {i}{s}

// View all posts by authorid
controller/author/id = {s}{i}

// view author top posts
controller/author/id/posts/top = {s}{i}{s}{s}


The actions also must match on HTTP verbs, you can use any you like.