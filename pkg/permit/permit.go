package permit

import (
	"regexp"
	"time"

	"github.com/pkg/errors"
)

type (
	Permit struct {
		Version    uint           `json:"version"`
		Key        string         `json:"key"`
		Domain     string         `json:"domain"`
		Expires    *time.Time     `json:"expires,omitempty"`
		Valid      bool           `json:"valid"`
		Attributes map[string]int `json:"attributes"`
		Contact    string         `json:"contact"`
		Entity     string         `json:"entity"`
		Issued     time.Time      `json:"issued"`
	}
)

const (
	// KeyLength
	KeyLength = 64
)

var (
	domainCheck       = regexp.MustCompile(`^([a-zA-Z0-9-_]+\.)*[a-zA-Z0-9][a-zA-Z0-9-_]+\.[a-zA-Z]{2,11}?$`)
	PermitNotFound    = errors.New("permit not found")
	DefaultAttributes = map[string]int{
		"system.enabled":                 1,
		"system.max-users":               -1,
		"system.max-organisations":       1,
		"system.max-teams":               -1,
		"messaging.enabled":              1,
		"messaging.max-users":            -1,
		"messaging.max-private-channels": -1,
		"messaging.max-public-channels":  -1,
		"messaging.max-messages":         -1,
		"compose.enabled":                1,
		"compose.max-namespaces":         -1,
		"compose.max-users":              -1,
		"compose.max-modules":            -1,
		"compose.max-charts":             -1,
		"compose.max-pages":              -1,
		"compose.max-triggers":           -1,
	}
)

func ValidateDomain(d string) bool {
	return domainCheck.MatchString(d)
}

func (p Permit) IsValid() bool {
	return p.Valid && !p.Expired() && ValidateDomain(p.Domain)
}

func (p Permit) Expired() bool {
	return p.Expires != nil && p.Expires.Before(time.Now())
}
