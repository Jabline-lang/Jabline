package stdlib

// CloseAllResources closes all globally tracked connections.
func CloseAllResources() {
	closeAllRedis()
	closeAllSSH()
}
