package renderer

import (
	"context"
	"fmt"
	"github.com/a-h/templ"
	"github.com/iota-uz/margo"
	"github.com/iota-uz/margo/parser"
	"github.com/iota-uz/margo/registry"
	"github.com/yuin/goldmark/ast"
	"io"
	"reflect"
	"slices"
	"strconv"
	"strings"
)

type ValueSetter struct{}

func NewValueSetter() *ValueSetter {
	return &ValueSetter{}
}

func (vs *ValueSetter) SetTypedValue(field reflect.Value, value interface{}) error {
	switch field.Kind() {
	case reflect.String:
		switch v := value.(type) {
		case string:
			field.SetString(v)
		case []byte:
			field.SetString(string(v))
		default:
			return fmt.Errorf("unsupported value type for string: %T", value)
		}
	case reflect.Bool:
		field.Set(reflect.ValueOf(value))
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return vs.setIntValue(field, value)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return vs.setUintValue(field, value)
	case reflect.Float32, reflect.Float64:
		return vs.setFloatValue(field, value)
	default:
		return fmt.Errorf("unsupported field type: %s", field.Kind())
	}
	return nil
}

func (vs *ValueSetter) setIntValue(field reflect.Value, value interface{}) error {
	i, err := strconv.ParseInt(value.(string), 10, 64)
	if err != nil {
		return fmt.Errorf("failed to parse int: %w", err)
	}
	field.SetInt(i)
	return nil
}

func (vs *ValueSetter) setUintValue(field reflect.Value, value interface{}) error {
	i, err := strconv.ParseUint(value.(string), 10, 64)
	if err != nil {
		return fmt.Errorf("failed to parse uint: %w", err)
	}
	field.SetUint(i)
	return nil
}

func (vs *ValueSetter) setFloatValue(field reflect.Value, value interface{}) error {
	f, err := strconv.ParseFloat(value.(string), 64)
	if err != nil {
		return fmt.Errorf("failed to parse float: %w", err)
	}
	field.SetFloat(f)
	return nil
}

type ComponentBuilder struct {
	layout    registry.Layout
	valSetter *ValueSetter
}

func NewComponentBuilder(reg registry.Layout) *ComponentBuilder {
	return &ComponentBuilder{
		layout:    reg,
		valSetter: NewValueSetter(),
	}
}

func (cb *ComponentBuilder) Build(component any, attrs []ast.Attribute) (templ.Component, error) {
	reflectV := reflect.ValueOf(component)
	props, err := cb.buildProps(reflectV, attrs)
	if err != nil {
		return nil, err
	}
	return reflectV.Call(props)[0].Interface().(templ.Component), nil
}

func (cb *ComponentBuilder) buildProps(componentFunc reflect.Value, attrs []ast.Attribute) ([]reflect.Value, error) {
	if componentFunc.Type().NumIn() == 0 {
		return []reflect.Value{}, nil
	}

	propsType := componentFunc.Type().In(0)
	props := reflect.New(propsType).Elem()
	var used []string
	for i := 0; i < propsType.NumField(); i++ {
		field := props.Field(i)
		if !field.IsValid() || !field.CanSet() {
			continue
		}
		for _, attr := range attrs {
			k := string(attr.Name)
			if strings.EqualFold(k, propsType.Field(i).Name) {
				used = append(used, k)
				if field.Kind() == reflect.Interface {
					if err := cb.setInterfaceValue(field, attr.Value); err != nil {
						return nil, err
					}
				} else {
					if err := cb.valSetter.SetTypedValue(field, attr.Value); err != nil {
						return nil, err
					}
				}
			}
		}
	}

	componentAttrs := templ.Attributes{}
	for _, attr := range attrs {
		if !slices.Contains(used, string(attr.Name)) {
			key := string(attr.Name)
			if v, ok := attr.Value.(string); ok {
				componentAttrs[key] = v
			} else {
				componentAttrs[key] = string(attr.Value.([]byte))
			}
		}
	}

	for i := 0; i < propsType.NumField(); i++ {
		field := props.Field(i)
		if !field.IsValid() || !field.CanSet() {
			continue
		}
		if field.Kind() == reflect.Map && field.Type().String() == "templ.Attributes" {
			field.Set(reflect.ValueOf(componentAttrs))
			for k := range componentAttrs {
				used = append(used, k)
			}
		}
	}

	if len(used) < len(attrs) {
		var errorMessage strings.Builder
		errorMessage.WriteString("unknown prop: ")
		for _, attr := range attrs {
			if !slices.Contains(used, string(attr.Name)) {
				errorMessage.WriteString(string(attr.Name))
				errorMessage.WriteString(" ")
			}
		}
		return nil, fmt.Errorf(errorMessage.String())
	}

	return []reflect.Value{props}, nil
}

func (cb *ComponentBuilder) setInterfaceValue(field reflect.Value, value interface{}) error {
	if field.Type().String() != "templ.Component" {
		return fmt.Errorf("unsupported interface type: %s", field.Type())
	}

	switch v := value.(type) {
	case *parser.ComponentNode:
		return cb.setComponentNodeValue(field, v)
	case *parser.TextNode:
		return cb.setTextNodeValue(field, v)
	case string:
		return cb.setStringValue(field, v)
	default:
		return fmt.Errorf("unsupported value type for interface: %T", value)
	}
}

func (cb *ComponentBuilder) setComponentNodeValue(field reflect.Value, node *parser.ComponentNode) error {
	cmpFunc, err := cb.GetComponent(node.Name, nil)
	if err != nil {
		return err
	}
	cmp, err := cb.Build(cmpFunc, node.Attributes())
	if err != nil {
		return fmt.Errorf("failed to build component %s: %w", node.Name, err)
	}
	field.Set(reflect.ValueOf(cmp))
	return nil
}

func (cb *ComponentBuilder) setTextNodeValue(field reflect.Value, node *parser.TextNode) error {
	field.Set(reflect.ValueOf(templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		return margo.New(cb.layout).Convert([]byte(node.Value), w)
	})))
	return nil
}

func (cb *ComponentBuilder) setStringValue(field reflect.Value, value string) error {
	field.Set(reflect.ValueOf(templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		return margo.New(cb.layout).Convert([]byte(value), w)
	})))
	return nil
}

func (cb *ComponentBuilder) GetComponent(name string, ns registry.Layout) (any, error) {
	if ns != nil {
		if c, ok := ns.Get(name); ok {
			return c, nil
		}
	}
	if c, ok := cb.layout.Get(name); ok {
		return c, nil
	}
	if ns != nil {
		return nil, fmt.Errorf("component %s not found in namespace: %s", name, ns.Name())
	}
	return nil, fmt.Errorf("component %s not found in layout %s", name, cb.layout.Name())
}
