//go:build std || viridian

package viridian

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/clc/secrets"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/prompt"
	"github.com/hazelcast/hazelcast-commandline-client/internal/viridian"
)

const (
	propAPIKey    = "api-key"
	propAPISecret = "api-secret"
	secretPrefix  = "viridian"
	propForce     = "force"
)

type LoginCmd struct{}

func (cm LoginCmd) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("login")
	short := "Logs in to Viridian using the given API key and API secret"
	long := fmt.Sprintf(`Logs in to Viridian using the given API key and API secret.
If not specified, the key and the secret will be asked in a prompt.

Alternatively, you can use the following environment variables:
* %s
* %s
`, viridian.EnvAPIKey, viridian.EnvAPISecret)
	cc.SetCommandHelp(long, short)
	cc.AddStringFlag(propAPIKey, "", "", false, "Viridian API Key")
	cc.AddStringFlag(propAPISecret, "", "", false, "Viridian API Secret")
	cc.SetPositionalArgCount(0, 0)
	if viridian.InternalOpsEnabled() {
		cc.SetCommandGroup("viridian")
	}
	return nil
}

func (cm LoginCmd) Exec(ctx context.Context, ec plug.ExecContext) error {
	key, secret, err := apiKeySecret(ec)
	if err != nil {
		return err
	}
	token, err := cm.retrieveToken(ctx, ec, key, secret)
	if err != nil {
		return err
	}

	if err = secrets.Save(ctx, viridian.APIClass(), secretPrefix, key, secret, token); err != nil {
		return err
	}
	ec.PrintlnUnnecessary("")
	ec.PrintlnUnnecessary("Viridian token was fetched and saved.")
	return nil
}

func (cm LoginCmd) retrieveToken(ctx context.Context, ec plug.ExecContext, key, secret string) (string, error) {
	ti, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		sp.SetText("Logging in")
		api, err := viridian.Login(ctx, secretPrefix, key, secret)
		if err != nil {
			return nil, err
		}
		return api.Token, err
	})
	if err != nil {
		return "", handleErrorResponse(ec, err)
	}
	stop()
	return ti.(string), nil
}

func apiKeySecret(ec plug.ExecContext) (key, secret string, err error) {
	pr := prompt.New(ec.Stdin(), ec.Stdout())
	key = ec.Props().GetString(propAPIKey)
	if key == "" {
		key = os.Getenv(viridian.EnvAPIKey)
	}
	if key == "" {
		key, err = pr.Text("API Key    : ")
		if err != nil {
			return "", "", fmt.Errorf("reading API key: %w", err)
		}
	}
	if key == "" {
		return "", "", errors.New("api key cannot be blank")
	}
	secret = ec.Props().GetString(propAPISecret)
	if secret == "" {
		secret = os.Getenv(viridian.EnvAPISecret)
	}
	if secret == "" {
		secret, err = pr.Password("API Secret : ")
		if err != nil {
			return "", "", fmt.Errorf("reading API secret: %w", err)
		}
	}
	if secret == "" {
		return "", "", errors.New("api secret cannot be blank")
	}
	return key, secret, nil
}

func init() {
	if viridian.InternalOpsEnabled() {
		Must(plug.Registry.RegisterCommand("login", &LoginCmd{}))
	} else {
		Must(plug.Registry.RegisterCommand("viridian:login", &LoginCmd{}))
	}
}
