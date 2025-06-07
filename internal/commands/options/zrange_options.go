package options

import "fmt"

// ZRangeOptions represents options for the ZRANGE command
type ZRangeOptions struct {
	*Options
	RangeType string // "BYSCORE", "BYLEX", or "" for index-based range
	Rev       bool
	Limit     struct {
		Offset int
		Count  int
	}
	WithScores bool
}

// NewZRangeOptions creates a new ZRangeOptions instance with predefined options
func NewZRangeOptions() *ZRangeOptions {
	opts := &ZRangeOptions{
		Options: NewOptions(),
	}

	// Register ZRANGE command options with their incompatibility rules
	opts.RegisterOption("BYSCORE", "Return elements with scores between min and max", []string{"BYLEX"})
	opts.RegisterOption("BYLEX", "Return elements with lexicographical ordering", []string{"BYSCORE"})
	opts.RegisterOption("REV", "Reverse the order of returned elements", nil)
	opts.RegisterOption("WITHSCORES", "Return scores along with members", nil)

	return opts
}

// IsByScore returns true if BYSCORE option is set
func (o *ZRangeOptions) IsByScore() bool {
	return o.RangeType == "BYSCORE"
}

// IsByLex returns true if BYLEX option is set
func (o *ZRangeOptions) IsByLex() bool {
	return o.RangeType == "BYLEX"
}

// IsRev returns true if REV option is set
func (o *ZRangeOptions) IsRev() bool {
	return o.Rev
}

// IsWithScores returns true if WITHSCORES option is set
func (o *ZRangeOptions) IsWithScores() bool {
	return o.WithScores
}

// SetRangeType sets the range type (BYSCORE or BYLEX)
func (o *ZRangeOptions) SetRangeType(rangeType string) error {
	switch rangeType {
	case "BYSCORE", "BYLEX":
		o.RangeType = rangeType
		return nil
	default:
		return fmt.Errorf("invalid range type: %s", rangeType)
	}
}

// SetLimit sets the LIMIT parameters
func (o *ZRangeOptions) SetLimit(offset, count int) error {
	if offset < 0 {
		return fmt.Errorf("offset must be non-negative")
	}
	if count < 0 {
		return fmt.Errorf("count must be non-negative")
	}
	o.Limit.Offset = offset
	o.Limit.Count = count
	return nil
}
