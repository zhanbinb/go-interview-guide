package main

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

// TestGoBuild 验证 go build 命令
func TestGoBuild(t *testing.T) {
	tmpFile := "/tmp/test-compile-build"
	out, err := exec.Command("go", "build", "-o", tmpFile, ".").CombinedOutput()
	if err != nil {
		t.Fatalf("go build failed: %v, output: %s", err, out)
	}
	defer os.Remove(tmpFile)

	if _, err := os.Stat(tmpFile); err != nil {
		t.Errorf("binary not created: %v", err)
	}
}

// TestGoRun 验证 go run 命令
func TestGoRun(t *testing.T) {
	out, err := exec.Command("go", "run", ".").CombinedOutput()
	if err != nil {
		t.Fatalf("go run failed: %v, output: %s", err, out)
	}
	// 应该输出菜单
	if !strings.Contains(string(out), "Go 编译演示") {
		t.Errorf("expected menu output, got: %s", out)
	}
}

// TestGoEnv 验证 go env 能拿到关键变量
func TestGoEnv(t *testing.T) {
	for _, key := range []string{"GOROOT", "GOPATH", "GOOS", "GOARCH"} {
		out, err := exec.Command("go", "env", key).Output()
		if err != nil {
			t.Errorf("go env %s failed: %v", key, err)
			continue
		}
		val := strings.TrimSpace(string(out))
		if val == "" {
			t.Errorf("go env %s is empty", key)
		}
	}
}

// TestCrossCompile 验证交叉编译（重要 feature）
func TestCrossCompile(t *testing.T) {
	tmpFile := "/tmp/test-compile-cross"
	defer os.Remove(tmpFile)

	cmd := exec.Command("go", "build", "-o", tmpFile, ".")
	cmd.Env = append(os.Environ(),
		"GOOS=linux",
		"GOARCH=amd64",
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("cross compile failed: %v, output: %s", err, out)
	}
	if _, err := os.Stat(tmpFile); err != nil {
		t.Errorf("cross-compiled binary not created: %v", err)
	}
}
