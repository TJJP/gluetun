package cyberghost

import (
	"strconv"
	"strings"

	"github.com/qdm12/gluetun/internal/configuration"
	"github.com/qdm12/gluetun/internal/constants"
	"github.com/qdm12/gluetun/internal/models"
	"github.com/qdm12/gluetun/internal/provider/utils"
)

func (c *Cyberghost) BuildConf(connection models.OpenVPNConnection,
	username string, settings configuration.OpenVPN) (lines []string) {
	if settings.Cipher == "" {
		settings.Cipher = constants.AES256cbc
	}

	if settings.Auth == "" {
		settings.Auth = constants.SHA256
	}

	lines = []string{
		"client",
		"dev tun",
		"nobind",
		"persist-key",
		"persist-tun",
		"remote-cert-tls server",
		"ping 10",
		"ping-exit 60",
		"ping-timer-rem",
		"tls-exit",

		// Cyberghost specific
		// "redirect-gateway def1",
		"ncp-disable",
		"script-security 2",
		"route-delay 5",

		// Added constant values
		"auth-nocache",
		"mute-replay-warnings",
		"pull-filter ignore \"auth-token\"", // prevent auth failed loops
		"auth-retry nointeract",
		"suppress-timestamps",

		// Modified variables
		"verb " + strconv.Itoa(settings.Verbosity),
		"auth-user-pass " + constants.OpenVPNAuthConf,
		connection.ProtoLine(),
		connection.RemoteLine(),
		"auth " + settings.Auth,
	}

	lines = append(lines, utils.CipherLines(settings.Cipher, settings.Version)...)

	if connection.Protocol == constants.UDP {
		lines = append(lines, "explicit-exit-notify")
	}

	if strings.HasSuffix(settings.Cipher, "-gcm") {
		lines = append(lines, "ncp-ciphers AES-256-GCM:AES-256-CBC:AES-128-GCM")
	}

	if !settings.Root {
		lines = append(lines, "user "+username)
	}

	if settings.MSSFix > 0 {
		lines = append(lines, "mssfix "+strconv.Itoa(int(settings.MSSFix)))
	}

	if settings.Provider.ExtraConfigOptions.OpenVPNIPv6 {
		lines = append(lines, "tun-ipv6")
	} else {
		lines = append(lines, `pull-filter ignore "route-ipv6"`)
		lines = append(lines, `pull-filter ignore "ifconfig-ipv6"`)
	}

	lines = append(lines, utils.WrapOpenvpnCA(
		constants.CyberghostCertificate)...)
	lines = append(lines, utils.WrapOpenvpnCert(
		settings.Provider.ExtraConfigOptions.ClientCertificate)...)
	lines = append(lines, utils.WrapOpenvpnKey(
		settings.Provider.ExtraConfigOptions.ClientKey)...)

	lines = append(lines, "")

	return lines
}
