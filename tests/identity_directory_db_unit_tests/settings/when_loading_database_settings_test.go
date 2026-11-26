package settings_test

import (
	"testing"

	"github.com/6022-labs/agentic-identity-directory/src/identity_directory_db/settings"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/v2"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestWhenLoadingDatabaseSettings(t *testing.T) {
	t.Parallel()

	t.Run("Given a padded url", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")
		_ = k.Load(confmap.Provider(map[string]any{
			"database.url": "  host=localhost dbname=identity_directory  ",
		}, "."), nil)

		t.Run("Should load it trimmed", func(t *testing.T) {
			t.Parallel()

			result := settings.NewDatabaseSettings(zap.NewNop(), k)

			assert.Equal(t, "host=localhost dbname=identity_directory", result.Url)
		})
	})

}
