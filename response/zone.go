package response

type ZoneCollectionRecord struct {
    Name        string  `json:"name"`
    Type        string  `json:"type"`
    Ttl         int     `json:"ttl"`
    Disabled    bool    `json:"disabled"`
    Content     string  `json:"content"`
}

type ZoneCollection struct {
    Id                  string                  `json:"id"`
    Name                string                  `json:"name"`
    Type                string                  `json:"type"`
    Url                 string                  `json:"url"`
    Kind                string                  `json:"kind"`
    Serial              int                     `json:"serial"`
    Notified_serial     int                     `json:"notified_serial"`
    Masters             []string                `json:"masters"`
    Dnssec              bool                    `json:"dnssec"`
    Nsec3param          string                  `json:"nsec3param"`
    Nsec3narrow         bool                    `json:"nsec3narrow"`
    Presigned           bool                    `json:"presigned"`
    Soa_edit            string                  `json:"soa_edit"`
    Soa_edit_api        string                  `json:"soa_edit_api"`
    Account             string                  `json:"account"`
    Nameservers         []string                `json:"nameservers"`
    Servers             []string                `json:"servers"`
    Recursion_desired   bool                    `json:"recursion_desired"`
    Records             []ZoneCollectionRecord  `json:"records"`
    Comments            []string                `json:"comments"`
}