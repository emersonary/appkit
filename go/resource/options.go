package resource

// DescribeOption customizes schema generation.
type DescribeOption func(*describeConfig)

type describeConfig struct {
	name        string
	label       string
	mode        Mode
	overrides   map[string]func(*Field)
	extraFields []Field
}

func WithName(name string) DescribeOption {
	return func(cfg *describeConfig) {
		cfg.name = name
	}
}

func WithLabel(label string) DescribeOption {
	return func(cfg *describeConfig) {
		cfg.label = label
	}
}

func WithMode(mode Mode) DescribeOption {
	return func(cfg *describeConfig) {
		cfg.mode = mode
	}
}

func WithExtraFields(fields ...Field) DescribeOption {
	return func(cfg *describeConfig) {
		cfg.extraFields = append(cfg.extraFields, fields...)
	}
}

func FieldOverride(key string, fn func(*Field)) DescribeOption {
	return func(cfg *describeConfig) {
		if cfg.overrides == nil {
			cfg.overrides = make(map[string]func(*Field))
		}
		cfg.overrides[key] = fn
	}
}

func applyDescribeOptions(opts []DescribeOption) describeConfig {
	cfg := describeConfig{
		mode: ModeEditOnly,
	}
	for _, opt := range opts {
		opt(&cfg)
	}
	return cfg
}
