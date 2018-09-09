// Package view provides thread-safe caching of HTML templates.
//
// Original taken from https://github.com/blue-jay/core.
package view

import (
	"context"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/oxtoacart/bpool"
)

// Buffer pool for output
var templateBufPool *bpool.BufferPool

// Template holds the root and children templates.
type Template struct {
	Root     string   `envae:"root_template"`
	Children []string `envae:"child_templates"`
}

// Mutexes which manage concurrent access to collections in the Info struct
var extendMutex sync.RWMutex
var modifyMutex sync.RWMutex
var templateCacheMutex sync.RWMutex

// Info holds view attributes.
type Info struct {
	BaseURI   string `envae:"view_base_uri"`
	Extension string `envae:"view_extension"`
	Folder    string `envae:"view_directory"`
	Caching   bool   `envae:"view_caching"`

	Data       interface{}
	flashes    []string // TODO: severity level
	user       interface{}
	dateFormat string

	base      string
	templates []string

	childTemplates []string
	rootTemplate   string

	extendList template.FuncMap
	modifyList []ModifyFunc

	templateCollection map[string]*template.Template
}

// viewModel holds the standard data always passed in to a view along with the
// page specific data
type viewModel struct {
	User       interface{}
	PageName   string
	Flashes    []string
	Data       interface{}
	DateFormat string
}

// *****************************************************************************
// Template Handling
// *****************************************************************************

// New accepts multiple templates and then returns a new view.
//
// TODO: Actually return a new view struct here rather than relying on the
// appenv.GetEnv() function to copy it.
func (v *Info) New(templateList ...string) *Info {
	v.templates = append(v.templates, templateList...)
	v.base = v.rootTemplate

	return v
}

// AddFlash adds a flash message to the data passed to the template.
func (v *Info) AddFlash(f string) {
	v.flashes = append(v.flashes, f)
}

// SetUser sets the user object for the view data.
func (v *Info) SetUser(u interface{}) {
	v.user = u
}

// SetDateFormat sets the date format for the view.
func (v *Info) SetDateFormat(f string) {
	v.dateFormat = f
}

// Base sets the new base template instead of reading from
// Template.Root of the config file.
func (v *Info) Base(base string) *Info {
	// Set the new base template
	v.base = base

	// Allow chaining
	return v
}

// Render parses one or more templates and outputs to the screen.
// Also returns an error if anything is wrong.
func (v *Info) Render(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	// Initialise the output buffer pool if necessary
	if templateBufPool == nil {
		templateBufPool = bpool.NewBufferPool(32)
	}

	// Use the first template supplied to New() as the page name
	// Sanitise for use as CSS class
	pageName := strings.Replace(v.templates[0], "/", "-", -1)

	// Add the base template
	v.templates = append([]string{v.base}, v.templates...)

	// Add the child templates
	v.templates = append(v.templates, v.childTemplates...)

	// Set the base template
	baseTemplate := v.templates[0]

	// Set the key name for caching
	key := strings.Join(v.templates, ":")

	// Get the template collection from cache
	templateCacheMutex.RLock()
	tc, ok := v.templateCollection[key]
	templateCacheMutex.RUnlock()

	// Get the extend list
	pc := v.extend()

	// If the template collection is not cached or caching is disabled
	if !ok || !v.Caching {
		// Loop through each template and test the full path
		for i, name := range v.templates {
			// Get the absolute path of the root template
			path, err := filepath.Abs(v.Folder + string(os.PathSeparator) + name + "." + v.Extension)
			if err != nil {
				return fmt.Errorf("view: template path error: %v", err.Error())
			}
			// Store the full template path
			v.templates[i] = path
		}

		// Determine if there is an error in the template syntax
		templates, err := template.New(key).Funcs(pc).ParseFiles(v.templates...)
		if err != nil {
			return fmt.Errorf("view: template parse error: %s, %s", err.Error(), v.templates)
		}

		// Cache the template collection
		templateCacheMutex.Lock()
		v.templateCollection[key] = templates
		templateCacheMutex.Unlock()

		// Save the template collection
		tc = templates
	}

	// Get the modify list
	sc := v.modify()

	// Loop through and call each one
	for _, fn := range sc {
		fn(w, r, v)
	}

	// Build the view model
	vm := viewModel{
		Data:       v.Data,
		User:       v.user,
		PageName:   pageName,
		Flashes:    v.flashes,
		DateFormat: v.dateFormat,
	}

	// Render the output to a buffer, check for errors, render buffer to screen
	buf := templateBufPool.Get()
	defer templateBufPool.Put(buf)
	err := tc.Funcs(pc).ExecuteTemplate(buf, baseTemplate+"."+v.Extension, vm)
	if err != nil {
		return fmt.Errorf("view: template render error: %v", err.Error())
	}
	buf.WriteTo(w)

	return nil
}
