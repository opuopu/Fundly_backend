# Fundly API Implementation Roadmap & Checklist

> **Total Endpoints:** 46  
> **Completed:** 19 / 46 (41.3%)  
> **Progress:** `[█████████░░░░░░░░░░░] 41.3%`  
> **Specification Reference:** [`docs/rest-api-design.md`](docs/rest-api-design.md)

---

## 📊 Progress by Module

| Module | Endpoints | Completed | Status |
| :--- | :---: | :---: | :--- |
| **1. Authentication** | 9 | 9 / 9 | 🟢 **Completed (100%)** |
| **2. User Profile** | 4 | 4 / 4 | 🟢 **Completed (100%)** |
| **3. Accounts** | 6 | 6 / 6 | 🟢 **Completed (100%)** |
| **4. Categories** | 4 | 0 / 4 | 🟡 In Progress (#20 next) |
| **5. Transactions** | 7 | 0 / 7 | ⚪ Pending |
| **6. Budgets** | 5 | 0 / 5 | ⚪ Pending |
| **7. Analytics & Stats** | 2 | 0 / 2 | ⚪ Pending |
| **8. Reports & Comparison** | 1 | 0 / 1 | ⚪ Pending |
| **9. Attachments** | 3 | 0 / 3 | ⚪ Pending |
| **10. Configuration** | 2 | 0 / 2 | ⚪ Pending |
| **11. Feedback** | 1 | 0 / 1 | ⚪ Pending |
| **12. Terms & Conditions** | 2 | 0 / 2 | ⚪ Pending |

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

- [ ] **#20** `GET /api/v1/categories` — List transaction categories (Income / Expense)
- [ ] **#21** `POST /api/v1/categories` — Create custom category (name, type, icon, color)
- [ ] **#22** `PATCH /api/v1/categories/:id` — Update category details
- [ ] **#23** `DELETE /api/v1/categories/:id` — Delete custom category

---

## 5. Transactions Domain (`/api/v1/transactions`)

- [ ] **#24** `GET /api/v1/transactions` — Query transactions with filters (`type`, `date`, `category_id`, `account_id`, pagination)
- [ ] **#25** `POST /api/v1/transactions` — Create income/expense transaction entry
- [ ] **#26** `GET /api/v1/transactions/:id` — Get transaction details and attachments
- [ ] **#27** `PATCH /api/v1/transactions/:id` — Update transaction entry
- [ ] **#28** `DELETE /api/v1/transactions/:id` — Delete transaction entry
- [ ] **#29** `GET /api/v1/transactions/search` — Quick keyword / group search across entries
- [ ] **#30** `POST /api/v1/transactions/transfer` — Execute money transfer between accounts

---

## 6. Budgets Domain (`/api/v1/budgets`)

- [ ] **#31** `GET /api/v1/budgets` — List monthly budgets with limit, spent, and usage percentage
- [ ] **#32** `POST /api/v1/budgets` — Create budget allocation (category, month, limit)
- [ ] **#33** `GET /api/v1/budgets/:id` — Get budget details by ID
- [ ] **#34** `PATCH /api/v1/budgets/:id` — Update budget limit and metadata
- [ ] **#35** `DELETE /api/v1/budgets/:id` — Delete budget allocation

---

## 7. Analytics & Stats Domain (`/api/v1/stats`)

- [ ] **#36** `GET /api/v1/stats/breakdown` — Category percentage breakdown for donut chart (`income` / `expense`, month)
- [ ] **#37** `GET /api/v1/stats/overview` — High-level monthly overview (total income, expense, savings)

---

## 8. Reports & Comparison Domain (`/api/v1/reports`)

- [ ] **#38** `POST /api/v1/reports/compare` — Compare two periods/months across selected categories

---

## 9. Attachments Domain (`/api/v1/attachments`)

- [ ] **#39** `POST /api/v1/attachments` — Upload receipt or proof image (`multipart/form-data`)
- [ ] **#40** `GET /api/v1/attachments/:id` — Download / view attachment metadata
- [ ] **#41** `DELETE /api/v1/attachments/:id` — Remove attachment

---

## 10. Configuration Domain (`/api/v1/configuration`)

- [ ] **#42** `GET /api/v1/configuration` — Retrieve user preferences and default currency
- [ ] **#43** `PUT /api/v1/configuration` — Update default currency and application settings

---

## 11. Feedback Domain (`/api/v1/feedback`)

- [ ] **#44** `POST /api/v1/feedback` — Submit user feedback and issue report

---

## 12. Terms & Conditions Domain (`/api/v1/terms`)

- [ ] **#45** `GET /api/v1/terms/latest` — Retrieve active Terms & Conditions with clauses
- [ ] **#46** `POST /api/v1/terms/acknowledge` — Record user acceptance of Terms version

---

## 🩺 System Routes (Completed)

- [x] **#00** `GET /health` — Real-time application & PostgreSQL health probe  
  *Implemented in:* [`internal/infrastructure/http/handler.go`](internal/infrastructure/http/handler.go)
- [x] **#00** `GET /api/v1/health` — Versioned health check endpoint  
  *Implemented in:* [`internal/infrastructure/http/handler.go`](internal/infrastructure/http/handler.go)
