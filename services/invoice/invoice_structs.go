package invoice

type Invoice struct {
	Method string `json:"method"`
	Data   []struct {
		InvoiceID           string      `json:"invoice_id"`
		SysUserid           string      `json:"sys_userid"`
		SysGroupid          string      `json:"sys_groupid"`
		SysPermUser         string      `json:"sys_perm_user"`
		SysPermGroup        string      `json:"sys_perm_group"`
		SysPermOther        string      `json:"sys_perm_other"`
		Idhash              string      `json:"idhash"`
		InvoiceType         string      `json:"invoice_type"`
		ReminderStep        string      `json:"reminder_step"`
		InvoiceCompanyID    string      `json:"invoice_company_id"`
		ClientID            string      `json:"client_id"`
		InvoiceOrderID      string      `json:"invoice_order_id"`
		InvoiceNumber       string      `json:"invoice_number"`
		InvoiceDate         string      `json:"invoice_date"`
		PaymentDate         string      `json:"payment_date"`
		CompanyName         string      `json:"company_name"`
		Gender              string      `json:"gender"`
		ContactName         string      `json:"contact_name"`
		Street              string      `json:"street"`
		Zip                 string      `json:"zip"`
		City                string      `json:"city"`
		State               string      `json:"state"`
		Country             string      `json:"country"`
		Email               string      `json:"email"`
		VatID               string      `json:"vat_id"`
		PaymentTerms        string      `json:"payment_terms"`
		PaymentGateway      string      `json:"payment_gateway"`
		StatusPrinted       string      `json:"status_printed"`
		StatusSent          string      `json:"status_sent"`
		StatusPaid          string      `json:"status_paid"`
		StatusReminded      string      `json:"status_reminded"`
		StatusRefunded      string      `json:"status_refunded"`
		StatusIrrecoverable string      `json:"status_irrecoverable"`
		InvoiceAmount       string      `json:"invoice_amount"`
		Notes               string      `json:"notes"`
		Annotation          interface{} `json:"annotation"`
		MandateReference    interface{} `json:"mandate_reference"`
		CollectionDate      string      `json:"collection_date"`
		SequenceType        string      `json:"sequence_type"`
	} `json:"data"`
}
