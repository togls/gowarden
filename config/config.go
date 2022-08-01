package config

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/google/wire"
	"github.com/rs/zerolog"
	"github.com/togls/gowarden/pkg/crypto"
)

var WireSet = wire.NewSet(
	New,

	LoadGlobalDomain,

	wire.Bind(new(IframeAncestorsGetter), new(*Core)),
)

type Core struct {
	Logger *zerolog.Logger

	PriKey *rsa.PrivateKey
	PubKey *rsa.PublicKey

	Folders
	WS
	Jobs
	Settings
	Advanced
}

func New(configFile string, logger *zerolog.Logger) (*Core, error) {
	core := defaultConfig()

	core.Logger = logger

	if configFile != "" {
		if err := core.mergeJson(configFile); err != nil {
			return nil, err
		}
	}

	if err := core.Folders.check(); err != nil {
		logger.Debug().Err(err).Msg("failed to check folders")
		return nil, err
	}

	if err := core.loadRSAKey(); err != nil {
		logger.Debug().Err(err).Msg("failed to load rsa key")
		return nil, err
	}

	return core, nil
}

func (core *Core) loadRSAKey() error {
	// check if pri key file exists
	_, err := os.Stat(core.PrivateKeyPath())
	if errors.Is(err, os.ErrNotExist) {
		rsaKey, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			return fmt.Errorf("failed to generate rsa key: %w", err)
		}

		core.PriKey = rsaKey

		// write key to file
		if err := crypto.WritePrivateKeyToPem(rsaKey, core.PrivateKeyPath()); err != nil {
			return fmt.Errorf("failed to write rsa key to file: %w", err)
		}
	} else {
		privKey, err := crypto.ReadPrivateKeyFromPem(core.PrivateKeyPath())
		if err != nil {
			return fmt.Errorf("failed to read rsa key from file: %w", err)
		}

		core.PriKey = privKey
	}

	// check if pub key file exists
	_, err = os.Stat(core.PublicKeyPath())
	if errors.Is(err, os.ErrNotExist) {
		pubKey := core.PriKey.PublicKey

		// write key to file
		if err := crypto.WritePublicKeyToPem(&pubKey, core.PublicKeyPath()); err != nil {
			return fmt.Errorf("failed to write rsa key to file: %w", err)
		}

		core.PubKey = &pubKey

	} else {
		pubKey, err := crypto.ReadPublicKeyFromPem(core.PublicKeyPath())
		if err != nil {
			return fmt.Errorf("failed to read rsa key from file: %w", err)
		}

		core.PubKey = pubKey
	}

	return nil
}

func (core *Core) mergeJson(configFile string) error {
	f, err := os.Open(configFile)
	if err != nil {
		return fmt.Errorf("failed to open config file: %w", err)
	}
	defer f.Close()

	dec := json.NewDecoder(f)
	if err := dec.Decode(core); err != nil {
		return fmt.Errorf("failed to decode config: %w", err)
	}

	return nil
}

type WS struct {
	Enabled bool
	Addres  string
	Port    int
}

type Jobs struct {
	PollInterval                  int
	SendPurge                     string
	TrashPurge                    string
	Incomplete2fa                 string
	EmergencyNotificationReminder string
	EmergencyRequestTimeout       string
}
