package config_test

import (
	"github.com/swift9/ares-config"
	"os"
	"testing"
)

func TestEnv(t *testing.T) {
	println(os.Getenv("ARES_CONFIG_FILE"))
}

func TestConfigString(t *testing.T) {
	println(config.GetString("ob.a", "777"))
}

func TestConfigNacosString(t *testing.T) {
	println(config.GetString("pullTaskRequestUrl", "pullTaskRequestUrl"))

}

func TestConfigBool(t *testing.T) {
	if config.GetBool("ob.bool", false) {
		t.Error("error")
	}

	if config.GetBool("ob.bool", false) == false {
		t.Log("ha~")
		println(1)
	}
}

func TestConfigInt64(t *testing.T) {
	println(config.GetInt64("test", 1))
	println(config.GetInt64("test2", 1))
}
