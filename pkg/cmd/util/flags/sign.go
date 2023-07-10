package flags

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// SignOptions contains options for the 'sign' command. These will be available to the subcommands and not configured directly for the sign command itself.
type SignOptions struct {
	OpenStackCoreFlags

	VaultURL        string
	VaultToken      string
	VaultMountPath  string
	VaultSecretPath string
	ImageID         string
}

// SetOptionsFromViper configures additional options passed in via viper for the struct from any subcommands.
func (o *SignOptions) SetOptionsFromViper() {
	o.OpenStackCoreFlags.SetOptionsFromViper()

	o.ImageID = viper.GetString(fmt.Sprintf("%s.image-id", viperSignPrefix))
	o.VaultURL = viper.GetString(fmt.Sprintf("%s.url", viperVaultPrefix))
	o.VaultToken = viper.GetString(fmt.Sprintf("%s.token", viperVaultPrefix))
	o.VaultMountPath = viper.GetString(fmt.Sprintf("%s.mount-path", viperVaultPrefix))
	o.VaultSecretPath = viper.GetString(fmt.Sprintf("%s.secret-name", viperVaultPrefix))
}

// AddFlags adds additional flags to the subcommands that call this.
func (o *SignOptions) AddFlags(cmd *cobra.Command) {
	o.OpenStackCoreFlags.AddFlags(cmd, viperOpenStackPrefix)

	StringVarWithViper(cmd, &o.ImageID, viperSignPrefix, "image-id", "", "The image ID of the image to sign")
	StringVarWithViper(cmd, &o.VaultURL, viperVaultPrefix, "url", "", "The Vault URL from which you will pull the private key")
	StringVarWithViper(cmd, &o.VaultToken, viperVaultPrefix, "token", "", "The token used to log into vault")
	StringVarWithViper(cmd, &o.VaultMountPath, viperVaultPrefix, "mount-path", "", "The mount path to the secret vault")
	StringVarWithViper(cmd, &o.VaultSecretPath, viperVaultPrefix, "secret-name", "", "The name of the secret within the mount path")

	cmd.MarkFlagsRequiredTogether("url", "token", "mount-path", "secret-name")
}

// SignGenerateOptions contains additional options for the 'generate' subcommand.
type SignGenerateOptions struct {
	Path string
}

// SetOptionsFromViper configures options passed in via viper for the struct.
func (o *SignGenerateOptions) SetOptionsFromViper() {
	o.Path = viper.GetString(fmt.Sprintf("%s.path", viperGeneratePrefix))
}

// AddFlags adds flags to the 'generate' subcommand and binds them to the 'generate' options.
func (o *SignGenerateOptions) AddFlags(cmd *cobra.Command) {
	StringVarWithViper(cmd, &o.Path, viperGeneratePrefix, "path", "/tmp/baski", "A directory location in which to output the generated keys")
}

// SignImageOptions contains additional options for the 'image' subcommand.
type SignImageOptions struct {
	SignOptions

	PrivateKey string
}

// SetOptionsFromViper configures options passed in via viper for the struct.
func (o *SignImageOptions) SetOptionsFromViper() {
	o.SignOptions.SetOptionsFromViper()

	o.PrivateKey = viper.GetString(fmt.Sprintf("%s.private-key", viperSignPrefix))
}

// AddFlags adds flags to the sign 'image' command and binds them to the sign 'image' options.
func (o *SignImageOptions) AddFlags(cmd *cobra.Command) {
	o.SignOptions.AddFlags(cmd)

	StringVarWithViper(cmd, &o.PrivateKey, viperSignPrefix, "private-key", "", "The path to the private key that will be used to sign the image")

	cmd.MarkFlagsMutuallyExclusive("url", "private-key")
}

// SignValidateOptions contains additional options for the 'validate' subcommand.
type SignValidateOptions struct {
	SignOptions

	PublicKey string
	Digest    string
}

// SetOptionsFromViper configures options passed in via viper for the struct.
func (o *SignValidateOptions) SetOptionsFromViper() {
	o.SignOptions.SetOptionsFromViper()

	o.PublicKey = viper.GetString(fmt.Sprintf("%s.public-key", viperSignPrefix))
	o.Digest = viper.GetString(fmt.Sprintf("%s.digest", viperSignPrefix))
}

// AddFlags adds flags to the 'validate' subcommand and binds them to the 'validate' options.
func (o *SignValidateOptions) AddFlags(cmd *cobra.Command) {
	o.SignOptions.AddFlags(cmd)

	StringVarWithViper(cmd, &o.PublicKey, viperSignPrefix, "public-key", "", "The path to the private key that will be used to sign the image")
	StringVarWithViper(cmd, &o.Digest, viperSignPrefix, "digest", "", "The digest to verify")

	cmd.MarkFlagsMutuallyExclusive("url", "public-key")
}
