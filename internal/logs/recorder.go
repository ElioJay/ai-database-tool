package logs

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Recorder struct {
	dir string
}

func New(configDir string) (*Recorder, error) {
	dir := filepath.Join(configDir, "logs")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, fmt.Errorf("创建日志目录失败: %w", err)
	}
	return &Recorder{dir: dir}, nil
}

func (r *Recorder) Dir() string { return r.dir }

func (r *Recorder) LogUser(input string) {
	r.write("USER", input)
}

func (r *Recorder) LogAI(explanation, sql string) {
	r.write("AI", strings.TrimSpace(explanation)+" | SQL: "+singleLine(sql))
}

func (r *Recorder) LogExec(sql string, ok bool, rowsAffected int64, durationMS int64) {
	status := "OK"
	if !ok {
		status = "FAIL"
	}
	r.write("EXEC["+status+"]", fmt.Sprintf("rows=%d duration_ms=%d sql=%s", rowsAffected, durationMS, singleLine(sql)))
}

func (r *Recorder) LogSystem(msg string) {
	r.write("SYSTEM", msg)
}

func (r *Recorder) write(tag, content string) {
	if r == nil {
		return
	}
	now := time.Now()
	path := filepath.Join(r.dir, now.Format("2006-01-02")+".log")
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return
	}
	defer f.Close()
	fmt.Fprintf(f, "[%s] [%s] %s\n", now.Format("15:04:05"), tag, content)
}

func singleLine(s string) string {
	return strings.Join(strings.Fields(s), " ")
}
