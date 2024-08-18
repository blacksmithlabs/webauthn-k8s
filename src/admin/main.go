package main

import "blacksmithlabs.dev/k8s-webauthn/admin/config"

var (
	sessionTimeout = config.GetSessionTimeout()
)

func main() {
	// Do something with sessionTimeout
}
