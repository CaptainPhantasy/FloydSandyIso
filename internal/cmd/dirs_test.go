package cmd

import (
	"bytes"
	"path/filepath"
	"sync"
	"testing"

	"github.com/CaptainPhantasy/FloydSandyIso/internal/config"
	"github.com/stretchr/testify/require"
)

var dirsTestMutex sync.Mutex

func setupDirsTest(t *testing.T) {
	dirsTestMutex.Lock()
	t.Cleanup(func() {
		config.ResetInstance()
		dirsTestMutex.Unlock()
	})
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpDir, "config"))
	t.Setenv("XDG_DATA_HOME", filepath.Join(tmpDir, "data"))
	t.Setenv("FLOYD_GLOBAL_DATA", "")
	config.ResetInstance()
}

func TestDirs(t *testing.T) {
	setupDirsTest(t)
	var b bytes.Buffer
	dirsCmd.SetOut(&b)
	dirsCmd.SetErr(&b)
	dirsCmd.SetIn(bytes.NewReader(nil))
	dirsCmd.Run(dirsCmd, nil)
	// Just verify it outputs something
	require.NotEmpty(t, b.String())
}

func TestConfigDir(t *testing.T) {
	setupDirsTest(t)
	var b bytes.Buffer
	configDirCmd.SetOut(&b)
	configDirCmd.SetErr(&b)
	configDirCmd.SetIn(bytes.NewReader(nil))
	configDirCmd.Run(configDirCmd, nil)
	require.NotEmpty(t, b.String())
}

func TestDataDir(t *testing.T) {
	setupDirsTest(t)
	var b bytes.Buffer
	dataDirCmd.SetOut(&b)
	dataDirCmd.SetErr(&b)
	dataDirCmd.SetIn(bytes.NewReader(nil))
	dataDirCmd.Run(dataDirCmd, nil)
	require.NotEmpty(t, b.String())
}
