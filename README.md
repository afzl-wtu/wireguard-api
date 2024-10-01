### Wireguard api developed in GO lang for effecient performnce and multithreading. This api returns wiregurd config in string on every GET request.
### That returned config can be directly consumed by  wireguard plugins in client devices.
### As soon as you get config connect to it. becuase if not connected after one minute it will auto expire.
### So if failed to connect or incase reconnecting again get a new config by making GET request to server

# Installation

Run 
```

wget https://git.io/wireguard -O wireguard-install.sh && sudo bash wireguard-install.sh

```

Create first client and finish setup. Then run

```

sudo systemctl stop wg-quick@wg0

```

Then run wireguard api binary with sudo

```

sudo ./api

```

On first start program will create all clients and setup server and then run on port 2200.

