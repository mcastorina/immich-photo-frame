package cache

type ClientDiagnostics struct {
	LocalConfigured      bool
	InMemoryConfigured   bool
	RemoteConfigured     bool
	RemoteConnectedError error
}

func (c Client) Diagnostics() ClientDiagnostics {
	diagnostics := ClientDiagnostics{}
	diagnostics.LocalConfigured = (c.local != nil)
	diagnostics.InMemoryConfigured = (c.cache != nil)
	diagnostics.RemoteConfigured = (c.remote != nil)
	diagnostics.RemoteConnectedError = c.remote.IsConnected()

	return diagnostics
}
