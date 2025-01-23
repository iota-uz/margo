package registry

import (
	"errors"
	"github.com/a-h/templ"
	"reflect"
	"strings"
)

type Layout interface {
	Name() string
	Namespace(name string) (Layout, error)
	Register(name string, component any) Layout
	Clone() Layout
	Get(name string) (any, bool)
	List() []string
}

type Registry interface {
	RegisterLayout(layout Layout) Registry
	Use(name string) (Layout, bool)
	Clone() Registry
	Layouts() []string
}

func New() Registry {
	return &registry{
		layouts: make(map[string]Layout),
	}
}

type registry struct {
	layouts map[string]Layout
}

func (r *registry) RegisterLayout(layout Layout) Registry {
	r.layouts[strings.ToLower(layout.Name())] = layout
	return r
}

func (r *registry) Clone() Registry {
	layouts := make(map[string]Layout, len(r.layouts))
	for k, v := range r.layouts {
		layouts[k] = v.Clone()
	}
	return &registry{
		layouts: layouts,
	}
}

func (r *registry) Use(name string) (Layout, bool) {
	v, ok := r.layouts[strings.ToLower(name)]
	if !ok {
		return nil, false
	}
	return v, true
}

func (r *registry) Layouts() []string {
	var res []string
	for k := range r.layouts {
		res = append(res, k)
	}
	return res
}

type item struct {
	value     any
	subLayout Layout
}

// NewLayout creates a new layout with the given name.
func NewLayout(name string) Layout {
	return &layout{
		name:       name,
		components: make(map[string]*item),
	}
}

type layout struct {
	name       string
	components map[string]*item
}

func (r *layout) Name() string {
	return r.name
}

func (r *layout) Namespace(name string) (Layout, error) {
	n := strings.ToLower(name)
	v, ok := r.components[n]
	if !ok {
		return nil, errors.New("namespace not found")
	}
	return v.subLayout, nil
}

// Register registers a component with the given name.
// The component signature must be func(struct) templ.Component.
func (r *layout) Register(name string, component any) Layout {
	n := strings.ToLower(name)
	t := reflect.TypeOf(component)
	if t.Kind() != reflect.Func {
		panic("component must be a function")
	}
	if t.NumIn() == 1 && t.In(0).Kind() != reflect.Struct {
		panic("component must take in a single struct")
	}
	if t.NumOut() != 1 {
		panic("component must return a single value")
	}
	if t.Out(0) != reflect.TypeOf((*templ.Component)(nil)).Elem() {
		panic("component must return templ.Component")
	}
	subLayout := NewLayout(name)
	r.components[n] = &item{
		value:     component,
		subLayout: subLayout,
	}
	return subLayout
}

func (r *layout) Clone() Layout {
	components := make(map[string]*item, len(r.components))
	for k, v := range r.components {
		components[k] = &item{
			value:     v.value,
			subLayout: v.subLayout.Clone(),
		}
	}
	return &layout{
		name:       r.name,
		components: components,
	}
}

func (r *layout) Get(name string) (any, bool) {
	v, ok := r.components[strings.ToLower(name)]
	if !ok {
		return nil, false
	}
	return v.value, true
}

func (r *layout) List() []string {
	var res []string
	for k := range r.components {
		res = append(res, k)
	}
	return res
}
