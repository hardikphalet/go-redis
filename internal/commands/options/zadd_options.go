package options

// ZAddOptions represents options for the ZADD command
type ZAddOptions struct {
	*Options
}

// NewZAddOptions creates a new ZAddOptions instance with predefined options
func NewZAddOptions() *ZAddOptions {
	opts := &ZAddOptions{
		Options: NewOptions(),
	}

	// Register ZADD command options with their incompatibility rules
	opts.RegisterOption("NX", "Only add new elements, don't update already existing elements", []string{"XX"})
	opts.RegisterOption("XX", "Only update elements that already exist, don't add new elements", []string{"NX"})
	opts.RegisterOption("GT", "Only update existing elements if the new score is greater than the current score", []string{"LT"})
	opts.RegisterOption("LT", "Only update existing elements if the new score is less than the current score", []string{"GT"})
	opts.RegisterOption("CH", "Modify the return value to return the number of changed elements instead of new elements", nil)
	opts.RegisterOption("INCR", "Increment the score of an element instead of setting it", []string{"NX", "XX", "GT", "LT"})

	return opts
}

// IsNX returns true if NX option is set
func (o *ZAddOptions) IsNX() bool {
	return o.IsSet("NX")
}

// IsXX returns true if XX option is set
func (o *ZAddOptions) IsXX() bool {
	return o.IsSet("XX")
}

// IsGT returns true if GT option is set
func (o *ZAddOptions) IsGT() bool {
	return o.IsSet("GT")
}

// IsLT returns true if LT option is set
func (o *ZAddOptions) IsLT() bool {
	return o.IsSet("LT")
}

// IsCH returns true if CH option is set
func (o *ZAddOptions) IsCH() bool {
	return o.IsSet("CH")
}

// IsINCR returns true if INCR option is set
func (o *ZAddOptions) IsINCR() bool {
	return o.IsSet("INCR")
}
