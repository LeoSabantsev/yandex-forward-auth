package actions

import (
	"os"
	"testing"

	"github.com/gobuffalo/suite/v4"

	"yandex_forward_auth/internal/config"
)

type ActionSuite struct {
	*suite.Action
}

func init() {
	os.Setenv("YA_AUTH_BASE_URL", "http://auth.example.com")
}

func Test_ActionSuite(t *testing.T) {
	action, err := suite.NewActionWithFixtures(App(config.Load()), os.DirFS("../fixtures"))
	if err != nil {
		t.Fatal(err)
	}

	as := &ActionSuite{
		Action: action,
	}
	suite.Run(t, as)
}
