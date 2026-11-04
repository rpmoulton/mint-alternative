-- +goose Up
-- +goose StatementBegin

CREATE TABLE institutions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    display_name TEXT NOT NULL,
    logo_url TEXT,
    website TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TRIGGER set_institution_updated_at BEFORE UPDATE ON institutions
FOR EACH ROW EXECUTE PROCEDURE set_updated_at();

CREATE TABLE accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    institution_id UUID REFERENCES institutions(id) ON DELETE SET NULL,
    name TEXT NOT NULL,
    account_type TEXT NOT NULL, -- depository, credit, loan, investment, other_asset, other_liability
    account_subtype TEXT NOT NULL DEFAULT '',
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    current_balance NUMERIC(18, 4) NOT NULL DEFAULT 0,
    available_balance NUMERIC(18, 4),
    credit_limit NUMERIC(18, 4),
    interest_rate NUMERIC(7, 4),
    source_type TEXT NOT NULL DEFAULT 'manual', -- manual, connected
    archived_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_accounts_user_id ON accounts(user_id);
CREATE INDEX idx_accounts_institution_id ON accounts(institution_id);

CREATE TRIGGER set_account_updated_at BEFORE UPDATE ON accounts
FOR EACH ROW EXECUTE PROCEDURE set_updated_at();

CREATE TABLE categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    parent_category_id UUID REFERENCES categories(id) ON DELETE SET NULL,
    name TEXT NOT NULL,
    type TEXT NOT NULL, -- income, expense, transfer
    icon TEXT,
    color TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_categories_user_id ON categories(user_id);
CREATE INDEX idx_categories_parent_id ON categories(parent_category_id);

CREATE TRIGGER set_category_updated_at BEFORE UPDATE ON categories
FOR EACH ROW EXECUTE PROCEDURE set_updated_at();

CREATE TABLE transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    category_id UUID REFERENCES categories(id) ON DELETE SET NULL,
    amount NUMERIC(18, 4) NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    transaction_date DATE NOT NULL,
    posted_date DATE,
    merchant_name TEXT,
    description TEXT NOT NULL,
    notes TEXT,
    transaction_type TEXT NOT NULL, -- debit, credit
    status TEXT NOT NULL DEFAULT 'posted', -- pending, posted
    source_type TEXT NOT NULL DEFAULT 'manual', -- manual, imported
    external_transaction_id TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_transactions_user_date ON transactions(user_id, transaction_date DESC);
CREATE INDEX idx_transactions_account ON transactions(account_id);
CREATE INDEX idx_transactions_category ON transactions(category_id);

CREATE TRIGGER set_transaction_updated_at BEFORE UPDATE ON transactions
FOR EACH ROW EXECUTE PROCEDURE set_updated_at();

CREATE TABLE budgets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    month SMALLINT NOT NULL CHECK (month BETWEEN 1 AND 12),
    year SMALLINT NOT NULL CHECK (year >= 2000),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (user_id, year, month)
);

CREATE INDEX idx_budgets_user_year_month ON budgets(user_id, year, month);

CREATE TRIGGER set_budget_updated_at BEFORE UPDATE ON budgets
FOR EACH ROW EXECUTE PROCEDURE set_updated_at();

CREATE TABLE budget_categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    budget_id UUID NOT NULL REFERENCES budgets(id) ON DELETE CASCADE,
    category_id UUID NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
    allocated_amount NUMERIC(18, 4) NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (budget_id, category_id)
);

CREATE INDEX idx_budget_categories_budget ON budget_categories(budget_id);

CREATE TRIGGER set_budget_category_updated_at BEFORE UPDATE ON budget_categories
FOR EACH ROW EXECUTE PROCEDURE set_updated_at();

-- Seed system-provided default categories (user_id IS NULL)
INSERT INTO categories (id, user_id, parent_category_id, name, type, icon, color) VALUES
    ('00000000-0000-0000-0000-000000000001', NULL, NULL, 'Salary', 'income', 'Briefcase', '#10B981'),
    ('00000000-0000-0000-0000-000000000002', NULL, NULL, 'Other Income', 'income', 'DollarSign', '#059669'),
    ('00000000-0000-0000-0000-000000000003', NULL, NULL, 'Transfers', 'transfer', 'ArrowLeftRight', '#6366F1'),
    ('00000000-0000-0000-0000-000000000004', NULL, NULL, 'Groceries', 'expense', 'ShoppingCart', '#F59E0B'),
    ('00000000-0000-0000-0000-000000000005', NULL, NULL, 'Restaurants', 'expense', 'Utensils', '#F97316'),
    ('00000000-0000-0000-0000-000000000006', NULL, NULL, 'Transportation', 'expense', 'Car', '#3B82F6'),
    ('00000000-0000-0000-0000-000000000007', NULL, NULL, 'Housing', 'expense', 'Home', '#8B5CF6'),
    ('00000000-0000-0000-0000-000000000008', NULL, NULL, 'Utilities', 'expense', 'Zap', '#EC4899'),
    ('00000000-0000-0000-0000-000000000009', NULL, NULL, 'Shopping', 'expense', 'ShoppingBag', '#EF4444'),
    ('00000000-0000-0000-0000-00000000000a', NULL, NULL, 'Entertainment', 'expense', 'Film', '#14B8A6'),
    ('00000000-0000-0000-0000-00000000000b', NULL, NULL, 'Healthcare', 'expense', 'HeartPulse', '#06B6D4'),
    ('00000000-0000-0000-0000-00000000000c', NULL, NULL, 'Education', 'expense', 'GraduationCap', '#64748B'),
    ('00000000-0000-0000-0000-00000000000d', NULL, NULL, 'Uncategorized', 'expense', 'HelpCircle', '#94A3B8');

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS budget_categories;
DROP TABLE IF EXISTS budgets;
DROP TABLE IF EXISTS transactions;
DROP TABLE IF EXISTS categories;
DROP TABLE IF EXISTS accounts;
DROP TABLE IF EXISTS institutions;
-- +goose StatementEnd
