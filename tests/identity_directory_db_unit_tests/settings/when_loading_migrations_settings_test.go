package settings_test

import (
	"testing"

	"github.com/6022-labs/agentic-identity-directory/src/identity_directory_db/settings"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/v2"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestWhenLoadingMigrationsSettings(t *testing.T) {
	t.Parallel()

	t.Run("Given an existing folder_path", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		k := koanf.New(".")
		_ = k.Load(confmap.Provider(map[string]any{"migrations.folder_path": dir}, "."), nil)

		t.Run("Should load the path verbatim", func(t *testing.T) {
			t.Parallel()

			result := settings.NewMigrationsSettings(zap.NewNop(), k)

			assert.Equal(t, dir, result.FolderPath)
		})
	})

	t.Run("Given a padded folder_path", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		k := koanf.New(".")
		_ = k.Load(confmap.Provider(map[string]any{"migrations.folder_path": "  " + dir + "  "}, "."), nil)

		t.Run("Should trim it", func(t *testing.T) {
			t.Parallel()

			result := settings.NewMigrationsSettings(zap.NewNop(), k)

			assert.Equal(t, dir, result.FolderPath)
		})
	})
}
