package query

// параметры, к-рые нужны для выполнения запроса
type languageProposalsListQueryParams struct {
	Sduserid int64 // 0 для незарег. польз.
	Commonid int64
}

type languageProposalsListQueryHeader struct {
	Commonid     int64
	Languageid   int64
	Languageslug string
}

type languageProposalsListQueryRecord struct {
	Commonid       int64
	Proposalid     int64
	Senseid        int64
	Proposalstatus string
	Phrase         string
	Word           string
	Phantom        bool
	OwnerId        int64
	Sdusernickname string
	Languageslug   string
	Iscommon       bool
	Ismine         bool
}

// Параметры шаблона
type languageProposalsListFormTemplateParamsType struct {
	P              *languageProposalsListQueryParams
	Header         *languageProposalsListQueryHeader
	Records        []*languageProposalsListQueryRecord
	IsLoggedIn     bool
	LoggedInUserId int64
}
