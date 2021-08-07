package purevpn

import (
	"math/rand"

	"github.com/qdm12/gluetun/internal/constants"
	"github.com/qdm12/gluetun/internal/models"
	"github.com/qdm12/gluetun/internal/provider/utils"
)

type Purevpn struct {
	servers    []models.PurevpnServer
	randSource rand.Source
	utils.NoPortForwarder
}

func New(servers []models.PurevpnServer, randSource rand.Source) *Purevpn {
	return &Purevpn{
		servers:         servers,
		randSource:      randSource,
		NoPortForwarder: utils.NewNoPortForwarding(constants.Purevpn),
	}
}
