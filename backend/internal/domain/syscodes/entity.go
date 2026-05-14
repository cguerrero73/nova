package syscodes

import "time"

type SysCode struct {
	ID        string    `json:"sys_id"`
	Type      string    `json:"sys_type"`
	Code      string    `json:"sys_code"`
	UCode     string    `json:"sys_ucode"`
	Desc      string    `json:"sys_desc"`
	System    *string   `json:"sys_system,omitempty"`
	NotUsed   *string   `json:"sys_notused,omitempty"`
	CreatedAt time.Time `json:"sys_created_at"`
	UpdatedAt time.Time `json:"sys_updated_at"`
}

func (s *SysCode) IsSystem() bool {
	return s.System != nil && *s.System == "+"
}

func (s *SysCode) IsActive() bool {
	return s.NotUsed == nil || *s.NotUsed != "+"
}

// GetUCode returns the user-facing code
func (s *SysCode) GetUCode() string {
	if s.UCode != "" {
		return s.UCode
	}
	return s.Code
}