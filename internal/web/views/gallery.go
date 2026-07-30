package views

// GalleryData returns fictional, non-sensitive presentation fixtures. The data
// documents component APIs; it is not application state or business behavior.
func GalleryData() Gallery {
	return Gallery{
		PageTitle: "Signal design system",
		Season:    "2026 Tree Lot Season",
		Context:   "Week 2 · Dec 1–7 · Pacific Time",
		Actor:     "Morgan Rivera · Committee",
		Navigation: []Link{
			{Label: "Foundations", Href: "#foundations", Icon: Icon{Name: "signal"}, Current: true},
			{Label: "Primitives", Href: "#primitives", Icon: Icon{Name: "components"}},
			{Label: "Domain patterns", Href: "#domain-patterns", Icon: Icon{Name: "calendar"}},
			{Label: "Response parity", Href: "#response-parity", Icon: Icon{Name: "swap"}},
		},
		Buttons: []Button{
			{Label: "Primary action", Variant: "primary", Type: "button", Icon: Icon{Name: "check"}},
			{Label: "Secondary", Variant: "secondary", Type: "button"},
			{Label: "Quiet", Variant: "quiet", Type: "button"},
			{Label: "Destructive", Variant: "destructive", Type: "button", Icon: Icon{Name: "warning"}},
			{Label: "Saving assignment", Variant: "primary", Type: "button", Loading: true},
			{ID: "signup-disabled", Label: "Sign Up Alex", Variant: "secondary", Type: "button", Disabled: true, DisabledReason: "Shift is full."},
		},
		Badges: []Badge{
			{Label: "Neutral", Variant: VariantNeutral},
			{Label: "Informational", Variant: VariantInfo},
			{Label: "Complete", Variant: VariantComplete},
			{Label: "Needs attention", Variant: VariantWarning},
			{Label: "Critical", Variant: VariantCritical},
			{Label: "Special Event", Variant: VariantSpecial},
			{Label: "Provisional", Variant: VariantProvisional},
		},
		Alerts: []Alert{
			{Title: "Schedule updated", Message: "Two shift times changed this week.", Variant: VariantInfo, Icon: Icon{Name: "info"}},
			{Title: "Alex is signed up", Message: "Saturday, Dec 6 · 9:00 AM–1:00 PM", Variant: VariantComplete, Icon: Icon{Name: "check"}},
			{Title: "Agreement needed", Message: "Jordan cannot sign up until confirmed.", Variant: VariantWarning, Icon: Icon{Name: "warning"}},
			{Title: "Shift critically understaffed", Message: "One adult and three scouts are still needed.", Variant: VariantCritical, Blocking: true, Icon: Icon{Name: "critical"}},
		},
		Fields: []Field{
			{ID: "display-name", Name: "display-name", Label: "Display name", Type: "text", Value: "Morgan Rivera", Hint: "Shown to people in your household.", Autocomplete: "name", Required: true},
			{ID: "mobile-phone", Name: "mobile-phone", Label: "Mobile phone", Type: "tel", Value: "(925) 555-0182", Hint: "Formatting does not determine account identity.", Autocomplete: "tel", InputMode: "tel"},
			{ID: "agreement-url", Name: "agreement-url", Label: "Agreement URL", Type: "url", Value: "example.com/rules", Error: "Enter an approved Google Docs URL beginning with https://.", Required: true},
			{ID: "season-name", Name: "season-name", Label: "Season name", Type: "text", Value: "2026 Tree Lot", Hint: "Disabled after season archive begins.", Disabled: true},
		},
		Cards: []Card{
			{Title: "Default card", Description: "A calm container for related content."},
			{Title: "Action card", Description: "One stable concept with a clear next action.", Badge: &Badge{Label: "Complete", Variant: VariantComplete}, Action: &Link{Label: "View details", Href: "#domain-patterns"}},
		},
		Table: ResponsiveTable{
			Caption: "Recent assignments · 3 results",
			Columns: []TableColumn{{Label: "Person"}, {Label: "Shift"}, {Label: "Status"}, {Label: "Hours", Numeric: true}},
			Rows: []TableRow{
				{Label: "Alex Rivera assignment", Cells: []TableCell{{Text: "Alex Rivera"}, {Text: "Sat, Dec 6 · Morning"}, {Badge: &Badge{Label: "Checked Out", Variant: VariantComplete}}, {Text: "4.00", Numeric: true}}},
				{Label: "Morgan Rivera assignment", Cells: []TableCell{{Text: "Morgan Rivera"}, {Text: "Sun, Dec 7 · Close"}, {Badge: &Badge{Label: "Signed Up", Variant: VariantInfo}}, {Text: "—", Numeric: true}}},
			},
		},
		Dialog: Dialog{
			ID:          "cancel-dialog",
			Title:       "Cancel Alex's Assignment?",
			Description: "Alex will be removed from Saturday Morning Sales. The open Scout slot will be available to other households.",
			CancelLabel: "Keep assignment",
			Action:      Button{Label: "Cancel Alex's Assignment", Variant: "primary", Type: "submit"},
			Open:        true,
		},
		Pagination: Pagination{CurrentPage: 1, TotalPages: 3, ResultCount: 24, NextURL: "?page=2"},
		Empty: EmptyState{
			Title:       "No draft shifts",
			Description: "This is expected after the schedule is published.",
			Icon:        Icon{Name: "calendar"},
			Action:      &Link{Label: "Create shift", Href: "#domain-patterns"},
		},
		Shifts: []ShiftCard{
			{Name: "Morning Sales", Date: "Sat, Dec 6", Time: "9 AM–1 PM", Location: "North Tree Lot", State: "available", Staffing: StaffingIndicator{Status: Badge{Label: "FULL", Variant: VariantComplete}, AdultsFilled: 2, AdultsTarget: 2, ScoutsFilled: 4, ScoutsTarget: 4}, Action: Link{Label: "View", Href: "#domain-patterns"}, Signup: Badge{Label: "Available", Variant: VariantNeutral}, Lifecycle: Badge{Label: "Published", Variant: VariantInfo}},
			{Name: "Afternoon Sales", Date: "Sat, Dec 6", Time: "1–5 PM", Location: "North Tree Lot", State: "low", Staffing: StaffingIndicator{Status: Badge{Label: "LOW", Variant: VariantWarning}, AdultsFilled: 1, AdultsTarget: 2, ScoutsFilled: 2, ScoutsTarget: 4}, Action: Link{Label: "Sign Up", Href: "#domain-patterns"}, Signup: Badge{Label: "Available", Variant: VariantInfo}, Lifecycle: Badge{Label: "Published", Variant: VariantInfo}},
			{Name: "Closing Crew", Date: "Sun, Dec 7", Time: "5–9 PM", Location: "North Tree Lot", State: "critical", Staffing: StaffingIndicator{Status: Badge{Label: "CRITICAL", Variant: VariantCritical}, AdultsFilled: 0, AdultsTarget: 2, ScoutsFilled: 1, ScoutsTarget: 4}, Action: Link{Label: "Help this shift", Href: "#domain-patterns"}, Signup: Badge{Label: "Available", Variant: VariantWarning}, Lifecycle: Badge{Label: "Published", Variant: VariantInfo}},
			{Name: "Community Kickoff", Date: "Sun, Dec 7", Time: "10 AM–2 PM", Location: "North Tree Lot", State: "special", Special: true, Staffing: StaffingIndicator{Status: Badge{Label: "OK", Variant: VariantComplete}, AdultsFilled: 4, AdultsTarget: 5, ScoutsFilled: 9, ScoutsTarget: 12}, Action: Link{Label: "View assignment", Href: "#domain-patterns"}, Signup: Badge{Label: "Signed Up", Variant: VariantInfo}, Lifecycle: Badge{Label: "Special Event", Variant: VariantSpecial}},
		},
		Agreements: []AgreementStatus{
			{Person: "Morgan Rivera", Detail: "Authenticated parent", Status: Badge{Label: "Confirmed", Variant: VariantComplete}, ConfirmedAt: "Nov 18, 2026"},
			{Person: "Alex Rivera", Detail: "Managed Scout · participation blocked", Status: Badge{Label: "Not Confirmed", Variant: VariantWarning}, Action: &Link{Label: "Confirm Agreement", Href: "#domain-patterns"}},
		},
		People: []PersonOption{
			{ID: "person-dana", Name: "Dana Rivera", Detail: "Adult · eligible for 1 remaining adult slot", Status: Badge{Label: "Eligible", Variant: VariantComplete}, Selected: true},
			{ID: "person-alex", Name: "Alex Rivera", Detail: "Scout · eligible for 1 remaining scout slot", Status: Badge{Label: "Eligible", Variant: VariantComplete}},
			{ID: "person-jordan", Name: "Jordan Rivera", Detail: "Scout", Status: Badge{Label: "Ineligible", Variant: VariantWarning}, Disabled: true, Explanation: "Seasonal Agreement Not Confirmed."},
			{ID: "person-morgan", Name: "Morgan Rivera", Detail: "Adult", Status: Badge{Label: "Assigned", Variant: VariantNeutral}, Disabled: true, Explanation: "Already assigned to this shift."},
		},
		Attendance: []AttendanceRow{
			{Initials: "JK", Person: "Jordan Kim", Role: "Adult", Origin: "Scheduled · Kim household", Status: Badge{Label: "Pending", Variant: VariantInfo}, RawEvents: "Check-in open", Action: &Button{Label: "Check In Jordan", Variant: "primary", Type: "button", FullWidth: true}},
			{Initials: "AR", Person: "Alex Rivera", Role: "Scout", Origin: "Scheduled", Status: Badge{Label: "Checked In", Variant: VariantComplete}, RawEvents: "5:03 PM by Morgan Rivera · no checkout yet"},
			{Initials: "SP", Person: "Sam Patel", Role: "Adult", Origin: "Scheduled", Status: Badge{Label: "Adjusted", Variant: VariantWarning}, RawEvents: "Raw events 4:58–5:31 PM", Adjustment: "Approved: 4.00 hours · reason retained separately"},
		},
		Announcement: Announcement{Title: "Lot closing early tonight", Priority: Badge{Label: "Urgent priority", Variant: VariantCritical}, Author: "Casey Nguyen · Committee", PublishedAt: "Dec 5 at 3:42 PM", ReadState: "Unread"},
		Delivery: []DeliveryChannel{
			{Name: "Web", Detail: "Canonical publication", Status: Badge{Label: "Published", Variant: VariantComplete}},
			{Name: "SMS", Detail: "43 delivered · 2 failed", Status: Badge{Label: "Needs retry", Variant: VariantWarning}},
			{Name: "Groups.io", Detail: "Optional channel", Status: Badge{Label: "Disabled", Variant: VariantNeutral}},
		},
		ScoutBucks: []ScoutBucksSummary{
			{State: "provisional", Status: Badge{Label: "Provisional", Variant: VariantProvisional}, CreditedHours: "1,248.75", Revision: "Before Treasurer finalization"},
			{State: "finalized", Status: Badge{Label: "Finalized", Variant: VariantComplete}, CreditedHours: "1,248.75", Distributable: "$12,500.00", EffectiveRate: "$10.0090", Rounding: "$0.03", Reconciliation: "$0.00 unallocated", Revision: "Finalized revision 3"},
		},
	}
}
