-- Insert Dashboard Widgets
INSERT INTO dashboard_widgets (widget_key, widget_name, widget_name_en, is_enabled, allowed_roles, display_order)
VALUES 
    ('financial_summary', 'الملخص المالي', 'Financial Summary', true, ARRAY['admin', 'finance', 'auditor'], 1),
    ('banks', 'الحسابات المصرفية', 'Bank Accounts', true, ARRAY['admin', 'finance', 'auditor'], 2),
    ('cash_accounts', 'الخزائن النقدية', 'Cash Accounts', true, ARRAY['admin', 'finance', 'auditor'], 3),
    ('projects', 'المشاريع', 'Projects', true, ARRAY['admin', 'finance', 'auditor'], 4),
    ('recent_income', 'آخر الإيرادات', 'Recent Income', true, ARRAY['admin', 'finance', 'auditor'], 5),
    ('recent_expenses', 'آخر المصروفات', 'Recent Expenses', true, ARRAY['admin', 'finance', 'auditor'], 6)
ON CONFLICT (widget_key) DO NOTHING;

SELECT 'Dashboard widgets inserted successfully!' as status;
