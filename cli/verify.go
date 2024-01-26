package main

import (
	"crypto/sha256"
	"fmt"
	"net"
	"os"
	"path/filepath"

	"github.com/edgelesssys/nunki/internal/atls"
	"github.com/edgelesssys/nunki/internal/attestation/snp"
	"github.com/edgelesssys/nunki/internal/coordapi"
	"github.com/edgelesssys/nunki/internal/fsstore"
	"github.com/edgelesssys/nunki/internal/grpc/dialer"
	"github.com/edgelesssys/nunki/internal/manifest"
	"github.com/google/go-sev-guest/abi"
	"github.com/google/go-sev-guest/kds"
	"github.com/google/go-sev-guest/validate"
	"github.com/spf13/cobra"
)

func newVerifyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "verify",
		Short: "Verify a nunki deployment",
		Long: `
		Verify a nunki deployment.

		This will connect to the given Coordinator using aTLS. During the connection
		initialization, the remote attestation of the Coordinator CVM happens and
		the connection will only be successful if the Coordinator conforms with the
		reference values embedded into the CLI.

		After the connection is established, the CLI will request the manifest histroy,
		all policies, and the certificates of the Coordinator certifcate authority.
	`,
		RunE: runVerify,
	}

	cmd.Flags().StringP("output", "o", verifyDir, "directory to write files to")
	cmd.Flags().StringP("coordinator", "c", "", "endpoint the coordinator can be reached at")
	must(cobra.MarkFlagRequired(cmd.Flags(), "coordinator"))

	return cmd
}

func runVerify(cmd *cobra.Command, _ []string) error {
	flags, err := parseVerifyFlags(cmd)
	if err != nil {
		return fmt.Errorf("parsing flags: %w", err)
	}

	log, err := newCLILogger(cmd)
	if err != nil {
		return err
	}
	log.Debug("Starting verification")

	kdsDir, err := cachedir("kds")
	if err != nil {
		return fmt.Errorf("getting cache dir: %w", err)
	}
	log.Debug("Using KDS cache dir", "dir", kdsDir)

	validateOptsGen := newCoordinatorValidateOptsGen()
	kdsCache := fsstore.New(kdsDir, log.WithGroup("kds-cache"))
	kdsGetter := snp.NewCachedHTTPSGetter(kdsCache, snp.NeverGCTicker, log.WithGroup("kds-getter"))
	validator := snp.NewValidator(validateOptsGen, kdsGetter, log.WithGroup("snp-validator"))
	dialer := dialer.New(atls.NoIssuer, validator, &net.Dialer{})

	log.Debug("Dialing coordinator", "endpoint", flags.coordinator)
	conn, err := dialer.Dial(cmd.Context(), flags.coordinator)
	if err != nil {
		return fmt.Errorf("Error: failed to dial coordinator: %w", err)
	}
	defer conn.Close()

	log.Debug("Getting manifest")
	client := coordapi.NewCoordAPIClient(conn)
	resp, err := client.GetManifests(cmd.Context(), &coordapi.GetManifestsRequest{})
	if err != nil {
		return fmt.Errorf("failed to get manifest: %w", err)
	}
	log.Debug("Got response")

	filelist := map[string][]byte{
		coordRootPEMFilename:   resp.CACert,
		coordIntermPEMFilename: resp.IntermCert,
	}
	for i, m := range resp.Manifests {
		filelist[fmt.Sprintf("manifest.%d.json", i)] = m
	}
	for _, p := range resp.Policies {
		sha256sum := sha256.Sum256(p)
		pHash := manifest.NewHexString(sha256sum[:])
		filelist[fmt.Sprintf("policy.%s.rego", pHash)] = p
	}
	if err := writeFilelist(flags.outputDir, filelist); err != nil {
		return fmt.Errorf("writing filelist: %w", err)
	}

	fmt.Fprintln(cmd.OutOrStdout(), "Successfully verified coordinator")

	return nil
}

type verifyFlags struct {
	coordinator string
	outputDir   string
}

func parseVerifyFlags(cmd *cobra.Command) (*verifyFlags, error) {
	coordinator, err := cmd.Flags().GetString("coordinator")
	if err != nil {
		return nil, err
	}
	outputDir, err := cmd.Flags().GetString("output")
	if err != nil {
		return nil, err
	}

	return &verifyFlags{
		coordinator: coordinator,
		outputDir:   outputDir,
	}, nil
}

func newCoordinatorValidateOptsGen() *snp.StaticValidateOptsGenerator {
	defaultManifest := manifest.Default()
	trustedIDKeyDigests, err := (&defaultManifest.ReferenceValues.SNP.TrustedIDKeyHashes).ByteSlices()
	if err != nil {
		panic(err) // We are decoding known values, tests should catch any failure.
	}

	return &snp.StaticValidateOptsGenerator{
		Opts: &validate.Options{
			GuestPolicy: abi.SnpPolicy{
				Debug: false,
				SMT:   true,
			},
			VMPL: new(int), // VMPL0
			MinimumTCB: kds.TCBParts{
				BlSpl:    3,
				TeeSpl:   0,
				SnpSpl:   8,
				UcodeSpl: 115,
			},
			MinimumLaunchTCB: kds.TCBParts{
				BlSpl:    3,
				TeeSpl:   0,
				SnpSpl:   8,
				UcodeSpl: 115,
			},
			PermitProvisionalFirmware: true,
			TrustedIDKeyHashes:        trustedIDKeyDigests,
			RequireIDBlock:            true,
		},
	}
}

func writeFilelist(dir string, filelist map[string][]byte) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating directory %s: %w", dir, err)
	}
	for filename, contents := range filelist {
		path := filepath.Join(dir, filename)
		if err := os.WriteFile(path, contents, 0o644); err != nil {
			return fmt.Errorf("writing %q: %w", path, err)
		}
	}
	return nil
}
