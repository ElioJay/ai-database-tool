package sqlplan

import "testing"

func TestClassifySQL(t *testing.T) {
	tests := []struct {
		name string
		sql  string
		want StatementKind
		risk RiskLevel
	}{
		{name: "select", sql: "select * from users", want: StatementSelect, risk: RiskLow},
		{name: "with select", sql: "WITH recent AS (SELECT * FROM orders) SELECT * FROM recent", want: StatementSelect, risk: RiskLow},
		{name: "insert", sql: "insert into users(id) values(1)", want: StatementDML, risk: RiskMedium},
		{name: "update", sql: " update users set name='a' where id=1", want: StatementDML, risk: RiskMedium},
		{name: "delete without where", sql: "delete from users", want: StatementDML, risk: RiskHigh},
		{name: "drop", sql: "drop table users", want: StatementDDL, risk: RiskCritical},
		{name: "multiple", sql: "select 1; delete from users", want: StatementOther, risk: RiskCritical},
		{name: "empty", sql: "   ", want: StatementOther, risk: RiskHigh},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Classify(tt.sql)
			if got.Kind != tt.want || got.Risk != tt.risk {
				t.Fatalf("Classify() = kind %s risk %s, want %s/%s", got.Kind, got.Risk, tt.want, tt.risk)
			}
		})
	}
}

func TestConfirmationPolicy(t *testing.T) {
	selectPolicy := ConfirmPolicy(Classification{Kind: StatementSelect, Risk: RiskLow})
	if selectPolicy.PhraseRequired {
		t.Fatalf("low risk select should not require phrase")
	}
	if selectPolicy.ActionLabel != "执行查询" {
		t.Fatalf("select action label = %q", selectPolicy.ActionLabel)
	}

	dropPolicy := ConfirmPolicy(Classification{Kind: StatementDDL, Risk: RiskCritical})
	if !dropPolicy.PhraseRequired {
		t.Fatalf("critical ddl should require phrase")
	}
	if dropPolicy.Phrase == "" {
		t.Fatalf("critical ddl phrase is empty")
	}
}
