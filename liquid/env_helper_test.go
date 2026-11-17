package liquid

// mustEnv builds an environment the test knows to be valid.
func mustEnv(config EnvironmentConfig) *Environment {
	env, err := NewEnvironment(config)
	if err != nil {
		panic(err)
	}
	return env
}
