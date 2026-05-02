package sqlplan

type StatementKind string

const (
	StatementSelect StatementKind = "SELECT"
	StatementDML    StatementKind = "DML"
	StatementDDL    StatementKind = "DDL"
	StatementOther  StatementKind = "OTHER"
)

type RiskLevel string

const (
	RiskLow      RiskLevel = "low"
	RiskMedium   RiskLevel = "medium"
	RiskHigh     RiskLevel = "high"
	RiskCritical RiskLevel = "critical"
)

type Classification struct {
	Kind       StatementKind
	Risk       RiskLevel
	Reason     string
	Multi      bool
	Normalized string
}

type Plan struct {
	Explanation string         `json:"explanation"`
	SQL         string         `json:"sql"`
	AIKind      StatementKind  `json:"statement_type"`
	AIRisk      RiskLevel      `json:"risk"`
	NeedConfirm bool           `json:"need_confirm"`
	Local       Classification `json:"-"`
}

type Policy struct {
	ActionLabel    string
	PhraseRequired bool
	Phrase         string
}
