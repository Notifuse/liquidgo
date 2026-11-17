package integration

import "github.com/Notifuse/liquidgo/v5/liquid"

// mustEnv builds an environment the code knows to be valid.
func mustEnv(config liquid.EnvironmentConfig) *liquid.Environment {
	env, err := liquid.NewEnvironment(config)
	if err != nil {
		panic(err)
	}
	return env
}
