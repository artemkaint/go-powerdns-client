package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/miekg/dns"
	"github.com/skynetservices/skydns1/msg"
	"github.com/artemkaint/docker-powerdns-dock/powerdns/response"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"os"
)

var (
	ErrNoHttpAddress   = errors.New("No HTTP address specified")
	ErrNoDnsAddress    = errors.New("No DNS address specified")
	ErrInvalidResponse = errors.New("Invalid HTTP response")
	ErrServiceNotFound = errors.New("Service not found")
	ErrConflictingUUID = errors.New("Conflicting UUID")
)

type (
	Client struct {
		base    string
		secret  string
		h       *http.Client
		basedns string
		domain  string
		d       *dns.Client
		DNS     bool // if true use the DNS when listing servies
	}

	NameCount map[string]int
)

func logger(msg string, buffer *bytes.Buffer) error {
	fmt.Fprintf(os.Stderr, "Log: Method:%s \n %s \n", msg, buffer)
	return nil
}

// NewClient creates a new skydns client with the specificed host address and
// DNS port.
func NewClient(base, secret, domain, basedns string) (*Client, error) {
	if base == "" {
		return nil, ErrNoHttpAddress
	}
	if basedns == "" {
		return nil, ErrNoDnsAddress
	}
	return &Client{
		base:    base,
		basedns: basedns,
		domain:  dns.Fqdn(domain),
		secret:  secret,
		h:       &http.Client{},
		d:       &dns.Client{},
	}, nil
}

// GET /servers
func (c *Client) GetServers() (*[]response.ServerResource, error) {
    b := bytes.NewBuffer(nil)
    req, err := c.newRequest("GET", c.joinUrl("servers"), b)
    if err != nil {
        return nil, err
    }
    resp, err := c.h.Do(req)
    if err != nil {
        return nil, err
    }
    if resp.Body != nil {
        defer resp.Body.Close()
    }
    switch resp.StatusCode {
    	case http.StatusOK:
    		break
    	default:
    		return nil, ErrInvalidResponse
    }
    var s *[]response.ServerResource
    if err := json.NewDecoder(resp.Body).Decode(&s); err != nil {
        return nil, err
    }
    return s, nil
}

// GET /servers/:server_id
func (c *Client) GetServer(uuid string) (*response.ServerResource, error) {
    b := bytes.NewBuffer(nil)
    req, err := c.newRequest("GET", c.joinUrl(fmt.Sprintf("servers/%s", uuid)), b)
    if err != nil {
        return nil, err
    }
    resp, err := c.h.Do(req)
    if err != nil {
        return nil, err
    }
    if resp.Body != nil {
        defer resp.Body.Close()
    }
    switch resp.StatusCode {
    	case http.StatusOK:
    		break
    	default:
    		return nil, ErrInvalidResponse
    }
    var s *response.ServerResource
    if err := json.NewDecoder(resp.Body).Decode(&s); err != nil {
        return nil, err
    }
    return s, nil
}

// GET /servers/:server_id/config
func (c *Client) GetServerConfigs() error {
    // TODO: need implement
    return nil
}

// GET /servers/:server_id/config/:config_setting_name
func (c *Client) GetServerConfig(uuid string) error {
    // TODO: need implement
    return nil
}

// GET /servers/:server_id/zones
func (c *Client) GetZones() (*[]response.ZoneCollection, error) {
    b := bytes.NewBuffer(nil)
    req, err := c.newRequest("GET", c.joinUrl(fmt.Sprintf("servers/%s/zones", "localhost")), b)
    if err != nil {
        return nil, err
    }
    resp, err := c.h.Do(req)
    if err != nil {
        return nil, err
    }
    if resp.Body != nil {
        defer resp.Body.Close()
    }
    switch resp.StatusCode {
    	case http.StatusOK:
    		break
    	default:
    		return nil, ErrInvalidResponse
    }
    var s *[]response.ZoneCollection
    if err := json.NewDecoder(resp.Body).Decode(&s); err != nil {
        return nil, err
    }
    return s, nil
}

// POST /servers/:server_id/zones
func (c *Client) AddZone(zone string, server interface{}) (*response.ZoneCollection, error) {
    if server == nil {
        server = "localhost"
    }
    jsonZone, err := json.Marshal(response.ZoneCollection{
        Name: zone,
        Kind: "Native",
        Masters: []string{},
        Nameservers: []string{},
    })
    if err != nil {
        return nil, err
    }
    b := bytes.NewBuffer([]byte(jsonZone))
    req, err := c.newRequest("POST", c.joinUrl(fmt.Sprintf("servers/%s/zones", server)), b)
    if err != nil {
        return nil, err
    }
    resp, err := c.h.Do(req)
    if err != nil {
        return nil, err
    }
    if resp.Body != nil {
        defer resp.Body.Close()
    }
    switch resp.StatusCode {
    	case http.StatusOK:
    		break
        case http.StatusCreated:
    		break
    	case http.StatusBadRequest:
    		return nil, ErrInvalidResponse
        case 422:
            var s *response.Error
            if err := json.NewDecoder(resp.Body).Decode(&s); err != nil {
                return nil, errors.New("Resource already exists")
            }
            var errorText string
            if errorText = "Resource already exists"; s.Error != "" {
                errorText = s.Error
            }
    		return nil, errors.New(errorText)
    	default:
    		return nil, ErrInvalidResponse
    }
    var s *response.ZoneCollection
    if err := json.NewDecoder(resp.Body).Decode(&s); err != nil {
        return nil, err
    }
    return s, nil
}

func (c *Client) Add(uuid string, s *msg.Service) error {
	b := bytes.NewBuffer(nil)
	if err := json.NewEncoder(b).Encode(s); err != nil {
		return err
	}
	_, error := c.AddZone(s.Version, nil)
	if error != nil {
        fmt.Fprintf(os.Stderr, "Log: ERROR: %s \n", error.Error())
	}
	return error
}

func (c *Client) Delete(uuid string) error {
	logger("Delete", nil)
	req, err := c.newRequest("DELETE", c.joinUrl(uuid), nil)
	if err != nil {
		return err
	}
	resp, err := c.h.Do(req)
	if err != nil {
		return err
	}
	if resp.Body != nil {
		defer resp.Body.Close()
	}
	return nil
}

func (c *Client) Get(uuid string) (*msg.Service, error) {
	logger("GET", nil)
	req, err := c.newRequest("GET", c.joinUrl(uuid), nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.h.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.Body != nil {
		defer resp.Body.Close()
	}
	switch resp.StatusCode {
	case http.StatusOK:
		break
	case http.StatusNotFound:
		return nil, ErrServiceNotFound
	default:
		return nil, ErrInvalidResponse
	}

	var s *msg.Service
	if err := json.NewDecoder(resp.Body).Decode(&s); err != nil {
		return nil, err
	}
	return s, nil
}

func (c *Client) Update(uuid string, ttl uint32) error {
	b := bytes.NewBuffer([]byte(fmt.Sprintf(`{"TTL":%d}`, ttl)))

	logger("Update", b)
	req, err := c.newRequest("PATCH", c.joinUrl(uuid), b)
	if err != nil {
		return err
	}
	resp, err := c.h.Do(req)
	if err != nil {
		return err
	}
	if resp.Body != nil {
		defer resp.Body.Close()
	}
	return nil
}

func (c *Client) GetAllServices() ([]*msg.Service, error) {
	logger("GetAll", nil)
	req, err := c.newRequest("GET", c.joinUrl(""), nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.h.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.Body != nil {
		defer resp.Body.Close()
	}

	var out []*msg.Service
	if resp.StatusCode == http.StatusOK {
		if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
			return nil, err
		}
	}
	return out, nil
}

func (c *Client) GetAllServicesDNS() ([]*msg.Service, error) {
	req, err := c.newRequestDNS("", dns.TypeSRV)
	if err != nil {
		return nil, err
	}
	resp, _, err := c.d.Exchange(req, c.basedns)
	if err != nil {
		return nil, err
	}
	// Handle UUID.skydns.local additional section stuff? TODO(miek)
	s := make([]*msg.Service, len(resp.Answer))
	for i, r := range resp.Answer {
		if v, ok := r.(*dns.SRV); ok {
			s[i] = &msg.Service{
				// TODO(miek): uehh, stuff it in Name?
				Name: v.Header().Name + " (Priority: " + strconv.Itoa(int(v.Priority)) + ", " + "Weight: " + strconv.Itoa(int(v.Weight)) + ")",
				Host: v.Target,
				Port: v.Port,
				TTL:  r.Header().Ttl,
			}
		}
	}
	return s, nil
}

func (c *Client) GetRegions() (NameCount, error) {
	logger("Get Regions" + fmt.Sprintf("%s/skydns/regions/", c.base), nil)
	req, err := c.newRequest("GET", fmt.Sprintf("%s/skydns/regions/", c.base), nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.h.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.Body != nil {
		defer resp.Body.Close()
	}

	var out NameCount
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) GetEnvironments() (NameCount, error) {
	req, err := c.newRequest("GET", fmt.Sprintf("%s/skydns/environments/", c.base), nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.h.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.Body != nil {
		defer resp.Body.Close()
	}

	var out NameCount
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) AddCallback(uuid string, cb *msg.Callback) error {
	buf := bytes.NewBuffer(nil)
	if err := json.NewEncoder(buf).Encode(cb); err != nil {
		return err
	}
	req, err := c.newRequest("PUT", fmt.Sprintf("%s/skydns/callbacks/%s", c.base, uuid), buf)
	if err != nil {
		return err
	}
	resp, err := c.h.Do(req)
	if err != nil {
		return err
	}
	if resp.Body != nil {
		defer resp.Body.Close()
	}

	switch resp.StatusCode {
	case http.StatusCreated:
		return nil
	case http.StatusNotFound:
		return ErrServiceNotFound
	default:
		return ErrInvalidResponse
	}
}

func (c *Client) joinUrl(action string) string {
	return fmt.Sprintf("%s/%s", c.base, action)
}

func (c *Client) newRequest(method, url string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, url, body)
	req.Header.Add("X-API-Key", "SOMEKEY")
	if c.secret != "" {
		req.Header.Add("Authorization", c.secret)
	}
	return req, err
}

func (c *Client) newRequestDNS(qname string, qtype uint16) (*dns.Msg, error) {
	m := new(dns.Msg)
	if qname == "" {
		m.SetQuestion(c.domain, qtype)
	} else {
		m.SetQuestion(qname+"."+c.domain, qtype)
	}
	return m, nil
}

func (c *Client) extractBaseFromLocation(location string) (string, error) {
	u, err := url.ParseRequestURI(location)
	if err != nil {
		return "", err
	}
	base := u.Scheme + "://" + u.Host
	return base, nil
}