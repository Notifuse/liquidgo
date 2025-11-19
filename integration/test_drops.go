package integration

import (
	"fmt"
	"strconv"

	"github.com/Notifuse/liquidgo/liquid"
)

// ThingWithToLiquid is a test type that implements ToLiquid.
type ThingWithToLiquid struct{}

// ToLiquid returns a liquid representation.
func (t *ThingWithToLiquid) ToLiquid() interface{} {
	return "foobar"
}

// SettingsDrop is a drop that provides settings access.
type SettingsDrop struct {
	*liquid.Drop
	settings map[string]interface{}
}

// NewSettingsDrop creates a new SettingsDrop.
func NewSettingsDrop(settings map[string]interface{}) *SettingsDrop {
	return &SettingsDrop{
		Drop:     liquid.NewDrop(),
		settings: settings,
	}
}

// LiquidMethodMissing handles missing method calls.
func (s *SettingsDrop) LiquidMethodMissing(key string) interface{} {
	return s.settings[key]
}

// IntegerDrop is a drop that wraps an integer value.
type IntegerDrop struct {
	*liquid.Drop
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
		Drop:  liquid.NewDrop(),
		value: intValue,
	}
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
	*liquid.Drop
	value bool
}

// NewBooleanDrop creates a new BooleanDrop.
func NewBooleanDrop(value bool) *BooleanDrop {
	return &BooleanDrop{
		Drop:  liquid.NewDrop(),
		value: value,
	}
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
	*liquid.Drop
	value string
}

// NewStringDrop creates a new StringDrop.
func NewStringDrop(value string) *StringDrop {
	return &StringDrop{
		Drop:  liquid.NewDrop(),
		value: value,
	}
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
type ErrorDrop struct {
	*liquid.Drop
}

// NewErrorDrop creates a new ErrorDrop.
func NewErrorDrop() *ErrorDrop {
	return &ErrorDrop{
		Drop: liquid.NewDrop(),
	}
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
type TemplateContextDrop struct {
	*liquid.Drop
}

// NewTemplateContextDrop creates a new TemplateContextDrop.
func NewTemplateContextDrop() *TemplateContextDrop {
	return &TemplateContextDrop{
		Drop: liquid.NewDrop(),
	}
}

// LiquidMethodMissing returns the method name.
func (t *TemplateContextDrop) LiquidMethodMissing(method string) interface{} {
	return method
}

// Foo returns a test value.
func (t *TemplateContextDrop) Foo() interface{} {
	return "fizzbuzz"
}

// Baz returns a value from registers.
func (t *TemplateContextDrop) Baz() interface{} {
	if t.Context() != nil {
		return t.Context().Registers().Get("lulz")
	}
	return nil
}

// SomethingWithLength is a drop with a length method that returns nil.
type SomethingWithLength struct {
	*liquid.Drop
}

// NewSomethingWithLength creates a new SomethingWithLength.
func NewSomethingWithLength() *SomethingWithLength {
	return &SomethingWithLength{
		Drop: liquid.NewDrop(),
	}
}

// Length returns nil.
func (s *SomethingWithLength) Length() interface{} {
	return nil
}

// DropWithUndefinedMethod is a drop with a foo method but missing other methods.
type DropWithUndefinedMethod struct {
	*liquid.Drop
}

// NewDropWithUndefinedMethod creates a new DropWithUndefinedMethod.
func NewDropWithUndefinedMethod() *DropWithUndefinedMethod {
	return &DropWithUndefinedMethod{
		Drop: liquid.NewDrop(),
	}
}

// Foo returns "foo".
func (d *DropWithUndefinedMethod) Foo() interface{} {
	return "foo"
}
