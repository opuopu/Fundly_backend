# 📮 Fundly Postman Collection & Environment

This folder contains the official Postman collection and environment configurations for the **Fundly** Personal Finance and Wealth Management REST API.

---

## 📁 Files in This Directory

| File | Description |
| :--- | :--- |
| [`Fundly_API.postman_collection.json`](./Fundly_API.postman_collection.json) | Complete collection containing **48 requests** across **13 domains**, with automated tests and token persistence scripts. |
| [`Fundly_Local.postman_environment.json`](./Fundly_Local.postman_environment.json) | Local environment configuration pointing to `http://localhost:8081` with default test credentials. |
| [`Fundly_Production.postman_environment.json`](./Fundly_Production.postman_environment.json) | Production/staging environment template. |

---

## ⚡ Quick Start

### 1. Import into Postman
1. Open **Postman**.
2. Click the **Import** button (top left).
3. Drag and drop both:
   - `Fundly_API.postman_collection.json`
   - `Fundly_Local.postman_environment.json`
4. In the top-right environment dropdown, select **Fundly - Local**.

### 2. Start the Server
Make sure the backend server is running:
```bash
# Start PostgreSQL (if using Docker)
docker compose up -d

# Start the Go server (runs on :8081 by default)
go run cmd/fundly/main.go
```

### 3. Verify Connectivity
Run the first request in the collection:
- `00. Health & Diagnostics` ➔ `Health Check (API v1)`
- Expected response: `200 OK` with `"status": "healthy"`.

---

## 🔐 Automated Authentication & Token Handling

You do **not** need to manually copy and paste JWT tokens!

1. When you execute **`Verify OTP`**, **`Login`**, **`Social Login`**, or **`Refresh Token`**, Postman's built-in test script automatically captures:
   - `tokens.access_token` ➔ saves to environment variable `{{accessToken}}`
   - `tokens.refresh_token` ➔ saves to environment variable `{{refreshToken}}`
   - `user.id` ➔ saves to environment variable `{{userId}}`
2. The collection root is pre-configured with **Bearer Token** authentication using `{{accessToken}}`. All protected endpoints automatically inherit this token.
3. When you create an **Account**, **Category**, **Transaction**, **Budget**, or **Attachment**, the response script automatically stores their respective IDs (`{{accountId}}`, `{{categoryId}}`, `{{transactionId}}`, `{{budgetId}}`, `{{attachmentId}}`) for subsequent calls.

---

## 🧭 Recommended Testing Workflow

Follow this logical sequence to test all features:

```text
1. Register & Verify OTP (or Login)
   ├── POST /api/v1/auth/register
   └── POST /api/v1/auth/verify-otp (or POST /api/v1/auth/login)
          └── Auto-populates {{accessToken}} & {{refreshToken}}

2. Profile & Settings
   ├── GET /api/v1/users/me
   └── GET /api/v1/configuration

3. Accounts Setup
   ├── POST /api/v1/accounts (Create Checking & Savings)
   ├── GET /api/v1/accounts (Lists accounts and sets {{fromAccountId}} / {{toAccountId}})
   └── GET /api/v1/accounts/summary

4. Categories
   ├── GET /api/v1/categories (Loads default seed categories like Food, Salary, etc.)
   └── POST /api/v1/categories (Create custom category)

5. Transactions & Transfers
   ├── POST /api/v1/transactions (Create expense or income)
   ├── GET /api/v1/transactions (Filter by date, account, category)
   ├── GET /api/v1/transactions/search (Search by keyword)
   └── POST /api/v1/transactions/transfer (Atomic transfer between 2 accounts)

6. Budgets & Progress
   ├── POST /api/v1/budgets (Set monthly category budget)
   └── GET /api/v1/budgets (View limit vs spent progress and usage percentage)

7. Analytics & Reporting
   ├── GET /api/v1/stats/breakdown (Category spending breakdown donut chart data)
   ├── GET /api/v1/stats/overview (Total income, expense, net savings, budget utilization)
   └── POST /api/v1/reports/compare (Compare spending between two months)

8. Attachments & Feedback
   ├── POST /api/v1/attachments (Upload receipt image)
   ├── POST /api/v1/feedback (Submit feedback)
   └── POST /api/v1/terms/acknowledge (Acknowledge Terms & Conditions v2.1)
```

---

## 🏷️ Collection Structure (48 Requests in 13 Folders)

```
Fundly API
├── 00. Health & Diagnostics (2 requests)
│   ├── Health Check (Root)             [GET /health]
│   └── Health Check (API v1)           [GET /api/v1/health]
├── 01. Authentication (9 requests)
│   ├── Register                        [POST /api/v1/auth/register]
│   ├── Verify OTP                      [POST /api/v1/auth/verify-otp]
│   ├── Resend OTP                      [POST /api/v1/auth/resend-otp]
│   ├── Login                           [POST /api/v1/auth/login]
│   ├── Social Login                    [POST /api/v1/auth/social-login]
│   ├── Refresh Token                   [POST /api/v1/auth/refresh]
│   ├── Forgot Password                 [POST /api/v1/auth/forgot-password]
│   ├── Reset Password                  [POST /api/v1/auth/reset-password]
│   └── Logout                          [POST /api/v1/auth/logout]
├── 02. Users (4 requests)
│   ├── Get Profile                     [GET /api/v1/users/me]
│   ├── Update Profile                  [PATCH /api/v1/users/me]
│   ├── Change Password                 [PUT /api/v1/users/me/password]
│   └── Upload Avatar                   [POST /api/v1/users/me/avatar]
├── 03. Accounts (6 requests)
│   ├── List Accounts                   [GET /api/v1/accounts]
│   ├── Create Account                  [POST /api/v1/accounts]
│   ├── Get Accounts Summary            [GET /api/v1/accounts/summary]
│   ├── Get Account by ID               [GET /api/v1/accounts/:id]
│   ├── Update Account                  [PATCH /api/v1/accounts/:id]
│   └── Delete Account                  [DELETE /api/v1/accounts/:id]
├── 04. Categories (4 requests)
│   ├── List Categories                 [GET /api/v1/categories]
│   ├── Create Category                 [POST /api/v1/categories]
│   ├── Update Category                 [PATCH /api/v1/categories/:id]
│   └── Delete Category                 [DELETE /api/v1/categories/:id]
├── 05. Transactions (7 requests)
│   ├── List Transactions               [GET /api/v1/transactions]
│   ├── Create Transaction              [POST /api/v1/transactions]
│   ├── Get Transaction by ID           [GET /api/v1/transactions/:id]
│   ├── Update Transaction              [PATCH /api/v1/transactions/:id]
│   ├── Delete Transaction              [DELETE /api/v1/transactions/:id]
│   ├── Search Transactions             [GET /api/v1/transactions/search]
│   └── Transfer Between Accounts       [POST /api/v1/transactions/transfer]
├── 06. Budgets (5 requests)
│   ├── List Budgets                    [GET /api/v1/budgets]
│   ├── Create Budget                   [POST /api/v1/budgets]
│   ├── Get Budget by ID                [GET /api/v1/budgets/:id]
│   ├── Update Budget                   [PATCH /api/v1/budgets/:id]
│   └── Delete Budget                   [DELETE /api/v1/budgets/:id]
├── 07. Stats & Analytics (2 requests)
│   ├── Category Breakdown (Donut Chart)[GET /api/v1/stats/breakdown]
│   └── Monthly Overview Metrics        [GET /api/v1/stats/overview]
├── 08. Reports & Comparisons (1 request)
│   └── Compare Periods                 [POST /api/v1/reports/compare]
├── 09. Attachments (3 requests)
│   ├── Upload Attachment               [POST /api/v1/attachments]
│   ├── Get Attachment Metadata         [GET /api/v1/attachments/:id]
│   └── Delete Attachment               [DELETE /api/v1/attachments/:id]
├── 10. User Configuration (2 requests)
│   ├── Get Configuration               [GET /api/v1/configuration]
│   └── Update Configuration            [PUT /api/v1/configuration]
├── 11. Feedback (1 request)
│   └── Submit Feedback                 [POST /api/v1/feedback]
└── 12. Terms & Conditions (2 requests)
    ├── Get Latest Terms (Public)       [GET /api/v1/terms/latest]
    └── Acknowledge Terms               [POST /api/v1/terms/acknowledge]
```

---

## 💰 Monetary Values Convention

To guarantee precision and prevent floating-point calculation discrepancies:
* All financial amounts (`balance`, `initial_balance`, `amount`, `limit_amount`) are represented as **integers in minor units (cents)**.
* Example: `$10.50` ➔ `1050`, `$200.00` ➔ `20000`, `$1,500.00` ➔ `150000`.

---

## 📋 Environment Variables Reference

| Variable | Description | Default / Example |
| :--- | :--- | :--- |
| `baseUrl` | Backend API host URL | `http://localhost:8081` |
| `accessToken` | JWT Access Token for protected routes | *(Auto-populated on login/verify)* |
| `refreshToken` | JWT Refresh Token for session renewal | *(Auto-populated on login/verify)* |
| `userId` | User ID of the authenticated user | *(Auto-populated on login/register)* |
| `testEmail` | Test email for auth flows | `user@example.com` |
| `testPassword` | Test password for auth flows | `Password123!` |
| `accountId` | Target Account ID for account/transaction tests | *(Auto-populated on account create/list)* |
| `fromAccountId` | Source account for transfers | *(Auto-populated on list accounts)* |
| `toAccountId` | Destination account for transfers | *(Auto-populated on list accounts)* |
| `categoryId` | Target Category ID for transaction/budget tests | `cat_food` *(or auto-populated)* |
| `transactionId` | Target Transaction ID for single transaction tests | *(Auto-populated on transaction create/list)* |
| `budgetId` | Target Budget ID for budget tests | *(Auto-populated on budget create/list)* |
| `attachmentId` | Target Attachment ID for attachment tests | *(Auto-populated on upload)* |
