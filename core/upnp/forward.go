package upnp

import (
	"fmt"
	"github.com/jcuga/go-upnp"
)

type UPnP struct {
	client upnp.IGD
}

func NewUPnP() (*UPnP, error) {
	d, err := upnp.Discover()
	if err != nil {
		return nil, fmt.Errorf("failed to do upnp discover: %w", err)
	}
	return &UPnP{client: d}, nil
}

func (u *UPnP) ForwardPortUDP(port uint16) error {
	err := u.client.Forward(port, "TON Torrent (Storage)", "UDP")
	if err != nil {
		return fmt.Errorf("failed to forward port using upnp: %w", err)
	}
	return nil
}

func (u *UPnP) ForwardPortTCP(port uint16) error {
	err := u.client.Forward(port, "TON Torrent (Storage)", "TCP")
	if err != nil {
		return fmt.Errorf("failed to forward port using upnp: %w", err)
	}
	return nil
}

func (u *UPnP) ExternalIP() (string, error) {
	ip, err := u.client.ExternalIP()
	if err != nil {
		return "", fmt.Errorf("failed to do external ip discover: %w", err)
	}
	return ip, nil
}
