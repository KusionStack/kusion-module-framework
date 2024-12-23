package registry

import (
	"errors"
	"fmt"
	"os"
	"runtime"

	"kcl-lang.io/kpm/pkg/client"
	"kcl-lang.io/kpm/pkg/downloader"
)

// NewKusionModuleClient returns a new client for Kusion Module Registry.
func NewKusionModuleClient() (*KusionModuleClient, error) {
	cli, err := client.NewKpmClient()
	if err != nil {
		return nil, err
	}
	cli.DepDownloader = downloader.NewOciDownloader(runtime.GOOS + "/" + runtime.GOARCH)

	return &KusionModuleClient{KpmClient: cli}, nil
}

// NewKusionModuleClientWithCredentials returns a new client for a
// private Kusion Module Registry with the specified registry host, username and password.
// NOTE: the precedence of credential information is as follows.
// 1. The input parameters of this helper function.
// 2. The environment variables of `KUSION_MODULE_REGISTRY_HOST`, `KUSION_MODULE_REGISTRY_USERNAME`
// and `KUSION_MODULE_REGISTRY_PASSWORD`.
// 3. The username and password stored in the credential configs, usually in the `$HOME/.kcl/kpm/.kpm/config/config.json`.
func NewKusionModuleClientWithCredentials(host, username, password string) (*KusionModuleClient, error) {
	cli, err := client.NewKpmClient()
	if err != nil {
		return nil, err
	}
	cli.DepDownloader = downloader.NewOciDownloader(runtime.GOOS + "/" + runtime.GOARCH)

	// Registry host must be set.
	if host == "" {
		if host = os.Getenv(EnvKusionModuleRegistryHost); host == "" {
			return nil, errors.New("registry host must be set")
		}
	}

	// Get the username and password in the local credential config files.
	creds, err := cli.GetCredentials(host)
	if err != nil {
		return nil, err
	}

	// Set the username with environment variable or local credentials if empty.
	if username == "" {
		if username = os.Getenv(EnvKusionModuleRegistryUsername); username == "" {
			username = creds.Username
		}
	}

	// Set the password with environment variable or local credentials if empty.
	if password == "" {
		if password = os.Getenv(EnvKusionModuleRegistryPassword); password == "" {
			password = creds.Password
		}
	}

	// Login to the private oci registry.
	if err = cli.LoginOci(host, username, password); err != nil {
		return nil, err
	}

	return &KusionModuleClient{KpmClient: cli}, nil
}

// DownloadKusionModules downloads the Kusion Module Dependencies declared in the
// `kcl.mod` file under the specified directory.
func (c *KusionModuleClient) DownloadKusionModules(dir string) error {
	info, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("directory [%s] does not exist", dir)
		}

		return fmt.Errorf("failed to stat directory [%s]: %v", dir, err)
	}

	if !info.IsDir() {
		return fmt.Errorf("%s is not a directory", dir)
	}

	// Load the Kusion Module Dependencies from the specified directory.
	kclPkg, err := c.LoadPkgFromPath(dir)
	if err != nil {
		return err
	}

	_, _, err = c.InitGraphAndDownloadDeps(kclPkg)

	return err
}
