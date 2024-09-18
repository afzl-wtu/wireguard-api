package interfaces

import model "github.com/afzl-wtu/wireguard-api/models"

type IStore interface {
	Init() error
	GetGlobalSettings() (model.GlobalSetting, error)
	GetServer() (model.Server, error)
	GetClients(hasQRCode bool) ([]model.ClientData, error)
	GetClientByID(clientID string, qrCode model.QRCodeSettings) (model.ClientData, error)
	SaveClient(client model.Client) error
	DeleteClient(clientID string) error
	SaveServerInterface(serverInterface model.ServerInterface) error
	SaveServerKeyPair(serverKeyPair model.ServerKeypair) error
	SaveGlobalSettings(globalSettings model.GlobalSetting) error
	GetPath() string
	SaveHashes(hashes model.ClientServerHashes) error
	GetHashes() (model.ClientServerHashes, error)
}
