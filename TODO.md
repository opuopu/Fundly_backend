# Fundly API Implementation Roadmap & Checklist

> **Total Endpoints:** 46  
> **Completed:** 46 / 46 (100%)  
> **Progress:** `[████████████████████████] 100%`  
> **Specification Reference:** [`docs/rest-api-design.md`](docs/rest-api-design.md)

---

## 📊 Progress by Module

| Module | Endpoints | Completed | Status |
| :--- | :---: | :---: | :--- |
| **1. Authentication** | 9 | 9 / 9 | 🟢 **Completed (100%)** |
| **2. User Profile** | 4 | 4 / 4 | 🟢 **Completed (100%)** |
| **3. Accounts** | 6 | 6 / 6 | 🟢 **Completed (100%)** |
| **4. Categories** | 4 | 4 / 4 | 🟢 **Completed (100%)** |
| **5. Transactions** | 7 | 7 / 7 | 🟢 **Completed (100%)** |
| **6. Budgets** | 5 | 5 / 5 | 🟢 **Completed (100%)** |
| **7. Analytics & Stats** | 2 | 2 / 2 | 🟢 **Completed (100%)** |
| **8. Reports & Comparison** | 1 | 1 / 1 | 🟢 **Completed (100%)** |
| **9. Attachments** | 3 | 3 / 3 | 🟢 **Completed (100%)** |
| **10. Configuration** | 2 | 2 / 2 | 🟢 **Completed (100%)** |
| **11. Feedback** | 1 | 1 / 1 | 🟢 **Completed (100%)** |
| **12. Terms & Conditions** | 2 | 2 / 2 | 🟢 **Completed (100%)** |

---

## 1. Authentication Domain (`/api/v1/auth`)

- [x] **#01** `POST /api/v1/auth/register` — Register new user and trigger OTP dispatch  
  *Implemented in:* [`internal/application/auth/register.go`](internal/application/auth/register.go), [`internal/infrastructure/http/auth_handler.go`](internal/infrastructure/http/auth_handler.go)  
  *Tests:* [`internal/application/auth/register_test.go`](internal/application/auth/register_test.go), [`internal/infrastructure/http/auth_handler_test.go`](internal/infrastructure/http/auth_handler_test.go)

- [x] **#02** `POST /api/v1/auth/verify-otp` — Verify 6-digit registration OTP & issue JWT tokens  
  *Implemented in:* [`internal/application/auth/verify_otp.go`](internal/application/auth/verify_otp.go), [`internal/infrastructure/security/jwt.go`](internal/infrastructure/security/jwt.go)  
  *Tests:* [`internal/application/auth/verify_otp_test.go`](internal/application/auth/verify_otp_test.go)

- [x] **#03** `POST /api/v1/auth/resend-otp` — Resend verification OTP code  
  *Implemented in:* [`internal/application/auth/resend_otp.go`](internal/application/auth/resend_otp.go)  
  *Tests:* [`internal/application/auth/resend_otp_test.go`](internal/application/auth/resend_otp_test.go)

- [x] **#04** `POST /api/v1/auth/login` — Email & password login  
  *Implemented in:* [`internal/application/auth/login.go`](internal/application/auth/login.go), [`internal/infrastructure/http/auth_handler.go`](internal/infrastructure/http/auth_handler.go)  
  *Tests:* [`internal/application/auth/login_test.go`](internal/application/auth/login_test.go), [`internal/infrastructure/http/auth_handler_test.go`](internal/infrastructure/http/auth_handler_test.go)

- [x] **#05** `POST /api/v1/auth/social-login` — OAuth login (Google, Apple, Facebook)  
  *Implemented in:* [`internal/application/auth/social_login.go`](internal/application/auth/social_login.go), [`internal/infrastructure/security/social.go`](internal/infrastructure/security/social.go), [`internal/infrastructure/http/auth_handler.go`](internal/infrastructure/http/auth_handler.go)  
  *Tests:* [`internal/application/auth/social_login_test.go`](internal/application/auth/social_login_test.go), [`internal/infrastructure/security/social_test.go`](internal/infrastructure/security/social_test.go), [`internal/infrastructure/http/auth_handler_test.go`](internal/infrastructure/http/auth_handler_test.go)

- [x] **#06** `POST /api/v1/auth/refresh` — Exchange refresh token for new access token  
  *Implemented in:* [`internal/application/auth/refresh_token.go`](internal/application/auth/refresh_token.go), [`internal/infrastructure/security/jwt.go`](internal/infrastructure/security/jwt.go), [`internal/infrastructure/http/auth_handler.go`](internal/infrastructure/http/auth_handler.go)  
  *Tests:* [`internal/application/auth/refresh_token_test.go`](internal/application/auth/refresh_token_test.go), [`internal/infrastructure/security/jwt_test.go`](internal/infrastructure/security/jwt_test.go), [`internal/infrastructure/http/auth_handler_test.go`](internal/infrastructure/http/auth_handler_test.go)

- [x] **#07** `POST /api/v1/auth/forgot-password` — Request password reset email / code  
  *Implemented in:* [`internal/application/auth/forgot_password.go`](internal/application/auth/forgot_password.go), [`internal/infrastructure/http/auth_handler.go`](internal/infrastructure/http/auth_handler.go)  
  *Tests:* [`internal/application/auth/forgot_password_test.go`](internal/application/auth/forgot_password_test.go), [`internal/infrastructure/http/auth_handler_test.go`](internal/infrastructure/http/auth_handler_test.go)

- [x] **#08** `POST /api/v1/auth/reset-password` — Confirm password reset with token  
  *Implemented in:* [`internal/application/auth/reset_password.go`](internal/application/auth/reset_password.go), [`internal/infrastructure/http/auth_handler.go`](internal/infrastructure/http/auth_handler.go)  
  *Tests:* [`internal/application/auth/reset_password_test.go`](internal/application/auth/reset_password_test.go), [`internal/infrastructure/http/auth_handler_test.go`](internal/infrastructure/http/auth_handler_test.go)

- [x] **#09** `POST /api/v1/auth/logout` — Revoke active session / refresh token  
  *Implemented in:* [`internal/application/auth/logout.go`](internal/application/auth/logout.go), [`internal/infrastructure/security/blacklist.go`](internal/infrastructure/security/blacklist.go), [`internal/infrastructure/http/auth_handler.go`](internal/infrastructure/http/auth_handler.go)  
  *Tests:* [`internal/application/auth/logout_test.go`](internal/application/auth/logout_test.go), [`internal/infrastructure/http/auth_handler_test.go`](internal/infrastructure/http/auth_handler_test.go)

---

## 2. User Profile Domain (`/api/v1/users`)

- [x] **#10** `GET /api/v1/users/me` — Fetch authenticated user profile  
  *Implemented in:* [`internal/application/user/get_profile.go`](internal/application/user/get_profile.go), [`internal/infrastructure/http/user_handler.go`](internal/infrastructure/http/user_handler.go)  
  *Tests:* [`internal/application/user/get_profile_test.go`](internal/application/user/get_profile_test.go), [`internal/infrastructure/http/user_handler_test.go`](internal/infrastructure/http/user_handler_test.go)

- [x] **#11** `PATCH /api/v1/users/me` — Update name and email address  
  *Implemented in:* [`internal/application/user/update_profile.go`](internal/application/user/update_profile.go), [`internal/infrastructure/http/user_handler.go`](internal/infrastructure/http/user_handler.go)  
  *Tests:* [`internal/application/user/update_profile_test.go`](internal/application/user/update_profile_test.go), [`internal/infrastructure/http/user_handler_test.go`](internal/infrastructure/http/user_handler_test.go)

- [x] **#12** `PUT /api/v1/users/me/password` — Change password (requires current password)  
  *Implemented in:* [`internal/application/user/change_password.go`](internal/application/user/change_password.go), [`internal/infrastructure/http/user_handler.go`](internal/infrastructure/http/user_handler.go)  
  *Tests:* [`internal/application/user/change_password_test.go`](internal/application/user/change_password_test.go), [`internal/infrastructure/http/user_handler_test.go`](internal/infrastructure/http/user_handler_test.go)

- [x] **#13** `POST /api/v1/users/me/avatar` — Upload and update user profile avatar  
  *Implemented in:* [`internal/application/user/upload_avatar.go`](internal/application/user/upload_avatar.go), [`internal/infrastructure/storage/local_storage.go`](internal/infrastructure/storage/local_storage.go), [`internal/infrastructure/http/user_handler.go`](internal/infrastructure/http/user_handler.go)  
  *Tests:* [`internal/application/user/upload_avatar_test.go`](internal/application/user/upload_avatar_test.go), [`internal/infrastructure/storage/local_storage_test.go`](internal/infrastructure/storage/local_storage_test.go), [`internal/infrastructure/http/user_handler_test.go`](internal/infrastructure/http/user_handler_test.go)

---

## 3. Accounts Domain (`/api/v1/accounts`)

- [x] **#14** `GET /api/v1/accounts` — List all user accounts with current balances  
  *Implemented in:* [`internal/application/account/list_accounts.go`](internal/application/account/list_accounts.go), [`internal/infrastructure/http/account_handler.go`](internal/infrastructure/http/account_handler.go)  
  *Tests:* [`internal/application/account/list_accounts_test.go`](internal/application/account/list_accounts_test.go), [`internal/infrastructure/http/account_handler_test.go`](internal/infrastructure/http/account_handler_test.go)

- [x] **#15** `POST /api/v1/accounts` — Create a new financial account (Cash, Card, Savings, etc.)  
  *Implemented in:* [`internal/application/account/create_account.go`](internal/application/account/create_account.go), [`internal/infrastructure/http/account_handler.go`](internal/infrastructure/http/account_handler.go)  
  *Tests:* [`internal/application/account/create_account_test.go`](internal/application/account/create_account_test.go), [`internal/infrastructure/http/account_handler_test.go`](internal/infrastructure/http/account_handler_test.go)

- [x] **#16** `GET /api/v1/accounts/:id` — Get account details by ID  
  *Implemented in:* [`internal/application/account/get_account.go`](internal/application/account/get_account.go), [`internal/infrastructure/http/account_handler.go`](internal/infrastructure/http/account_handler.go)  
  *Tests:* [`internal/application/account/get_account_test.go`](internal/application/account/get_account_test.go), [`internal/infrastructure/http/account_handler_test.go`](internal/infrastructure/http/account_handler_test.go)

- [x] **#17** `PATCH /api/v1/accounts/:id` — Update account details (name, group, description)  
  *Implemented in:* [`internal/application/account/update_account.go`](internal/application/account/update_account.go), [`internal/infrastructure/http/account_handler.go`](internal/infrastructure/http/account_handler.go)  
  *Tests:* [`internal/application/account/update_account_test.go`](internal/application/account/update_account_test.go), [`internal/infrastructure/http/account_handler_test.go`](internal/infrastructure/http/account_handler_test.go)

- [x] **#18** `DELETE /api/v1/accounts/:id` — Delete / archive an account  
  *Implemented in:* [`internal/application/account/delete_account.go`](internal/application/account/delete_account.go), [`internal/infrastructure/http/account_handler.go`](internal/infrastructure/http/account_handler.go)  
  *Tests:* [`internal/application/account/delete_account_test.go`](internal/application/account/delete_account_test.go), [`internal/infrastructure/http/account_handler_test.go`](internal/infrastructure/http/account_handler_test.go)

- [x] **#19** `GET /api/v1/accounts/summary` — Get aggregated balance totals (Cash, Cards, Net Balance)  
  *Implemented in:* [`internal/application/account/get_summary.go`](internal/application/account/get_summary.go), [`internal/infrastructure/http/account_handler.go`](internal/infrastructure/http/account_handler.go)  
  *Tests:* [`internal/application/account/get_summary_test.go`](internal/application/account/get_summary_test.go), [`internal/infrastructure/http/account_handler_test.go`](internal/infrastructure/http/account_handler_test.go)

---

## 4. Categories Domain (`/api/v1/categories`)

- [x] **#20** `GET /api/v1/categories` — List transaction categories (Income / Expense)  
  *Implemented in:* [`internal/application/category/list_categories.go`](internal/application/category/list_categories.go), [`internal/infrastructure/http/category_handler.go`](internal/infrastructure/http/category_handler.go)  
  *Tests:* [`internal/application/category/list_categories_test.go`](internal/application/category/list_categories_test.go), [`internal/infrastructure/http/category_handler_test.go`](internal/infrastructure/http/category_handler_test.go)

- [x] **#21** `POST /api/v1/categories` — Create custom category (name, type, icon, color)  
  *Implemented in:* [`internal/application/category/create_category.go`](internal/application/category/create_category.go), [`internal/infrastructure/http/category_handler.go`](internal/infrastructure/http/category_handler.go)  
  *Tests:* [`internal/application/category/create_category_test.go`](internal/application/category/create_category_test.go), [`internal/infrastructure/http/category_handler_test.go`](internal/infrastructure/http/category_handler_test.go)

- [x] **#22** `PATCH /api/v1/categories/:id` — Update category details  
  *Implemented in:* [`internal/application/category/update_category.go`](internal/application/category/update_category.go), [`internal/infrastructure/http/category_handler.go`](internal/infrastructure/http/category_handler.go)  
  *Tests:* [`internal/application/category/update_category_test.go`](internal/application/category/update_category_test.go), [`internal/infrastructure/http/category_handler_test.go`](internal/infrastructure/http/category_handler_test.go)

- [x] **#23** `DELETE /api/v1/categories/:id` — Delete custom category  
  *Implemented in:* [`internal/application/category/delete_category.go`](internal/application/category/delete_category.go), [`internal/infrastructure/http/category_handler.go`](internal/infrastructure/http/category_handler.go)  
  *Tests:* [`internal/application/category/delete_category_test.go`](internal/application/category/delete_category_test.go), [`internal/infrastructure/http/category_handler_test.go`](internal/infrastructure/http/category_handler_test.go)

---

## 5. Transactions Domain (`/api/v1/transactions`)

- [x] **#24** `GET /api/v1/transactions` — Query transactions with filters (`type`, `date`, `category_id`, `account_id`, pagination)
  *Implemented in:* [`internal/application/transaction/list_transactions.go`](internal/application/transaction/list_transactions.go), [`internal/infrastructure/http/transaction_handler.go`](internal/infrastructure/http/transaction_handler.go)

- [x] **#25** `POST /api/v1/transactions` — Create income/expense transaction entry (applies account balance effect)
  *Implemented in:* [`internal/application/transaction/create_transaction.go`](internal/application/transaction/create_transaction.go), [`internal/domain/transaction/transaction.go`](internal/domain/transaction/transaction.go)

- [x] **#26** `GET /api/v1/transactions/:id` — Get transaction details
  *Implemented in:* [`internal/application/transaction/get_transaction.go`](internal/application/transaction/get_transaction.go)

- [x] **#27** `PATCH /api/v1/transactions/:id` — Update transaction entry (reconciles balances)
  *Implemented in:* [`internal/application/transaction/update_transaction.go`](internal/application/transaction/update_transaction.go)

- [x] **#28** `DELETE /api/v1/transactions/:id` — Delete transaction entry (reverses balance)
  *Implemented in:* [`internal/application/transaction/delete_transaction.go`](internal/application/transaction/delete_transaction.go)

- [x] **#29** `GET /api/v1/transactions/search` — Quick keyword search across entries
  *Implemented in:* [`internal/application/transaction/search_transactions.go`](internal/application/transaction/search_transactions.go)

- [x] **#30** `POST /api/v1/transactions/transfer` — Execute money transfer between accounts
  *Implemented in:* [`internal/application/transaction/transfer_transaction.go`](internal/application/transaction/transfer_transaction.go), reuses [`internal/application/transfer/transfer.go`](internal/application/transfer/transfer.go)

---

## 6. Budgets Domain (`/api/v1/budgets`)

- [x] **#31** `GET /api/v1/budgets` — List monthly budgets with limit, spent, and usage percentage
  *Implemented in:* [`internal/application/budget/list_budgets.go`](internal/application/budget/list_budgets.go), [`internal/infrastructure/http/budget_handler.go`](internal/infrastructure/http/budget_handler.go)

- [x] **#32** `POST /api/v1/budgets` — Create budget allocation (category, month, limit)
  *Implemented in:* [`internal/application/budget/create_budget.go`](internal/application/budget/create_budget.go), [`internal/domain/budget/budget.go`](internal/domain/budget/budget.go)

- [x] **#33** `GET /api/v1/budgets/:id` — Get budget details by ID (real-time spent/remaining/usage)
  *Implemented in:* [`internal/application/budget/get_budget.go`](internal/application/budget/get_budget.go)

- [x] **#34** `PATCH /api/v1/budgets/:id` — Update budget limit and metadata
  *Implemented in:* [`internal/application/budget/update_budget.go`](internal/application/budget/update_budget.go)

- [x] **#35** `DELETE /api/v1/budgets/:id` — Delete budget allocation
  *Implemented in:* [`internal/application/budget/delete_budget.go`](internal/application/budget/delete_budget.go)

  *Spent/remaining/usage computed from EXPENSE transactions via* [`postgres.TransactionRepository.SpentByCategoryMonth`](internal/infrastructure/persistence/postgres/transaction_repo.go)

---

## 7. Analytics & Stats Domain (`/api/v1/stats`)

- [x] **#36** `GET /api/v1/stats/breakdown` — Category percentage breakdown for donut chart (`income` / `expense`, month)
  *Implemented in:* [`internal/application/stat/breakdown.go`](internal/application/stat/breakdown.go), [`internal/infrastructure/http/stat_handler.go`](internal/infrastructure/http/stat_handler.go)

- [x] **#37** `GET /api/v1/stats/overview` — High-level monthly overview (income, expense, net, budget utilization)
  *Implemented in:* [`internal/application/stat/overview.go`](internal/application/stat/overview.go)

  *Aggregations in* [`postgres.StatRepository`](internal/infrastructure/persistence/postgres/stat_repo.go) *(domain model/interface in* [`internal/domain/stat/stat.go`](internal/domain/stat/stat.go)*)*

---

## 8. Reports & Comparison Domain (`/api/v1/reports`)

- [x] **#38** `POST /api/v1/reports/compare` — Compare two periods/months across selected categories
  *Implemented in:* [`internal/application/report/compare.go`](internal/application/report/compare.go), [`internal/infrastructure/http/report_handler.go`](internal/infrastructure/http/report_handler.go) *(reuses the stats aggregation)*

---

## 9. Attachments Domain (`/api/v1/attachments`)

- [x] **#39** `POST /api/v1/attachments` — Upload receipt or proof image (`multipart/form-data`)
  *Implemented in:* [`internal/application/attachment/upload_attachment.go`](internal/application/attachment/upload_attachment.go), [`internal/infrastructure/http/attachment_handler.go`](internal/infrastructure/http/attachment_handler.go)

- [x] **#40** `GET /api/v1/attachments/:id` — View attachment metadata
  *Implemented in:* [`internal/application/attachment/get_attachment.go`](internal/application/attachment/get_attachment.go)

- [x] **#41** `DELETE /api/v1/attachments/:id` — Remove attachment
  *Implemented in:* [`internal/application/attachment/delete_attachment.go`](internal/application/attachment/delete_attachment.go)

  *Added `user_id` to `attachments` via* [`migrations/000002_attachments_user_id.up.sql`](migrations/000002_attachments_user_id.up.sql) *for ownership. Transactions now link `attachment_ids` on create and expose `attachment_count` (lists) / `attachments[]` (detail).*

---

## 10. Configuration Domain (`/api/v1/configuration`)

- [x] **#42** `GET /api/v1/configuration` — Retrieve user preferences and default currency (falls back to defaults)
  *Implemented in:* [`internal/application/config/config.go`](internal/application/config/config.go), [`internal/infrastructure/http/config_handler.go`](internal/infrastructure/http/config_handler.go)

- [x] **#43** `PUT /api/v1/configuration` — Update default currency and application settings (upsert)
  *Implemented in:* [`internal/application/config/config.go`](internal/application/config/config.go), [`internal/domain/config/config.go`](internal/domain/config/config.go)

---

## 11. Feedback Domain (`/api/v1/feedback`)

- [x] **#44** `POST /api/v1/feedback` — Submit user feedback and issue report
  *Implemented in:* [`internal/application/feedback/submit_feedback.go`](internal/application/feedback/submit_feedback.go), [`internal/infrastructure/http/feedback_handler.go`](internal/infrastructure/http/feedback_handler.go)

---

## 12. Terms & Conditions Domain (`/api/v1/terms`)

- [x] **#45** `GET /api/v1/terms/latest` — Retrieve active Terms & Conditions with clauses (public)
  *Implemented in:* [`internal/application/terms/get_latest.go`](internal/application/terms/get_latest.go), [`internal/infrastructure/http/terms_handler.go`](internal/infrastructure/http/terms_handler.go)

- [x] **#46** `POST /api/v1/terms/acknowledge` — Record user acceptance of Terms version
  *Implemented in:* [`internal/application/terms/acknowledge.go`](internal/application/terms/acknowledge.go), [`internal/infrastructure/persistence/postgres/terms_repo.go`](internal/infrastructure/persistence/postgres/terms_repo.go)

---

## 🩺 System Routes (Completed)

- [x] **#00** `GET /health` — Real-time application & PostgreSQL health probe  
  *Implemented in:* [`internal/infrastructure/http/handler.go`](internal/infrastructure/http/handler.go)
- [x] **#00** `GET /api/v1/health` — Versioned health check endpoint  
  *Implemented in:* [`internal/infrastructure/http/handler.go`](internal/infrastructure/http/handler.go)
