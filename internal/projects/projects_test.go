package projects

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/CaptainPhantasy/FloydSandyIso/internal/config"
)

var testMutex sync.Mutex

func setupTest(t *testing.T) {
	testMutex.Lock()
	t.Cleanup(func() {
		config.ResetInstance()
		testMutex.Unlock()
	})
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)
	t.Setenv("FLOYD_GLOBAL_DATA", "")
	config.ResetInstance()
}

func TestRegisterAndList(t *testing.T) {
	setupTest(t)
	Register("/home/user/project1", "/home/user/project1/.floyd")
	projects, _ := List()
	if len(projects) != 1 {
		t.Fatalf("Expected 1 project, got %d", len(projects))
	}
}

func TestRegisterUpdatesExisting(t *testing.T) {
	setupTest(t)
	Register("/home/user/project1", "/home/user/project1/.floyd")
	time.Sleep(10 * time.Millisecond)
	Register("/home/user/project1", "/home/user/project1/.floyd-new")
	projects, _ := List()
	if len(projects) != 1 {
		t.Fatalf("Expected 1 project after update, got %d", len(projects))
	}
}

func TestLoadEmptyFile(t *testing.T) {
	setupTest(t)
	projects, _ := List()
	if len(projects) != 0 {
		t.Errorf("Expected 0 projects, got %d", len(projects))
	}
}

func TestProjectsFilePath(t *testing.T) {
	setupTest(t)
	tmpDir := os.Getenv("XDG_DATA_HOME")
	expected := filepath.Join(tmpDir, "floyd", "projects.json")
	actual := projectsFilePath()
	if actual != expected {
		t.Errorf("Expected %s, got %s", expected, actual)
	}
}

func TestRegisterWithParentDataDir(t *testing.T) {
	setupTest(t)
	Register("/home/user/monorepo/packages/app", "/home/user/monorepo/.floyd")
	projects, _ := List()
	if len(projects) != 1 {
		t.Fatalf("Expected 1 project, got %d", len(projects))
	}
}

func TestRegisterWithExternalDataDir(t *testing.T) {
	setupTest(t)
	Register("/home/user/project", "/var/data/floyd/myproject")
	projects, _ := List()
	if len(projects) != 1 {
		t.Fatalf("Expected 1 project, got %d", len(projects))
	}
}
