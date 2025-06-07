package options

import (
	"fmt"
	"time"
)

// SetOptions represents options for the SET command
type SetOptions struct {
	*Options
	ExpiryTime time.Time
	ExpiryType string // "EX", "PX", "EXAT", "PXAT", "KEEPTTL"
}

// NewSetOptions creates a new SetOptions instance with predefined options
func NewSetOptions() *SetOptions {
	opts := &SetOptions{
		Options: NewOptions(),
	}

	// Register SET command options with their incompatibility rules
	opts.RegisterOption("NX", "Only set the key if it does not already exist", []string{"XX"})
	opts.RegisterOption("XX", "Only set the key if it already exists", []string{"NX"})
	opts.RegisterOption("GET", "Return the old string stored at key, or nil if key did not exist", nil)

	return opts
}

// IsNX returns true if NX option is set
func (o *SetOptions) IsNX() bool {
	return o.IsSet("NX")
}

// IsXX returns true if XX option is set
func (o *SetOptions) IsXX() bool {
	return o.IsSet("XX")
}

// IsGET returns true if GET option is set
func (o *SetOptions) IsGET() bool {
	return o.IsSet("GET")
}

// IsKEEPTTL returns true if KEEPTTL option is set
func (o *SetOptions) IsKEEPTTL() bool {
	return o.ExpiryType == "KEEPTTL"
}

// SetExpiry sets the expiry time and type
func (o *SetOptions) SetExpiry(expiryType string, value int64) error {
	switch expiryType {
	case "EX":
		o.ExpiryTime = time.Now().Add(time.Duration(value) * time.Second)
		o.ExpiryType = "EX"
	case "PX":
		o.ExpiryTime = time.Now().Add(time.Duration(value) * time.Millisecond)
		o.ExpiryType = "PX"
	case "EXAT":
		o.ExpiryTime = time.Unix(value, 0)
		o.ExpiryType = "EXAT"
	case "PXAT":
		o.ExpiryTime = time.Unix(0, value*int64(time.Millisecond))
		o.ExpiryType = "PXAT"
	case "KEEPTTL":
		o.ExpiryType = "KEEPTTL"
	default:
		return fmt.Errorf("invalid expiry type: %s", expiryType)
	}
	return nil
}
