package utils

import (
	"testing"

	"github.com/rs/zerolog/log"
)

func Test_Zerolog(t *testing.T) {
	log.Info().Str("cd", "aa").Msg("asdf")
}
