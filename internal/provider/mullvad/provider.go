package mullvad

import (
	"math/rand"

	"github.com/qdm12/gluetun/internal/constants"
	"github.com/qdm12/gluetun/internal/models"
	"github.com/qdm12/gluetun/internal/provider/utils"
)

type Mullvad struct {
	servers    []models.MullvadServer
	randSource rand.Source
	utils.NoPortForwarder
}

func New(servers []models.MullvadServer, randSource rand.Source) *Mullvad {
	return &Mullvad{
		servers:         servers,
		randSource:      randSource,
		NoPortForwarder: utils.NewNoPortForwarding(constants.Mullvad),
	}
}
