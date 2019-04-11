//go:generate stringer -type=Enum

// Package privilegecode contains, well, privilege codes
package privilegecode

// Enum contains privilege codes and must be in sync with tprivilegekind table (../../sql/privilege.sql)
type Enum int

const (
	// Login ...
	Login Enum = iota + 1
	// ManageAccess = Manage access ...
	ManageAccess
	// EditLanguageAttributes = Edit language attributes
	EditLanguageAttributes
	// AcceptOrDeclineChangeRequests = Accept/decline change requests
	AcceptOrDeclineChangeRequests
)
