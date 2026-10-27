ALTER TABLE budgets ADD COLUMN IF NOT EXISTS date DATE;
UPDATE budgets SET date = (month || '-01')::date WHERE date IS NULL;
ALTER TABLE budgets ALTER COLUMN date SET NOT NULL;
