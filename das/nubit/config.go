package nubit

import (
	"github.com/spf13/pflag"
)

type NubitConfig struct {
	Enable    bool   `koanf:"enable"`
	Url       string `koanf:"url"`
	Namespace string `koanf:"namespace"`
	Authkey   string `koanf:"authkey"`
}

func NubitConfigAddOptions(prefix string, f *pflag.FlagSet) {
	f.Bool(prefix+".enable", false, "enable to submit blobs to Nubit DA")
	f.String(prefix+".url", "http://localhost:26656", "the address to use the Nuport RPC service")
	f.String(prefix+".namespace", "nitro-dev", "the namespace used to identify the integration")
	f.String(prefix+".authkey", "", "the auth key for the Nuport RPC service")
}

var DefaultNubitConfig = NubitConfig{
	Enable:    false,
	Url:       "http://localhost:26656",
	Namespace: "nitro-dev",
	Authkey:   "",
}
