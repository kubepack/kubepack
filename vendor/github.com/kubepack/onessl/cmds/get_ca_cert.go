package cmds

import (
	"bufio"
	"crypto/rsa"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"k8s.io/client-go/util/cert"
)

func NewCmdGetCACert() *cobra.Command {
	var cn string
	cmd := &cobra.Command{
		Use:               "ca-cert",
		Short:             "Prints self-sgned CA certificate from PEM encoded RSA private key",
		DisableAutoGenTag: true,
		Run: func(cmd *cobra.Command, args []string) {
			reader := bufio.NewReader(os.Stdin)
			keyBytes, err := ioutil.ReadAll(reader)
			if err != nil {
				Fatal(errors.Wrap(err, "failed to read private key"))
			}
			key, err := cert.ParsePrivateKeyPEM(keyBytes)
			if err != nil {
				Fatal(errors.Wrap(err, "failed to parse private key"))
			}
			rsaKey, ok := key.(*rsa.PrivateKey)
			if !ok {
				Fatal(errors.Wrapf(err, "only supports rsa private key. Found %v", reflect.ValueOf(key).Kind()))
			}
			crt, err := cert.NewSelfSignedCACert(cert.Config{CommonName: cn}, rsaKey)
			if err != nil {
				Fatal(errors.Wrap(err, "failed to generate self-signed certificate"))
			}
			fmt.Println(string(cert.EncodeCertPEM(crt)))
			os.Exit(0)
		},
	}
	cmd.Flags().StringVar(&cn, "common-name", cn, "Common Name used in CA certificate.")
	return cmd
}
