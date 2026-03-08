package config

type Config struct {
	RootDir  string
	BuildCmd string
	ExecCmd  string
}

func NewConfig(root, build, exec string) Config {
	return Config{
		RootDir:  root,
		BuildCmd: build,
		ExecCmd:  exec,
	}
}
