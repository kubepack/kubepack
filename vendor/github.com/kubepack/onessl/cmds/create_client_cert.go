package cmds

import (
	"crypto/x509"
	"fmt"
	"os"

	"github.com/appscode/go/log"
	"github.com/appscode/kutil/tools/certstore"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"k8s.io/client-go/util/cert"
)

func NewCmdCreateClient(certDir string) *cobra.Command {
	var (
		org       string
		overwrite bool
	)
	cmd := &cobra.Command{
		Use:               "client-cert",
		Short:             "Generate client certificate pair",
		DisableAutoGenTag: true,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				log.Fatalln("Missing client name.")
			}
			if len(args) > 1 {
				log.Fatalln("Multiple client name found.")
			}

			cfg := cert.Config{
				CommonName:   args[0],
				Usages:       []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
				Organization: []string{org},
			}

			store, err := certstore.NewCertStore(afero.NewOsFs(), certDir)
			if err != nil {
				fmt.Printf("Failed to create certificate store. Reason: %v.", err)
				os.Exit(1)
			}
			if store.IsExists(Filename(cfg)) && overwrite {
				fmt.Printf("Client certificate found at %s. Do you want to overwrite?", store.Location())
				os.Exit(1)
			}

			if err := store.LoadCA(); err != nil {
				fmt.Printf("Failed to load ca certificate. Reason: %v.", err)
				os.Exit(1)
			}

			crt, key, err := store.NewClientCertPair(cfg.CommonName, cfg.Organization...)
			if err != nil {
				fmt.Printf("Failed to generate client certificate pair. Reason: %v.", err)
				os.Exit(1)
			}
			err = store.WriteBytes(Filename(cfg), crt, key)
			if err != nil {
				fmt.Printf("Failed to init client certificate pair. Reason: %v.", err)
				os.Exit(1)
			}
			fmt.Println("Wrote client certificates in ", store.Location())
			os.Exit(0)
		},
	}

	cmd.Flags().StringVar(&certDir, "cert-dir", certDir, "Path to directory where pki files are stored.")
	cmd.Flags().StringVarP(&org, "organization", "o", org, "Name of client organizations.")
	cmd.Flags().BoolVar(&overwrite, "overwrite", overwrite, "Overwrite existing cert/key pair")
	return cmd
}
