package options

// ExpireOptions represents options for the EXPIRE command
type ExpireOptions struct {
	*Options
}

// NewExpireOptions creates a new ExpireOptions instance with predefined options
func NewExpireOptions() *ExpireOptions {
	opts := &ExpireOptions{
		Options: NewOptions(),
	}

	// Register EXPIRE command options with their incompatibility rules
	opts.RegisterOption("NX", "Set expiry only if the key has no expiry", []string{"XX", "GT", "LT"})
	opts.RegisterOption("XX", "Set expiry only if the key has an existing expiry", []string{"NX", "GT", "LT"})
	opts.RegisterOption("GT", "Set expiry only if the new expiry is greater than current one", []string{"NX", "XX", "LT"})
	opts.RegisterOption("LT", "Set expiry only if the new expiry is less than current one", []string{"NX", "XX", "GT"})

	return opts
}

// IsNX returns true if NX option is set
func (o *ExpireOptions) IsNX() bool {
	return o.IsSet("NX")
}

// IsXX returns true if XX option is set
func (o *ExpireOptions) IsXX() bool {
	return o.IsSet("XX")
}

// IsGT returns true if GT option is set
func (o *ExpireOptions) IsGT() bool {
	return o.IsSet("GT")
}

// IsLT returns true if LT option is set
func (o *ExpireOptions) IsLT() bool {
	return o.IsSet("LT")
}
