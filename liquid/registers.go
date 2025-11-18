package liquid

// Registers provides a registry for template-level data.
// It supports both static (immutable) and dynamic (mutable) values.
type Registers struct {
	static  map[string]interface{}
	changes map[string]interface{}
}

// NewRegisters creates a new Registers instance.
func NewRegisters(registers interface{}) *Registers {
	var static map[string]interface{}

	if r, ok := registers.(*Registers); ok {
		static = r.static
	} else if m, ok := registers.(map[string]interface{}); ok {
		static = m
	} else {
		static = make(map[string]interface{})
	}

	return &Registers{
		static:  static,
		changes: make(map[string]interface{}),
	}
}

// Set sets a value in the registers (creates a change).
func (r *Registers) Set(key string, value interface{}) {
	r.changes[key] = value
}

// Get gets a value from the registers (checks changes first, then static).
func (r *Registers) Get(key string) interface{} {
	if val, ok := r.changes[key]; ok {
		return val
	}
	return r.static[key]
}

// Delete deletes a key from changes.
func (r *Registers) Delete(key string) {
	delete(r.changes, key)
}

// Fetch gets a value with a default or block.
func (r *Registers) Fetch(key string, defaultValue interface{}) interface{} {
	if val, ok := r.changes[key]; ok {
		return val
	}
	if val, ok := r.static[key]; ok {
		return val
	}
	return defaultValue
}

// HasKey returns true if the key exists in either changes or static.
func (r *Registers) HasKey(key string) bool {
	_, hasChange := r.changes[key]
	_, hasStatic := r.static[key]
	return hasChange || hasStatic
}

// Static returns the static registers map.
func (r *Registers) Static() map[string]interface{} {
	return r.static
}

// Changes returns the changes map.
func (r *Registers) Changes() map[string]interface{} {
	return r.changes
}
