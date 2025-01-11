package config

// RegionConfig holds region-specific parameters
type RegionConfig struct {
	Gl string // Country code
	Lr string // Language region
	Hl string // Host language
}

// RegionConfigs maps region codes to their configurations
var RegionConfigs = map[string]RegionConfig{
	"in-en": {"in", "lang_en", "en-IN"}, // India (English
	"in-hi": {"in", "lang_hi", "hi-IN"}, // India (Hindi
	"in-bn": {"in", "lang_bn", "bn-IN"}, // India (Bengali
	"in-te": {"in", "lang_te", "te-IN"}, // India (Telugu
	"in-ta": {"in", "lang_ta", "ta-IN"}, // India (Tamil
	"us":    {"us", "lang_en", "en-US"}, // United
	"uk":    {"gb", "lang_en", "en-GB"}, // United
}
