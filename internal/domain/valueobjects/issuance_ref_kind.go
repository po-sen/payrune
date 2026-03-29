package valueobjects

import "strings"

type IssuanceRefKind string

const (
	IssuanceRefKindHDPathAbsolute IssuanceRefKind = "hd_path_absolute"
	IssuanceRefKindCreate2Salt    IssuanceRefKind = "create2_salt"
)

var issuanceRefKinds = map[string]IssuanceRefKind{
	"hd_path_absolute": IssuanceRefKindHDPathAbsolute,
	"create2_salt":     IssuanceRefKindCreate2Salt,
}

func ParseIssuanceRefKind(raw string) (IssuanceRefKind, bool) {
	normalized := strings.ToLower(strings.TrimSpace(raw))
	if normalized == "" {
		return "", false
	}

	kind, ok := issuanceRefKinds[normalized]
	return kind, ok
}

func (k IssuanceRefKind) IsZero() bool {
	return strings.TrimSpace(string(k)) == ""
}
