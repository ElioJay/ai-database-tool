package sqlplan

import "testing"

func TestParseAIResponseAcceptsJSONAndOverridesClassification(t *testing.T) {
	raw := `{
	  "explanation": "删除用户表",
	  "sql": "drop table users",
	  "statement_type": "SELECT",
	  "risk": "low",
	  "need_confirm": false
	}`

	plan, err := ParseAIResponse(raw)
	if err != nil {
		t.Fatalf("ParseAIResponse() error = %v", err)
	}
	if plan.SQL != "drop table users" {
		t.Fatalf("SQL = %q", plan.SQL)
	}
	if plan.Local.Kind != StatementDDL || plan.Local.Risk != RiskCritical {
		t.Fatalf("local classification = %s/%s", plan.Local.Kind, plan.Local.Risk)
	}
	if !plan.NeedConfirm {
		t.Fatalf("dangerous SQL should need confirm regardless of AI response")
	}
}

func TestParseAIResponseAcceptsFencedJSON(t *testing.T) {
	raw := "```json\n{\"explanation\":\"查用户\",\"sql\":\"select * from users\",\"statement_type\":\"SELECT\",\"risk\":\"low\",\"need_confirm\":true}\n```"
	plan, err := ParseAIResponse(raw)
	if err != nil {
		t.Fatalf("ParseAIResponse() error = %v", err)
	}
	if plan.Local.Kind != StatementSelect {
		t.Fatalf("kind = %s", plan.Local.Kind)
	}
}

func TestParseAIResponseRejectsMissingSQL(t *testing.T) {
	if _, err := ParseAIResponse(`{"explanation":"x"}`); err == nil {
		t.Fatalf("expected missing SQL error")
	}
}
