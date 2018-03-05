package cmds

import (
	"crypto/x509"
	"fmt"
	"net"
	"os"

	"github.com/appscode/go/log"
	"github.com/appscode/kutil/tools/certstore"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"k8s.io/client-go/util/cert"
)

func NewCmdCreateServer(certDir string) *cobra.Command {
	var (
		sans = cert.AltNames{
			IPs: []net.IP{net.ParseIP("127.0.0.1")},
		}
		overwrite bool
	)
	cmd := &cobra.Command{
		Use:               "server-cert",
		Short:             "Generate server certificate pair",
		DisableAutoGenTag: true,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) > 1 {
				log.Fatalln("Multiple server name found.")
			}
			if len(args) == 0 {
				args = []string{"server"}
			}

			cfg := cert.Config{
				CommonName: args[0],
				AltNames:   sans,
				Usages:     []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
			}

			store, err := certstore.NewCertStore(afero.NewOsFs(), certDir)
			if err != nil {
				fmt.Printf("Failed to create certificate store. Reason: %v.", err)
				os.Exit(1)
			}
			if store.IsExists(Filename(cfg)) && !overwrite {
				fmt.Printf("Server certificate found at %s. Do you want to overwrite?", store.Location())
				os.Exit(1)
			}

			if err = store.LoadCA(); err != nil {
				fmt.Printf("CA certificates not found in %s. Run `init ca`", store.Location())
				os.Exit(1)
			}

			crt, key, err := store.NewServerCertPair(cfg.CommonName, cfg.AltNames)
			if err != nil {
				fmt.Printf("Failed to generate server certificate pair. Reason: %v.", err)
				os.Exit(1)
			}
			err = store.WriteBytes(Filename(cfg), crt, key)
			if err != nil {
				fmt.Printf("Failed to init server certificate pair. Reason: %v.", err)
				os.Exit(1)
			}
			fmt.Println("Wrote server certificates in ", store.Location())
			os.Exit(0)
		},
	}

	cmd.Flags().StringVar(&certDir, "cert-dir", certDir, "Path to directory where pki files are stored.")
	cmd.Flags().IPSliceVar(&sans.IPs, "ips", sans.IPs, "Alternative IP addresses")
	cmd.Flags().StringSliceVar(&sans.DNSNames, "domains", sans.DNSNames, "Alternative Domain names")
	cmd.Flags().BoolVar(&overwrite, "overwrite", overwrite, "Overwrite existing cert/key pair")
	return cmd
}
