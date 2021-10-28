package hostnamedata

// HostnameData keeps hostname and additional identifiers (such as cloud identifiers) together
type HostnameData struct {
	Hostname    string
	Identifiers []string
}

// JustHostname builds HostnameData out of hostname only
func JustHostname(hostname string, err error) (*HostnameData, error) {
	if hostname == "" {
		return nil, err
	}
	return &HostnameData{hostname, []string{}}, err
}
