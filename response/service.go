package response

type ServerResource struct {
    Type        string  `json:"type"`
    Id          string  `json:"id"`
    Url         string  `json:"url"`
    Daemon_type string  `json:"daemon_type"`
    Version     string  `json:"version"`
    Config_url  string  `json:"config_url"`
    Zones_url   string  `json:"zones_url"`
}