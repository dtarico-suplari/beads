package main

import (
	"strings"
	"testing"

	"github.com/steveyegge/beads/internal/types"
)

func TestIsMachineCheckableGate(t *testing.T) {
	tests := []struct {
		name  string
		issue *types.Issue
		want  bool
	}{
		{
			name:  "nil issue",
			issue: nil,
			want:  false,
		},
		{
			name: "non-gate issue",
			issue: &types.Issue{
				IssueType: "task",
			},
			want: false,
		},
		{
			name: "gate with human await type",
			issue: &types.Issue{
				IssueType: "gate",
				AwaitType: "human",
			},
			want: false,
		},
		{
			name: "gate with gh:pr await type",
			issue: &types.Issue{
				IssueType: "gate",
				AwaitType: "gh:pr",
			},
			want: true,
		},
		{
			name: "gate with gh:run await type",
			issue: &types.Issue{
				IssueType: "gate",
				AwaitType: "gh:run",
			},
			want: true,
		},
		{
			name: "gate with timer await type",
			issue: &types.Issue{
				IssueType: "gate",
				AwaitType: "timer",
			},
			want: true,
		},
		{
			name: "gate with bead await type",
			issue: &types.Issue{
				IssueType: "gate",
				AwaitType: "bead",
			},
			want: true,
		},
		{
			name: "gate with script await type",
			issue: &types.Issue{
				IssueType: "gate",
				AwaitType: "script",
			},
			want: true,
		},
		{
			name: "gate with empty await type",
			issue: &types.Issue{
				IssueType: "gate",
				AwaitType: "",
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isMachineCheckableGate(tt.issue)
			if got != tt.want {
				t.Errorf("isMachineCheckableGate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCheckGateSatisfaction_NonGateIssues(t *testing.T) {
	// Non-gate issues should always pass (return nil)
	tests := []struct {
		name  string
		issue *types.Issue
	}{
		{
			name:  "nil issue",
			issue: nil,
		},
		{
			name: "task issue",
			issue: &types.Issue{
				IssueType: "task",
				Title:     "Regular task",
			},
		},
		{
			name: "bug issue",
			issue: &types.Issue{
				IssueType: "bug",
				Title:     "A bug",
			},
		},
		{
			name: "gate with human await (not machine-checkable)",
			issue: &types.Issue{
				IssueType: "gate",
				AwaitType: "human",
				Title:     "Human gate",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := checkGateSatisfaction(tt.issue)
			if err != nil {
				t.Errorf("checkGateSatisfaction() returned error for non-machine-checkable issue: %v", err)
			}
		})
	}
}

func TestCheckGateSatisfaction_GHPRWithoutAwaitID(t *testing.T) {
	// gh:pr gate without an await_id is unsatisfied (no PR to check)
	issue := &types.Issue{
		IssueType: "gate",
		AwaitType: "gh:pr",
		AwaitID:   "",
		Title:     "PR gate without ID",
	}

	err := checkGateSatisfaction(issue)
	if err == nil {
		t.Error("checkGateSatisfaction() should return error for gh:pr gate without await_id")
	}
	if err != nil && !strings.Contains(err.Error(), "no PR number") {
		t.Errorf("error should mention 'no PR number', got: %v", err)
	}
}

func TestCheckGateSatisfaction_GHRunWithoutAwaitID(t *testing.T) {
	// gh:run gate without an await_id is unsatisfied (no run to check)
	issue := &types.Issue{
		IssueType: "gate",
		AwaitType: "gh:run",
		AwaitID:   "",
		Title:     "Run gate without ID",
	}

	err := checkGateSatisfaction(issue)
	if err == nil {
		t.Error("checkGateSatisfaction() should return error for gh:run gate without await_id")
	}
	if err != nil && !strings.Contains(err.Error(), "no run ID") {
		t.Errorf("error should mention 'no run ID', got: %v", err)
	}
}

func TestCheckGateSatisfaction_BeadGateInvalidFormat(t *testing.T) {
	// bead gate with invalid await_id should return an error
	issue := &types.Issue{
		IssueType: "gate",
		AwaitType: "bead",
		AwaitID:   "invalid-no-colon",
		Title:     "Bead gate with bad format",
	}

	err := checkGateSatisfaction(issue)
	if err == nil {
		t.Error("checkGateSatisfaction() should return error for bead gate with invalid await_id format")
	}
}

func TestCheckGateSatisfaction_ErrorMessageFormat(t *testing.T) {
	// Verify error messages contain the force override hint
	issue := &types.Issue{
		IssueType: "gate",
		AwaitType: "bead",
		AwaitID:   "invalid-no-colon",
		Title:     "Test gate",
	}

	err := checkGateSatisfaction(issue)
	if err == nil {
		t.Fatal("expected error")
	}
	errMsg := err.Error()
	if !strings.Contains(errMsg, "--force") {
		t.Errorf("error message should mention --force, got: %s", errMsg)
	}
	if !strings.Contains(errMsg, "gate condition not satisfied") {
		t.Errorf("error message should mention 'gate condition not satisfied', got: %s", errMsg)
	}
}

func TestCheckGateSatisfaction_ScriptGatePassesThrough(t *testing.T) {
	// Script gates are handled by checkScriptGate (which runs first).
	// checkGateSatisfaction should return nil for script gates since
	// they've already been validated.
	issue := &types.Issue{
		IssueType: "gate",
		AwaitType: "script",
		AwaitID:   "audit-acceptance test-issue",
		Title:     "Script gate",
	}

	err := checkGateSatisfaction(issue)
	if err != nil {
		t.Errorf("checkGateSatisfaction() should return nil for script gate, got: %v", err)
	}
}

func TestCheckScript_PassingScript(t *testing.T) {
	resolved, reason, _ := checkScript("true")
	if !resolved {
		t.Errorf("checkScript(true) should resolve, got reason: %s", reason)
	}
}

func TestCheckScript_FailingScript(t *testing.T) {
	resolved, reason, _ := checkScript("false")
	if resolved {
		t.Error("checkScript(false) should not resolve")
	}
	if reason == "" {
		t.Error("expected reason to be set for failing script")
	}
}

func TestCheckScript_ScriptWithArgs(t *testing.T) {
	resolved, reason, _ := checkScript("test 1 -eq 1")
	if !resolved {
		t.Errorf("checkScript('test 1 -eq 1') should resolve, got reason: %s", reason)
	}
}

func TestCheckScript_ScriptWithFailingArgs(t *testing.T) {
	resolved, reason, _ := checkScript("test 1 -eq 2")
	if resolved {
		t.Error("checkScript('test 1 -eq 2') should not resolve")
	}
	if !strings.Contains(reason, "exit") {
		t.Errorf("reason should mention exit status, got: %s", reason)
	}
}

func TestCheckScript_EmptyCommand(t *testing.T) {
	resolved, reason, _ := checkScript("")
	if resolved {
		t.Error("checkScript('') should not resolve")
	}
	if !strings.Contains(reason, "empty") {
		t.Errorf("reason should mention empty command, got: %s", reason)
	}
}

func TestCheckScript_CapturesStdout(t *testing.T) {
	resolved, _, stdout := checkScript("echo 'Plan generated at /tmp/plan.md'")
	if !resolved {
		t.Error("checkScript(echo) should resolve")
	}
	if stdout != "Plan generated at /tmp/plan.md" {
		t.Errorf("expected stdout to be captured, got: %q", stdout)
	}
}

func TestCheckScript_EmptyStdoutOnPass(t *testing.T) {
	resolved, _, stdout := checkScript("true")
	if !resolved {
		t.Error("checkScript(true) should resolve")
	}
	if stdout != "" {
		t.Errorf("expected empty stdout for 'true', got: %q", stdout)
	}
}

func TestCheckScript_NoStdoutOnFail(t *testing.T) {
	resolved, _, stdout := checkScript("echo 'some output' && false")
	if resolved {
		t.Error("checkScript should not resolve")
	}
	if stdout != "" {
		t.Errorf("expected empty stdout on failure, got: %q", stdout)
	}
}

func TestCheckScript_NonexistentCommand(t *testing.T) {
	resolved, reason, _ := checkScript("nonexistent-command-xyz-12345") //nolint:dogsled
	if resolved {
		t.Error("checkScript with nonexistent command should not resolve")
	}
	if reason == "" {
		t.Error("expected reason for nonexistent command")
	}
}

func TestCheckScriptGate_Pass(t *testing.T) {
	issue := &types.Issue{
		IssueType: "gate",
		AwaitType: "script",
		AwaitID:   "true",
		Title:     "Script gate that passes",
	}

	err := checkScriptGate(issue)
	if err != nil {
		t.Errorf("checkScriptGate() should pass for script gate with 'true': %v", err)
	}
}

func TestCheckScriptGate_NonGateIssue(t *testing.T) {
	issue := &types.Issue{
		IssueType: "task",
		Title:     "Regular task",
	}

	err := checkScriptGate(issue)
	if err != nil {
		t.Errorf("checkScriptGate() should return nil for non-gate issue: %v", err)
	}
}

func TestCheckScriptGate_NilIssue(t *testing.T) {
	err := checkScriptGate(nil)
	if err != nil {
		t.Errorf("checkScriptGate() should return nil for nil issue: %v", err)
	}
}

func TestCheckGateSatisfaction_ScriptGateFail(t *testing.T) {
	// Script gates use checkScriptGate (non-bypassable), not checkGateSatisfaction.
	// checkGateSatisfaction returns nil for script gates (they're handled separately).
	issue := &types.Issue{
		IssueType: "gate",
		AwaitType: "script",
		AwaitID:   "false",
		Title:     "Script gate that fails",
	}

	err := checkScriptGate(issue)
	if err == nil {
		t.Error("checkScriptGate() should return error for script gate with 'false'")
	}
	if err != nil && !strings.Contains(err.Error(), "cannot be overridden") {
		t.Errorf("error should mention 'cannot be overridden', got: %v", err)
	}
}

func TestCheckScriptGate_EmptyAwaitID(t *testing.T) {
	issue := &types.Issue{
		IssueType: "gate",
		AwaitType: "script",
		AwaitID:   "",
		Title:     "Script gate without command",
	}

	err := checkScriptGate(issue)
	if err == nil {
		t.Error("checkScriptGate() should return error for script gate without await_id")
	}
}
