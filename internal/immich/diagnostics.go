package immich

// ClientDiagnostics holds the information from the call to [Diagnostics].
type ClientDiagnostics struct {
	LocalConfigured      bool
	InMemoryConfigured   bool
	RemoteConfigured     bool
	RemoteConnectedError string
}

// Diagnostics reports how the client is configured and checks if the remote is connected.
//
// TODO: Add in-memory and local information like memory used / configured.
func (c Client) Diagnostics() ClientDiagnostics {
	remoteConnected := ""
	if err := c.remote.IsConnected(); err != nil {
		remoteConnected = err.Error()
	}
	return ClientDiagnostics{
		LocalConfigured:      (c.local != nil),
		InMemoryConfigured:   (c.cache != nil),
		RemoteConfigured:     (c.remote != nil),
		RemoteConnectedError: remoteConnected,
	}
}
