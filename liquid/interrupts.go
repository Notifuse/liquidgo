package liquid

// Interrupt is any command that breaks processing of a block (ex: a for loop).
type Interrupt struct {
	Message string
}

// NewInterrupt creates a new interrupt with the given message.
func NewInterrupt(message string) *Interrupt {
	if message == "" {
		message = "interrupt"
	}
	return &Interrupt{Message: message}
}

// BreakInterrupt is thrown whenever a {% break %} is called.
type BreakInterrupt struct {
	*Interrupt
}

// NewBreakInterrupt creates a new break interrupt.
func NewBreakInterrupt() *BreakInterrupt {
	return &BreakInterrupt{
		Interrupt: NewInterrupt("break"),
	}
}

// ContinueInterrupt is thrown whenever a {% continue %} is called.
type ContinueInterrupt struct {
	*Interrupt
}

// NewContinueInterrupt creates a new continue interrupt.
func NewContinueInterrupt() *ContinueInterrupt {
	return &ContinueInterrupt{
		Interrupt: NewInterrupt("continue"),
	}
}
