package audit

type AuditEventDTO struct {
	TS        int      `json:"ts"`
	Metrics   []string `json:"metrics"`
	IPAddress string   `json:"ip_address"`
}
