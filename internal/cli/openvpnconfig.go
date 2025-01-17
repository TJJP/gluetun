package cli

import (
	"fmt"
	"strings"
	"time"

	"github.com/qdm12/gluetun/internal/configuration"
	"github.com/qdm12/gluetun/internal/constants"
	"github.com/qdm12/gluetun/internal/provider"
	"github.com/qdm12/gluetun/internal/storage"
	"github.com/qdm12/golibs/logging"
	"github.com/qdm12/golibs/params"
)

type OpenvpnConfigMaker interface {
	OpenvpnConfig(logger logging.Logger) error
}

func (c *CLI) OpenvpnConfig(logger logging.Logger) error {
	var allSettings configuration.Settings
	err := allSettings.Read(params.NewEnv(), logger)
	if err != nil {
		return err
	}
	allServers, err := storage.New(logger, constants.ServersData).
		SyncServers(constants.GetAllServers())
	if err != nil {
		return err
	}
	providerConf := provider.New(allSettings.OpenVPN.Provider.Name, allServers, time.Now)
	connection, err := providerConf.GetOpenVPNConnection(allSettings.OpenVPN.Provider.ServerSelection)
	if err != nil {
		return err
	}
	lines := providerConf.BuildConf(connection, "nonroortuser", allSettings.OpenVPN)
	fmt.Println(strings.Join(lines, "\n"))
	return nil
}
