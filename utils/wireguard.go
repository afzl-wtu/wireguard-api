package utils

import (
	"time"

	log "github.com/sirupsen/logrus"
	"golang.zx2c4.com/wireguard/wgctrl"
)

func GetInActiveClients() []string {
	wgClient, err := wgctrl.New()
	if err != nil {
		log.Error("Failed to create wgctrl client: ", err)
		return nil
	}

	devices, err := wgClient.Devices()
	if err != nil {
		log.Error("Failed to get devices: ", err)
		return nil
	}
	inActiveDevices := make([]string, 0)
	for _, device := range devices {
		log.Info("Device: ", device.Name)
		for _, peer := range device.Peers {
			if time.Since(peer.LastHandshakeTime).Seconds() > 100 {
				inActiveDevices = append(inActiveDevices, peer.PublicKey.String())
			}
		}
	}
	return inActiveDevices
}
