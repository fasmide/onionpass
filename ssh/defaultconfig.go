package ssh

import (
	"fmt"
	"io/ioutil"

	"golang.org/x/crypto/ssh"
)

// DefaultConfig generates a default ssh.ServerConfig
func DefaultConfig() (*ssh.ServerConfig, error) {
	config := &ssh.ServerConfig{
		// Remove to disable password auth.
		PasswordCallback: func(c ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
			return nil, nil
		},

		// Remove to disable public key auth.
		PublicKeyCallback: func(c ssh.ConnMetadata, pubKey ssh.PublicKey) (*ssh.Permissions, error) {
			return &ssh.Permissions{
				// Record the public key used for authentication.
				Extensions: map[string]string{
					"pubkey-fp": ssh.FingerprintSHA256(pubKey),
				},
			}, nil
		},
	}

	privateBytes, err := ioutil.ReadFile("id_rsa")
	if err != nil {
		return nil, fmt.Errorf("Failed to load private key: %s", err)
	}

	private, err := ssh.ParsePrivateKey(privateBytes)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse private key: %s", err)
	}

	config.AddHostKey(private)

	return config, nil
}
