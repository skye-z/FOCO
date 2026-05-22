package platform

import "time"

type AdminSetting struct {
	Key       string    `xorm:"'key' pk"`
	ValueJSON string    `xorm:"'value_json' text notnull"`
	UpdatedAt time.Time `xorm:"'updated_at' notnull"`
}

func (AdminSetting) TableName() string { return "admin_settings" }

type LLMSettings struct {
	Provider string `json:"provider"`
	BaseURL  string `json:"base_url"`
	APIKey   string `json:"api_key,omitempty"`
	Model    string `json:"model"`
	Enabled  bool   `json:"enabled"`
}

type LLMSettingsSummary struct {
	Provider   string `json:"provider"`
	BaseURL    string `json:"base_url"`
	Model      string `json:"model"`
	Enabled    bool   `json:"enabled"`
	Configured bool   `json:"configured"`
}

type AdminSettings struct {
	LLM              LLMSettingsSummary `json:"llm"`
	RegistrationOpen bool               `json:"registration_open"`
}

type AdminSettingsUpdate struct {
	LLM              LLMSettings `json:"llm"`
	RegistrationOpen bool        `json:"registration_open"`
}

type PublicSettings struct {
	RegistrationOpen bool `json:"registration_open"`
}
