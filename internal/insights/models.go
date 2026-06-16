package insights

type Summary struct {
	IncomeTotal         float64           `json:"income_total"`
	ExpenseTotal        float64           `json:"expense_total"`
	Profit              float64           `json:"profit"`
	PendingCollection   float64           `json:"pending_collection"`
	ActiveOrders        int               `json:"active_orders"`
	PaidCompletedOrders int               `json:"paid_completed_orders"`
	OverdueOrders       int               `json:"overdue_orders"`
	SalesByPlatform     []PlatformSales   `json:"sales_by_platform"`
	TopExpenseMerchants []MerchantExpense `json:"top_expense_merchants"`
}

type PlatformSales struct {
	Platform string  `json:"platform"`
	Count    int     `json:"count"`
	Total    float64 `json:"total"`
}

type MerchantExpense struct {
	ComercioID int     `json:"comercio_id"`
	Name       string  `json:"name"`
	Total      float64 `json:"total"`
}

type SummaryFilter struct {
	From string
	To   string
}
