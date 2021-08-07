package configuration

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/qdm12/gluetun/internal/constants"
	"github.com/qdm12/golibs/params"
)

// OpenVPN contains settings to configure the OpenVPN client.
type OpenVPN struct {
	User      string   `json:"user"`
	Password  string   `json:"password"`
	Verbosity int      `json:"verbosity"`
	Flags     []string `json:"flags"`
	MSSFix    uint16   `json:"mssfix"`
	Root      bool     `json:"run_as_root"`
	Cipher    string   `json:"cipher"`
	Auth      string   `json:"auth"`
	Provider  Provider `json:"provider"`
	Config    string   `json:"custom_config"`
	Version   string   `json:"version"`
}

func (settings *OpenVPN) String() string {
	return strings.Join(settings.lines(), "\n")
}

func (settings *OpenVPN) lines() (lines []string) {
	lines = append(lines, lastIndent+"OpenVPN:")

	lines = append(lines, indent+lastIndent+"Version: "+settings.Version)

	lines = append(lines, indent+lastIndent+"Verbosity level: "+strconv.Itoa(settings.Verbosity))

	if len(settings.Flags) > 0 {
		lines = append(lines, indent+lastIndent+"Flags: "+strings.Join(settings.Flags, " "))
	}

	if settings.Root {
		lines = append(lines, indent+lastIndent+"Run as root: enabled")
	}

	if len(settings.Cipher) > 0 {
		lines = append(lines, indent+lastIndent+"Custom cipher: "+settings.Cipher)
	}
	if len(settings.Auth) > 0 {
		lines = append(lines, indent+lastIndent+"Custom auth algorithm: "+settings.Auth)
	}

	if len(settings.Config) > 0 {
		lines = append(lines, indent+lastIndent+"Custom configuration: "+settings.Config)
	}

	if settings.Provider.Name == "" {
		lines = append(lines, indent+lastIndent+"Provider: custom configuration")
	} else {
		lines = append(lines, indent+lastIndent+"Provider:")
		for _, line := range settings.Provider.lines() {
			lines = append(lines, indent+indent+line)
		}
	}

	return lines
}

var (
	ErrInvalidVPNProvider = errors.New("invalid VPN provider")
)

func (settings *OpenVPN) read(r reader) (err error) {
	vpnsp, err := r.env.Inside("VPNSP", []string{
		"cyberghost", "fastestvpn", "hidemyass", "ipvanish", "ivpn", "mullvad", "nordvpn",
		"privado", "pia", "private internet access", "privatevpn", "protonvpn",
		"purevpn", "surfshark", "torguard", constants.VPNUnlimited, "vyprvpn", "windscribe"},
		params.Default("private internet access"))
	if err != nil {
		return fmt.Errorf("environment variable VPNSP: %w", err)
	}
	if vpnsp == "pia" { // retro compatibility
		vpnsp = "private internet access"
	}

	settings.Provider.Name = vpnsp

	settings.Config, err = r.env.Get("OPENVPN_CUSTOM_CONFIG", params.CaseSensitiveValue())
	if err != nil {
		return fmt.Errorf("environment variable OPENVPN_CUSTOM_CONFIG: %w", err)
	}
	customConfig := settings.Config != ""

	if customConfig {
		settings.Provider.Name = ""
	}

	credentialsRequired := !customConfig && settings.Provider.Name != constants.VPNUnlimited

	settings.User, err = r.getFromEnvOrSecretFile("OPENVPN_USER", credentialsRequired, []string{"USER"})
	if err != nil {
		return fmt.Errorf("environment variable OPENVPN_USER: %w", err)
	}
	// Remove spaces in user ID to simplify user's life, thanks @JeordyR
	settings.User = strings.ReplaceAll(settings.User, " ", "")

	if settings.Provider.Name == constants.Mullvad {
		settings.Password = "m"
	} else {
		settings.Password, err = r.getFromEnvOrSecretFile("OPENVPN_PASSWORD", credentialsRequired, []string{"PASSWORD"})
		if err != nil {
			return err
		}
	}

	settings.Version, err = r.env.Inside("OPENVPN_VERSION",
		[]string{constants.Openvpn24, constants.Openvpn25}, params.Default(constants.Openvpn25))
	if err != nil {
		return fmt.Errorf("environment variable OPENVPN_VERSION: %w", err)
	}

	settings.Verbosity, err = r.env.IntRange("OPENVPN_VERBOSITY", 0, 6, params.Default("1")) //nolint:gomnd
	if err != nil {
		return fmt.Errorf("environment variable OPENVPN_VERBOSITY: %w", err)
	}

	settings.Flags = []string{}
	flagsStr, err := r.env.Get("OPENVPN_FLAGS")
	if err != nil {
		return fmt.Errorf("environment variable OPENVPN_FLAGS: %w", err)
	}
	if flagsStr != "" {
		settings.Flags = strings.Fields(flagsStr)
	}

	settings.Root, err = r.env.YesNo("OPENVPN_ROOT", params.Default("yes"))
	if err != nil {
		return fmt.Errorf("environment variable OPENVPN_ROOT: %w", err)
	}

	settings.Cipher, err = r.env.Get("OPENVPN_CIPHER")
	if err != nil {
		return fmt.Errorf("environment variable OPENVPN_CIPHER: %w", err)
	}

	settings.Auth, err = r.env.Get("OPENVPN_AUTH")
	if err != nil {
		return fmt.Errorf("environment variable OPENVPN_AUTH: %w", err)
	}

	const maxMSSFix = 10000
	mssFix, err := r.env.IntRange("OPENVPN_MSSFIX", 0, maxMSSFix, params.Default("0"))
	if err != nil {
		return fmt.Errorf("environment variable OPENVPN_MSSFIX: %w", err)
	}
	settings.MSSFix = uint16(mssFix)
	return settings.readProvider(r)
}

func (settings *OpenVPN) readProvider(r reader) error {
	var readProvider func(r reader) error
	switch settings.Provider.Name {
	case "": // custom config
		readProvider = func(r reader) error { return nil }
	case constants.Cyberghost:
		readProvider = settings.Provider.readCyberghost
	case constants.Fastestvpn:
		readProvider = settings.Provider.readFastestvpn
	case constants.HideMyAss:
		readProvider = settings.Provider.readHideMyAss
	case constants.Ipvanish:
		readProvider = settings.Provider.readIpvanish
	case constants.Ivpn:
		readProvider = settings.Provider.readIvpn
	case constants.Mullvad:
		readProvider = settings.Provider.readMullvad
	case constants.Nordvpn:
		readProvider = settings.Provider.readNordvpn
	case constants.Privado:
		readProvider = settings.Provider.readPrivado
	case constants.PrivateInternetAccess:
		readProvider = settings.Provider.readPrivateInternetAccess
	case constants.Privatevpn:
		readProvider = settings.Provider.readPrivatevpn
	case constants.Protonvpn:
		readProvider = settings.Provider.readProtonvpn
	case constants.Purevpn:
		readProvider = settings.Provider.readPurevpn
	case constants.Surfshark:
		readProvider = settings.Provider.readSurfshark
	case constants.Torguard:
		readProvider = settings.Provider.readTorguard
	case constants.VPNUnlimited:
		readProvider = settings.Provider.readVPNUnlimited
	case constants.Vyprvpn:
		readProvider = settings.Provider.readVyprvpn
	case constants.Windscribe:
		readProvider = settings.Provider.readWindscribe
	default:
		return fmt.Errorf("%w: %s", ErrInvalidVPNProvider, settings.Provider.Name)
	}
	return readProvider(r)
}
