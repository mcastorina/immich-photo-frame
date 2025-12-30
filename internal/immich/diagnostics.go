package immich

// ClientDiagnostics holds the information from the call to [Diagnostics].
type ClientDiagnostics struct {
	LocalConfigured      bool
	InMemoryConfigured   bool
	RemoteConfigured     bool
	RemoteConnectedError error
}

// Diagnostics reports how the client is configured and checks if the remote is connected.
//
// TODO: Add in-memory and local information like memory used / configured.
func (c Client) Diagnostics() ClientDiagnostics {
	diagnostics := ClientDiagnostics{}
	diagnostics.LocalConfigured = (c.local != nil)
	diagnostics.InMemoryConfigured = (c.cache != nil)
	diagnostics.RemoteConfigured = (c.remote != nil)
	diagnostics.RemoteConnectedError = c.remote.IsConnected()

	return diagnostics
}
