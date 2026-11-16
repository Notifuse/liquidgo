package integration

import (
	"fmt"
	"strconv"

	"github.com/Notifuse/liquidgo/v5/liquid"
)

// ThingWithToLiquid is a test type that implements ToLiquid.
type ThingWithToLiquid struct{}

// ToLiquid returns a liquid representation.
func (t *ThingWithToLiquid) ToLiquid() interface{} {
	return "foobar"
}

// SettingsDrop is a drop that provides settings access.
type SettingsDrop struct {
	settings map[string]interface{}
}

// NewSettingsDrop creates a new SettingsDrop.
func NewSettingsDrop(settings map[string]interface{}) *SettingsDrop {
	return &SettingsDrop{
		settings: settings,
	}
}

// LiquidGet returns any key from the settings, as Ruby's
// liquid_method_missing.
func (s *SettingsDrop) LiquidGet(_ *liquid.Context, key string) (interface{}, bool, error) {
	return s.settings[key], true, nil
}

// IntegerDrop is a drop that wraps an integer value.
type IntegerDrop struct {
	value int
}

// NewIntegerDrop creates a new IntegerDrop.
func NewIntegerDrop(value interface{}) *IntegerDrop {
	var intValue int
	switch v := value.(type) {
	case int:
		intValue = v
	case string:
		var err error
		intValue, err = strconv.Atoi(v)
		if err != nil {
			intValue = 0
		}
	default:
		intValue = 0
	}
	return &IntegerDrop{
		value: intValue,
	}
}

// LiquidGet has no key.
func (i *IntegerDrop) LiquidGet(*liquid.Context, string) (interface{}, bool, error) {
	return nil, false, nil
}

// ToLiquidValue returns the integer value.
func (i *IntegerDrop) ToLiquidValue() interface{} {
	return i.value
}

// String returns the string representation.
func (i *IntegerDrop) String() string {
	return strconv.Itoa(i.value)
}

// BooleanDrop is a drop that wraps a boolean value.
type BooleanDrop struct {
	value bool
}

// NewBooleanDrop creates a new BooleanDrop.
func NewBooleanDrop(value bool) *BooleanDrop {
	return &BooleanDrop{
		value: value,
	}
}

// LiquidGet has no key.
func (b *BooleanDrop) LiquidGet(*liquid.Context, string) (interface{}, bool, error) {
	return nil, false, nil
}

// ToLiquidValue returns the boolean value.
func (b *BooleanDrop) ToLiquidValue() interface{} {
	return b.value
}

// String returns the string representation.
func (b *BooleanDrop) String() string {
	if b.value {
		return "Yay"
	}
	return "Nay"
}

// StringDrop is a drop that wraps a string value.
type StringDrop struct {
	value string
}

// NewStringDrop creates a new StringDrop.
func NewStringDrop(value string) *StringDrop {
	return &StringDrop{
		value: value,
	}
}

// LiquidGet has no key.
func (s *StringDrop) LiquidGet(*liquid.Context, string) (interface{}, bool, error) {
	return nil, false, nil
}

// ToLiquidValue returns the string value.
func (s *StringDrop) ToLiquidValue() interface{} {
	return s.value
}

// String returns the string representation.
func (s *StringDrop) String() string {
	return s.value
}

// ErrorDrop is a drop that raises various types of errors for testing.
type ErrorDrop struct{}

// NewErrorDrop creates a new ErrorDrop.
func NewErrorDrop() *ErrorDrop {
	return &ErrorDrop{}
}

// LiquidGet calls the method of key, which raises.
func (e *ErrorDrop) LiquidGet(_ *liquid.Context, key string) (interface{}, bool, error) {
	switch key {
	case "standard_error":
		return e.StandardError(), true, nil
	case "argument_error":
		return e.ArgumentError(), true, nil
	case "syntax_error":
		return e.SyntaxError(), true, nil
	case "runtime_error":
		return e.RuntimeError(), true, nil
	case "exception":
		return e.Exception(), true, nil
	}
	return nil, false, nil
}

// StandardError raises a StandardError.
func (e *ErrorDrop) StandardError() interface{} {
	panic(liquid.NewStandardError("standard error"))
}

// ArgumentError raises an ArgumentError.
func (e *ErrorDrop) ArgumentError() interface{} {
	panic(liquid.NewArgumentError("argument error"))
}

// SyntaxError raises a SyntaxError.
func (e *ErrorDrop) SyntaxError() interface{} {
	panic(liquid.NewSyntaxError("syntax error"))
}

// RuntimeError raises a runtime error.
func (e *ErrorDrop) RuntimeError() interface{} {
	panic("runtime error")
}

// Exception raises a generic exception.
func (e *ErrorDrop) Exception() interface{} {
	panic(fmt.Errorf("exception"))
}

// TemplateContextDrop is a drop that can access the template context.
type TemplateContextDrop struct{}

// NewTemplateContextDrop creates a new TemplateContextDrop.
func NewTemplateContextDrop() *TemplateContextDrop {
	return &TemplateContextDrop{}
}

// ContextProxyDrop is a proxy for context access.
type ContextProxyDrop struct{}

// LiquidGet reads key in the context of the render.
func (c *ContextProxyDrop) LiquidGet(ctx *liquid.Context, key string) (interface{}, bool, error) {
	return ctx.Get(key), true, nil
}

// LiquidGet reads the methods of the drop, then key in the context of the
// render, or gives key back.
func (t *TemplateContextDrop) LiquidGet(ctx *liquid.Context, key string) (interface{}, bool, error) {
	switch key {
	case "scopes":
		return len(ctx.Scopes()), true, nil
	case "loop_pos":
		if forloop, ok := ctx.Get("forloop").(liquid.LiquidGetter); ok {
			return forloop.LiquidGet(ctx, "index")
		}
		return nil, true, nil
	case "foo":
		return "fizzbuzz", true, nil
	case "baz":
		return ctx.Registers().Get("lulz"), true, nil
	case "context":
		return &ContextProxyDrop{}, true, nil
	}
	if result := ctx.Get(key); result != nil {
		return result, true, nil
	}
	return key, true, nil
}

// SomethingWithLength is a drop with a length method that returns nil.
type SomethingWithLength struct{}

// NewSomethingWithLength creates a new SomethingWithLength.
func NewSomethingWithLength() *SomethingWithLength {
	return &SomethingWithLength{}
}

// LiquidGet has the key length, nil.
func (s *SomethingWithLength) LiquidGet(_ *liquid.Context, key string) (interface{}, bool, error) {
	return nil, key == "length", nil
}

// DropWithUndefinedMethod is a drop with a foo method but missing other methods.
type DropWithUndefinedMethod struct{}

// NewDropWithUndefinedMethod creates a new DropWithUndefinedMethod.
func NewDropWithUndefinedMethod() *DropWithUndefinedMethod {
	return &DropWithUndefinedMethod{}
}

// LiquidGet has the key foo.
func (d *DropWithUndefinedMethod) LiquidGet(_ *liquid.Context, key string) (interface{}, bool, error) {
	if key == "foo" {
		return "foo", true, nil
	}
	return nil, false, nil
}
