package agent

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/CaptainPhantasy/FloydSandyIso/internal/config"
	"github.com/CaptainPhantasy/FloydSandyIso/internal/db"
	"github.com/CaptainPhantasy/FloydSandyIso/internal/message"
	"github.com/CaptainPhantasy/FloydSandyIso/internal/session"
	"github.com/stretchr/testify/require"
)

// TestCreateHandoffFile_AppendBehavior verifies that createHandoffFile
// APPENDS to existing HANDOFF.md instead of OVERWRITING it.
// This is critical for preserving Terminal Shadow logs.
func TestCreateHandoffFile_AppendBehavior(t *testing.T) {
	ctx := context.Background()

	// Create temp directory with existing HANDOFF.md
	workingDir := filepath.Join(t.TempDir(), "project")
	err := os.MkdirAll(workingDir, 0755)
	require.NoError(t, err)

	handoffPath := filepath.Join(workingDir, "HANDOFF.md")

	// Simulate Terminal Shadow content that should be preserved
	existingContent := `# Project Handoff Document

## BOOT LOG

- 2026-02-28 10:00:00 UTC — Shadow daemon started

## LOST CONTEXT INSURANCE / Decision Log

### Important Decision
We decided to use SQLite for storage.
`
	err = os.WriteFile(handoffPath, []byte(existingContent), 0644)
	require.NoError(t, err)

	// Set up database and services
	conn, err := db.Connect(ctx, t.TempDir())
	require.NoError(t, err)
	defer conn.Close()

	q := db.New(conn)
	sessions := session.NewService(q, conn)
	messages := message.NewService(q, conn)

	// Create a test session with todos
	sess, err := sessions.Create(ctx, "Test Session for Handoff")
	require.NoError(t, err)

	// Save session with todos
	sess.Todos = []session.Todo{
		{Content: "Implement feature X", Status: session.TodoStatusInProgress, ActiveForm: "Implementing feature X"},
		{Content: "Write tests", Status: session.TodoStatusPending, ActiveForm: "Writing tests"},
	}
	_, err = sessions.Save(ctx, sess)
	require.NoError(t, err)

	// Initialize config for the working directory
	cfg, err := config.Init(workingDir, "", false)
	require.NoError(t, err)

	// Create session agent
	agent := NewSessionAgent(SessionAgentOptions{
		Sessions: sessions,
		Messages: messages,
	}).(*sessionAgent)
	agent.cfg = cfg

	// Call createHandoffFile
	err = agent.createHandoffFile(ctx, sess.ID)
	require.NoError(t, err)

	// Read the resulting file
	resultContent, err := os.ReadFile(handoffPath)
	require.NoError(t, err)
	resultStr := string(resultContent)

	// CRITICAL VERIFICATION: Original content must be preserved
	require.Contains(t, resultStr, "## BOOT LOG", "Original BOOT LOG section must be preserved")
	require.Contains(t, resultStr, "Shadow daemon started", "Original shadow content must be preserved")
	require.Contains(t, resultStr, "## LOST CONTEXT INSURANCE / Decision Log", "Original Decision Log must be preserved")
	require.Contains(t, resultStr, "We decided to use SQLite", "Original decision content must be preserved")

	// Verify handoff section was APPENDED
	require.Contains(t, resultStr, "## SESSION HANDOFF", "Handoff section should be appended")
	require.Contains(t, resultStr, sess.ID, "Session ID should be in handoff section")
	require.Contains(t, resultStr, "Test Session for Handoff", "Session title should be in handoff section")
	require.Contains(t, resultStr, "Implement feature X", "Active todos should be in handoff section")

	// Verify the original content appears BEFORE the handoff section (APPEND not OVERWRITE)
	bootLogIndex := strings.Index(resultStr, "## BOOT LOG")
	handoffIndex := strings.Index(resultStr, "## SESSION HANDOFF")
	require.Less(t, bootLogIndex, handoffIndex, "Original content must appear before handoff section (APPEND not OVERWRITE)")
}

// TestCreateHandoffFile_CreatesFileIfNotExists verifies that createHandoffFile
// creates a new HANDOFF.md if one doesn't exist.
func TestCreateHandoffFile_CreatesFileIfNotExists(t *testing.T) {
	ctx := context.Background()

	// Temp directory WITHOUT existing HANDOFF.md
	workingDir := filepath.Join(t.TempDir(), "project")
	err := os.MkdirAll(workingDir, 0755)
	require.NoError(t, err)

	handoffPath := filepath.Join(workingDir, "HANDOFF.md")

	// Verify file doesn't exist
	_, err = os.Stat(handoffPath)
	require.True(t, os.IsNotExist(err), "HANDOFF.md should not exist initially")

	// Set up database and services
	conn, err := db.Connect(ctx, t.TempDir())
	require.NoError(t, err)
	defer conn.Close()

	q := db.New(conn)
	sessions := session.NewService(q, conn)
	messages := message.NewService(q, conn)

	// Create a test session
	sess, err := sessions.Create(ctx, "Test Session")
	require.NoError(t, err)

	// Initialize config
	cfg, err := config.Init(workingDir, "", false)
	require.NoError(t, err)

	// Create session agent
	agent := NewSessionAgent(SessionAgentOptions{
		Sessions: sessions,
		Messages: messages,
	}).(*sessionAgent)
	agent.cfg = cfg

	// Call createHandoffFile
	err = agent.createHandoffFile(ctx, sess.ID)
	require.NoError(t, err)

	// Verify file was created
	_, err = os.Stat(handoffPath)
	require.NoError(t, err, "HANDOFF.md should be created")

	// Verify content
	resultContent, err := os.ReadFile(handoffPath)
	require.NoError(t, err)
	require.Contains(t, string(resultContent), "## SESSION HANDOFF")
}

// TestCreateHandoffFile_MultipleHandoffs verifies that multiple handoffs
// append to the same file without overwriting previous handoffs.
func TestCreateHandoffFile_MultipleHandoffs(t *testing.T) {
	ctx := context.Background()

	workingDir := filepath.Join(t.TempDir(), "project")
	err := os.MkdirAll(workingDir, 0755)
	require.NoError(t, err)
	handoffPath := filepath.Join(workingDir, "HANDOFF.md")

	// Set up database and services
	conn, err := db.Connect(ctx, t.TempDir())
	require.NoError(t, err)
	defer conn.Close()

	q := db.New(conn)
	sessions := session.NewService(q, conn)
	messages := message.NewService(q, conn)

	// Initialize config
	cfg, err := config.Init(workingDir, "", false)
	require.NoError(t, err)

	// Create first session and handoff
	sess1, err := sessions.Create(ctx, "First Session")
	require.NoError(t, err)

	agent := NewSessionAgent(SessionAgentOptions{
		Sessions: sessions,
		Messages: messages,
	}).(*sessionAgent)
	agent.cfg = cfg

	err = agent.createHandoffFile(ctx, sess1.ID)
	require.NoError(t, err)

	// Create second session and handoff
	sess2, err := sessions.Create(ctx, "Second Session")
	require.NoError(t, err)

	err = agent.createHandoffFile(ctx, sess2.ID)
	require.NoError(t, err)

	// Read file and verify both handoffs exist
	resultContent, err := os.ReadFile(handoffPath)
	require.NoError(t, err)
	resultStr := string(resultContent)

	require.Contains(t, resultStr, sess1.ID, "First session handoff must be preserved")
	require.Contains(t, resultStr, sess2.ID, "Second session handoff must exist")

	// Count occurrences of "## SESSION HANDOFF" to verify two handoffs
	count := strings.Count(resultStr, "## SESSION HANDOFF")
	require.Equal(t, 2, count, "Should have exactly 2 handoff sections")
}

// TestCreateHandoffFile_TodosFormattedCorrectly verifies that todos are
// formatted with correct status icons.
func TestCreateHandoffFile_TodosFormattedCorrectly(t *testing.T) {
	ctx := context.Background()

	workingDir := filepath.Join(t.TempDir(), "project")
	err := os.MkdirAll(workingDir, 0755)
	require.NoError(t, err)
	handoffPath := filepath.Join(workingDir, "HANDOFF.md")

	// Set up database and services
	conn, err := db.Connect(ctx, t.TempDir())
	require.NoError(t, err)
	defer conn.Close()

	q := db.New(conn)
	sessions := session.NewService(q, conn)
	messages := message.NewService(q, conn)

	// Create session with todos of different statuses
	sess, err := sessions.Create(ctx, "Todo Test Session")
	require.NoError(t, err)

	sess.Todos = []session.Todo{
		{Content: "Completed task", Status: session.TodoStatusCompleted, ActiveForm: "Completing task"},
		{Content: "In-progress task", Status: session.TodoStatusInProgress, ActiveForm: "Working on task"},
		{Content: "Pending task", Status: session.TodoStatusPending, ActiveForm: "Starting task"},
	}
	_, err = sessions.Save(ctx, sess)
	require.NoError(t, err)

	// Initialize config
	cfg, err := config.Init(workingDir, "", false)
	require.NoError(t, err)

	agent := NewSessionAgent(SessionAgentOptions{
		Sessions: sessions,
		Messages: messages,
	}).(*sessionAgent)
	agent.cfg = cfg

	err = agent.createHandoffFile(ctx, sess.ID)
	require.NoError(t, err)

	// Verify todo formatting
	resultContent, err := os.ReadFile(handoffPath)
	require.NoError(t, err)
	resultStr := string(resultContent)

	require.Contains(t, resultStr, "[x] Completed task", "Completed todo should have [x] prefix")
	require.Contains(t, resultStr, "[~] In-progress task", "In-progress todo should have [~] prefix")
	require.Contains(t, resultStr, "[ ] Pending task", "Pending todo should have [ ] prefix")
}
