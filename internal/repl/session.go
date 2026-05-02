package repl

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/chzyer/readline"

	"github.com/aidbt-tool/aidbt/internal/config"
	"github.com/aidbt-tool/aidbt/internal/database"
	"github.com/aidbt-tool/aidbt/internal/logs"
	"github.com/aidbt-tool/aidbt/internal/prompt"
	"github.com/aidbt-tool/aidbt/internal/provider"
	"github.com/aidbt-tool/aidbt/internal/render"
	"github.com/aidbt-tool/aidbt/internal/sqlplan"
)

type Session struct {
	cfg           *config.Config
	configDir     config.ConfigDir
	prov          provider.Provider
	db            *database.DB
	connName      string
	conn          config.ConnectionConfig
	schemaSummary string
	recorder      *logs.Recorder
	history       []provider.Message
}

func NewSession(cfg *config.Config, cd config.ConfigDir) (*Session, error) {
	providerName, pc, err := cfg.CurrentProvider()
	if err != nil {
		return nil, err
	}
	prov, err := provider.Build(providerName, pc)
	if err != nil {
		return nil, err
	}
	connName, cc, err := cfg.CurrentConnection()
	if err != nil {
		return nil, err
	}
	db, err := database.Open(connName, toDBConn(cc))
	if err != nil {
		return nil, err
	}
	rec, err := logs.New(cd.Path)
	if err != nil {
		db.Close()
		return nil, err
	}
	s := &Session{cfg: cfg, configDir: cd, prov: prov, db: db, connName: connName, conn: cc, recorder: rec}
	if err := s.RefreshSchema(context.Background()); err != nil {
		rec.LogSystem("schema 探测失败: " + err.Error())
		s.schemaSummary = "schema 探测失败：" + err.Error()
	}
	return s, nil
}

func (s *Session) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

func (s *Session) Run() error {
	defer s.Close()
	rl, err := readline.NewEx(&readline.Config{
		Prompt:          "> ",
		HistoryFile:     filepath.Join(s.recorder.Dir(), "readline"),
		InterruptPrompt: "^C",
		EOFPrompt:       "/exit",
	})
	if err != nil {
		return err
	}
	defer rl.Close()

	fmt.Printf("aidbt 已启动（provider: %s，连接: %s/%s）\n", s.prov.Name(), s.connName, s.conn.Type)
	fmt.Println("输入中文数据库需求，或输入 /help 查看帮助")
	for {
		line, err := rl.Readline()
		if err != nil {
			break
		}
		input := strings.TrimSpace(line)
		if input == "" {
			continue
		}
		if meta := HandleMeta(input); meta.Handled {
			if meta.ShouldExit {
				break
			}
			s.applyMeta(meta)
			continue
		}
		if err := s.HandleQuery(context.Background(), input); err != nil {
			fmt.Printf("错误：%v\n", err)
		}
	}
	fmt.Println("再见！")
	return nil
}

func (s *Session) HandleQuery(ctx context.Context, input string) error {
	system := prompt.Build(prompt.Context{
		DBType:       s.conn.Type,
		Connection:   s.connName,
		Schema:       schemaName(s.conn),
		SchemaDigest: s.schemaSummary,
		MaxRows:      s.cfg.UI.MaxRows,
	})
	msgs := append([]provider.Message{{Role: "system", Content: system}}, s.history...)
	msgs = append(msgs, provider.Message{Role: "user", Content: input})

	ch, err := s.prov.Stream(ctx, msgs)
	if err != nil {
		return err
	}
	fmt.Println("AI：")
	raw, err := provider.Collect(ch, func(delta string) {
		if s.cfg.UI.Stream {
			fmt.Print(delta)
		}
	})
	if s.cfg.UI.Stream {
		fmt.Println()
	}
	if err != nil {
		return err
	}
	plan, err := sqlplan.ParseAIResponse(raw)
	if err != nil {
		return err
	}
	if plan.SQL == "" {
		fmt.Println(plan.Explanation)
		return nil
	}
	s.recorder.LogUser(input)
	s.recorder.LogAI(plan.Explanation, plan.SQL)

	if !askConfirm(plan, s.connName, s.conn.Type) {
		fmt.Println("已取消。")
		return nil
	}
	result, err := s.execute(ctx, plan)
	if err != nil {
		s.recorder.LogExec(plan.SQL, false, 0, 0)
		return err
	}
	s.recorder.LogExec(plan.SQL, true, result.RowsAffected, result.DurationMS)
	s.history = append(s.history,
		provider.Message{Role: "user", Content: input},
		provider.Message{Role: "assistant", Content: raw},
	)
	return nil
}

func (s *Session) execute(ctx context.Context, plan *sqlplan.Plan) (database.QueryResult, error) {
	if plan.Local.Kind == sqlplan.StatementSelect {
		result, err := s.db.Query(ctx, plan.SQL, s.cfg.UI.MaxRows)
		if err != nil {
			return database.QueryResult{}, err
		}
		fmt.Print(render.Table(result.Columns, result.Rows, render.TableOptions{MaxRows: s.cfg.UI.MaxRows, MaxCellWidth: 40}))
		fmt.Printf("耗时：%dms\n", result.DurationMS)
		if result.Truncated {
			fmt.Printf("结果超过 %d 行，已截断显示。\n", s.cfg.UI.MaxRows)
		}
		return result, nil
	}
	result, err := s.db.Exec(ctx, plan.SQL)
	if err != nil {
		return database.QueryResult{}, err
	}
	fmt.Printf("执行完成，影响行数：%d，耗时：%dms\n", result.RowsAffected, result.DurationMS)
	return result, nil
}

func (s *Session) RefreshSchema(ctx context.Context) error {
	if err := s.db.Ping(ctx); err != nil {
		return err
	}
	schema, err := database.ProbeSchema(ctx, s.db.SQL, s.db.Spec, toDBConn(s.conn))
	if err != nil {
		return err
	}
	s.schemaSummary = schema.Summary(database.SchemaOptions{
		Include:            s.conn.Include,
		Exclude:            s.conn.Exclude,
		MaxTables:          80,
		MaxColumnsPerTable: 24,
	})
	return nil
}

func (s *Session) applyMeta(meta MetaResult) {
	if meta.ShowHelp {
		printHelp()
	}
	if meta.ResetHistory {
		s.history = nil
		s.recorder.LogSystem("对话历史已清空")
		fmt.Println("已清空对话历史。")
	}
	if meta.RefreshSchema {
		if err := s.RefreshSchema(context.Background()); err != nil {
			fmt.Printf("schema 刷新失败：%v\n", err)
		} else {
			fmt.Println("schema 已刷新。")
		}
	}
}

func toDBConn(cc config.ConnectionConfig) database.ConnectionConfig {
	return database.ConnectionConfig{
		Type:     cc.Type,
		Host:     cc.Host,
		Port:     cc.Port,
		Username: cc.Username,
		Password: cc.Password,
		Database: cc.Database,
		Schema:   cc.Schema,
		DSN:      cc.DSN,
		Include:  cc.Include,
		Exclude:  cc.Exclude,
	}
}

func schemaName(cc config.ConnectionConfig) string {
	if cc.Schema != "" {
		return cc.Schema
	}
	if cc.Database != "" {
		return cc.Database
	}
	return cc.Username
}
