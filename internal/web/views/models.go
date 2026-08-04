// Package views owns purpose-built presentation models and server-rendered HTML.
// It intentionally has no dependency on domain or persistence packages.
package views

type Variant string

const (
	VariantNeutral     Variant = "neutral"
	VariantInfo        Variant = "info"
	VariantComplete    Variant = "complete"
	VariantWarning     Variant = "warning"
	VariantCritical    Variant = "critical"
	VariantSpecial     Variant = "special"
	VariantProvisional Variant = "provisional"
)

type Icon struct {
	Name  string
	Label string
}

type Link struct {
	Label      string
	Href       string
	Variant    string
	Icon       Icon
	External   bool
	Current    bool
	BadgeLabel string
}

type Button struct {
	ID             string
	Label          string
	Variant        string
	Type           string
	Icon           Icon
	Disabled       bool
	DisabledReason string
	Loading        bool
	FullWidth      bool
}

type Badge struct {
	Label   string
	Variant Variant
}

type Alert struct {
	Title    string
	Message  string
	Variant  Variant
	Blocking bool
	Icon     Icon
	Action   *Link
}

type Field struct {
	ID           string
	Name         string
	Label        string
	Type         string
	Value        string
	Hint         string
	Error        string
	Autocomplete string
	InputMode    string
	Required     bool
	Disabled     bool
}

type Card struct {
	Title       string
	Description string
	Badge       *Badge
	Action      *Link
}

type TableColumn struct {
	Label   string
	Numeric bool
}

type TableCell struct {
	Text    string
	Numeric bool
	Badge   *Badge
}

type TableRow struct {
	Label string
	Cells []TableCell
}

type ResponsiveTable struct {
	Caption string
	Columns []TableColumn
	Rows    []TableRow
}

type Dialog struct {
	ID          string
	Title       string
	Description string
	CancelLabel string
	Action      Button
	Open        bool
}

type Pagination struct {
	CurrentPage int
	TotalPages  int
	ResultCount int
	PreviousURL string
	NextURL     string
}

type EmptyState struct {
	Title       string
	Description string
	Icon        Icon
	Action      *Link
}

type StaffingIndicator struct {
	Status       Badge
	AdultsFilled int
	AdultsTarget int
	ScoutsFilled int
	ScoutsTarget int
}

type ShiftCard struct {
	Name      string
	Date      string
	Time      string
	Location  string
	State     string
	Special   bool
	Staffing  StaffingIndicator
	Action    Link
	Signup    Badge
	Lifecycle Badge
}

type AgreementStatus struct {
	Person      string
	Detail      string
	Status      Badge
	ConfirmedAt string
	Action      *Link
}

type PersonOption struct {
	ID          string
	Name        string
	Detail      string
	Status      Badge
	Selected    bool
	Disabled    bool
	Explanation string
}

type AttendanceRow struct {
	Initials   string
	Person     string
	Role       string
	Origin     string
	Status     Badge
	RawEvents  string
	Adjustment string
	Action     *Button
}

type Announcement struct {
	Title       string
	Priority    Badge
	Author      string
	PublishedAt string
	ReadState   string
}

type DeliveryChannel struct {
	Name   string
	Detail string
	Status Badge
}

type ScoutBucksSummary struct {
	State          string
	Status         Badge
	CreditedHours  string
	Distributable  string
	EffectiveRate  string
	Rounding       string
	Reconciliation string
	Revision       string
}

type Home struct {
	PageTitle    string
	Brand        string
	Headline     string
	Supporting   string
	Navigation   []Link
	CSRFToken    string
	SmokeMessage string
	SmokeInput   string
}

type BootstrapStage string

const (
	BootstrapStageEntry   BootstrapStage = "entry"
	BootstrapStageEnroll  BootstrapStage = "enroll"
	BootstrapStagePasskey BootstrapStage = "passkey"
)

type BootstrapPage struct {
	PageTitle            string
	Brand                string
	Stage                BootstrapStage
	Navigation           []Link
	CSRFToken            string
	Token                string
	Email                string
	FirstName            string
	LastName             string
	PreferredDisplayName string
	Alert                *Alert
	Fields               []Field
}

type SignInPage struct {
	PageTitle  string
	Brand      string
	Navigation []Link
	CSRFToken  string
	EmailHint  string
	Alert      *Alert
}

type LandingPage struct {
	PageTitle   string
	Brand       string
	Heading     string
	Supporting  string
	DisplayName string
	Navigation  []Link
}

type AccountPage struct {
	PageTitle    string
	Brand        string
	DisplayName  string
	PrimaryEmail string
	Navigation   []Link
}

type Gallery struct {
	PageTitle      string
	Season         string
	Context        string
	Actor          string
	Navigation     []Link
	Buttons        []Button
	Badges         []Badge
	Alerts         []Alert
	Fields         []Field
	Cards          []Card
	Table          ResponsiveTable
	Dialog         Dialog
	Pagination     Pagination
	Empty          EmptyState
	Shifts         []ShiftCard
	Agreements     []AgreementStatus
	People         []PersonOption
	Attendance     []AttendanceRow
	Announcement   Announcement
	Delivery       []DeliveryChannel
	ScoutBucks     []ScoutBucksSummary
	ParityMessage  string
	ParityFragment bool
}
