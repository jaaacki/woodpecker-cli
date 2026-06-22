package commands

import (
	"github.com/jaaacki/woodpecker-cli/internal/api"
	"github.com/jaaacki/woodpecker-cli/internal/client"
)

const versionUnavailableNote = "Woodpecker 3.x exposes no /version endpoint"

type versionProbe struct {
	Available bool        `json:"available"`
	Value     api.Version `json:"value,omitempty"`
	Note      string      `json:"note,omitempty"`
}

func probeVersion(c *client.Client) versionProbe {
	var version api.Version
	if err := c.GetJSON(c.URL("version"), &version); err != nil {
		return versionProbe{
			Available: false,
			Note:      versionUnavailableNote,
		}
	}
	return versionProbe{
		Available: true,
		Value:     version,
	}
}
