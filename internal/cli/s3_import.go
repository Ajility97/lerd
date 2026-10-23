package cli

import (
	"fmt"
	"slices"
	"strings"

	"github.com/geodro/lerd/internal/feedback"
	"github.com/geodro/lerd/internal/serviceops"
	"github.com/spf13/cobra"
)

// The remote side of an import, vars so tests stand in for the server.
var (
	s3RemoteBuckets     = serviceops.S3RemoteBuckets
	s3RemoteObjectCount = serviceops.S3RemoteObjectCount
	s3MirrorInto        = serviceops.S3MirrorInto
)

// NewS3ImportCmd copies a bucket from any S3 server, such as another local
// environment's MinIO, into a bucket on lerd's own S3 service.
func NewS3ImportCmd() *cobra.Command {
	var endpoint, access, secret, from, to string
	cmd := &cobra.Command{
		Use:   "s3:import",
		Short: "Copy a bucket from another S3 server into lerd's S3 service",
		Long: `Copies every object of a bucket on another S3 server into a bucket on lerd's
own S3 service, creating it when missing. Objects with the same key are
overwritten and nothing is deleted.

  lerd s3:import --endpoint 127.0.0.1:9000 --access-key key --secret-key secret --bucket uploads --to shop`,
		Args: cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			host := strings.TrimPrefix(strings.TrimPrefix(endpoint, "http://"), "https://")
			return runS3Import(serviceops.S3Remote{Endpoint: strings.TrimSuffix(host, "/"), Access: access, Secret: secret}, from, to)
		},
	}
	cmd.Flags().StringVar(&endpoint, "endpoint", "", "The server to copy from, host:port")
	cmd.Flags().StringVar(&access, "access-key", "", "Its access key")
	cmd.Flags().StringVar(&secret, "secret-key", "", "Its secret key")
	cmd.Flags().StringVar(&from, "bucket", "", "The bucket to copy; the server's only bucket when omitted")
	cmd.Flags().StringVar(&to, "to", "", "The bucket on lerd's S3 service to copy into")
	_ = cmd.MarkFlagRequired("endpoint")
	return cmd
}

func runS3Import(remote serviceops.S3Remote, from, to string) error {
	if to == "" {
		return fmt.Errorf("name the bucket to copy into with --to")
	}
	buckets, err := s3RemoteBuckets(remote)
	if err != nil {
		return fmt.Errorf("listing the buckets on %s: %w", remote.Endpoint, err)
	}
	source, err := pickSourceBucket(buckets, from)
	if err != nil {
		return err
	}
	count, err := s3RemoteObjectCount(remote, source)
	if err != nil {
		return fmt.Errorf("listing bucket %q: %w", source, err)
	}
	feedback.Begin()
	if count == 0 {
		feedback.Line("bucket " + feedback.Val(source) + " is empty, nothing to copy")
		return nil
	}
	if err := ensureServiceRunning("rustfs"); err != nil {
		return fmt.Errorf("starting rustfs: %w", err)
	}
	dest := s3BucketName(to)
	if _, err := createS3Bucket(dest); err != nil {
		return fmt.Errorf("creating bucket %q: %w", dest, err)
	}
	step := feedback.Start(fmt.Sprintf("copying %d objects from %s into %s", count, source, dest))
	if err := s3MirrorInto(remote, source, dest, nil); err != nil {
		step.Fail(err)
		return err
	}
	step.OK("")
	return nil
}

// pickSourceBucket takes the bucket the caller named, or the server's only one;
// with several and none named it lists them rather than guess.
func pickSourceBucket(buckets []string, want string) (string, error) {
	if len(buckets) == 0 {
		return "", fmt.Errorf("the server holds no buckets")
	}
	if want != "" {
		if !slices.Contains(buckets, want) {
			return "", fmt.Errorf("no bucket %q on the server, it holds %s", want, strings.Join(buckets, ", "))
		}
		return want, nil
	}
	if len(buckets) == 1 {
		return buckets[0], nil
	}
	return "", fmt.Errorf("the server holds %s; pick one with --bucket", strings.Join(buckets, ", "))
}
