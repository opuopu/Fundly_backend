# Fundly REST API Design Specification

**Document Version:** 1.0.0  
**Target Base URL:** `/api/v1`  
**Protocol:** HTTPS / JSON  
**Authentication:** Bearer JWT (Access Token + Refresh Token)

---

## 📋 Table of Contents
1. [Overview & Architectural Context](#1-overview--architectural-context)
2. [Design-to-API Traceability Matrix](#2-design-to-api-traceability-matrix)
3. [Global Conventions & Response Envelopes](#3-global-conventions--response-envelopes)
4. [API Summary (45 Endpoints)](#4-api-summary-45-endpoints)
5. [Detailed Endpoint Specifications](#5-detailed-endpoint-specifications)
   - [5.1 Authentication & User Domain](#51-authentication--user-domain)
   - [5.2 Accounts Domain](#52-accounts-domain)
   - [5.3 Categories Domain](#53-categories-domain)
   - [5.4 Transactions Domain](#54-transactions-domain)
   - [5.5 Budgets Domain](#55-budgets-domain)
   - [5.6 Stats & Analytics Domain](#56-stats--analytics-domain)
   - [5.7 Comparison & Reporting Domain](#57-comparison--reporting-domain)
   - [5.8 Attachments Domain](#58-attachments-domain)
   - [5.9 User Configuration Domain](#59-user-configuration-domain)
   - [5.10 Feedback Domain](#510-feedback-domain)
   - [5.11 Terms & Conditions Domain](#511-terms--conditions-domain)
6. [Mapping to DDD Go Architecture](#6-mapping-to-ddd-go-architecture)

---

## 1. Overview & Architectural Context

This document outlines the REST API design for **Fundly**, synthesized from the mobile application screens provided in the `/design` directory. 

The API layer acts as the **Driving / Primary Adapter** in the Clean / Domain-Driven Design (DDD) architecture. It receives HTTP requests, maps JSON payloads to domain entities and application commands/queries, delegates business logic to Application Services, and formats the output into consistent JSON responses.

---

## 2. Design-to-API Traceability Matrix

Every screen in the design is backed by specific REST endpoints:

| Screen / UI Feature | Screenshot Reference | Associated Domain | Required APIs |
| :--- | :--- | :--- | :--- |
| **Registration** | `Screenshot 11-07-10.png` | User / Auth | `POST /api/v1/auth/register` |
| **OTP Verification** | `11-07-24, 11-07-35, 11-07-45` | User / Auth | `POST /api/v1/auth/verify-otp`, `POST /api/v1/auth/resend-otp` |
| **Login & Social Auth** | `Screenshot 11-07-53.png` | User / Auth | `POST /api/v1/auth/login`, `POST /api/v1/auth/social-login`, `POST /api/v1/auth/forgot-password` |
| **Home / Daily Transactions** | `Screenshot 11-10-54.png` | Transaction / Account | `GET /api/v1/transactions`, `GET /api/v1/accounts/summary`, `GET /api/v1/categories` |
| **Daily Transaction Details** | `Screenshot 11-11-04.png` | Transaction | `GET /api/v1/transactions/:id`, `PATCH /api/v1/transactions/:id`, `DELETE /api/v1/transactions/:id` |
| **Budget Allocations List** | `Screenshot 11-11-12.png` | Budget | `GET /api/v1/budgets` |
| **Budget Details** | `Screenshot 11-11-21.png` | Budget | `GET /api/v1/budgets/:id`, `PATCH /api/v1/budgets/:id`, `DELETE /api/v1/budgets/:id` |
| **Expense History** | `Screenshot 11-11-31.png` | Transaction | `GET /api/v1/transactions?type=EXPENSE` |
| **Expenditure Details** | `Screenshot 11-11-43.png` | Transaction | `GET /api/v1/transactions/:id`, `PATCH /api/v1/transactions/:id`, `DELETE /api/v1/transactions/:id` |
| **Income History** | `Screenshot 11-11-53.png` | Transaction | `GET /api/v1/transactions?type=INCOME` |
| **Income Details** | `Screenshot 11-12-01.png` | Transaction | `GET /api/v1/transactions/:id`, `PATCH /api/v1/transactions/:id`, `DELETE /api/v1/transactions/:id` |
| **Add Entry (Income/Expense)** | `Screenshots 11-12-16, 11-12-33`| Transaction | `POST /api/v1/transactions` |
| **Add Entry (Budget)** | `Screenshot 11-12-24.png` | Budget | `POST /api/v1/budgets` |
| **Keyword Search** | `Screenshot 11-12-40.png` | Transaction | `GET /api/v1/transactions/search` |
| **Comparison & Filters** | `Screenshots 11-12-50, 11-12-59`| Analytics / Report | `POST /api/v1/reports/compare` |
| **Income Analytics Donut** | `Screenshot 11-13-27.png` | Analytics / Stats | `GET /api/v1/stats/breakdown?type=INCOME` |
| **Expense Analytics Donut** | `Screenshot 11-13-35.png` | Analytics / Stats | `GET /api/v1/stats/breakdown?type=EXPENSE` |
| **Accounts Tab (Cash / Card)**| `Screenshot 11-13-45.png` | Account | `GET /api/v1/accounts`, `GET /api/v1/transactions?account_id=...` |
| **All Accounts List** | `Screenshot 11-13-56.png` | Account | `GET /api/v1/accounts`, `DELETE /api/v1/accounts/:id` |
| **Add Account Modal** | `Screenshot 11-14-07.png` | Account | `POST /api/v1/accounts` |
| **Profile & Settings Menu** | `Screenshot 11-14-17.png` | User / Auth | `GET /api/v1/users/me`, `POST /api/v1/auth/logout` |
| **Edit Profile Details** | `Screenshot 11-14-27.png` | User | `PATCH /api/v1/users/me` |
| **Change Password** | `Screenshot 11-14-37.png` | User / Auth | `PUT /api/v1/users/me/password` |
| **Feedback Submission** | `Screenshot 11-20-25.png` | Feedback | `POST /api/v1/feedback` |
| **Terms & Conditions** | `Screenshot 11-20-37.png` | Terms | `GET /api/v1/terms/latest`, `POST /api/v1/terms/acknowledge` |
| **Configuration** | `Screenshot 11-20-44.png` | Configuration | `GET /api/v1/configuration`, `PUT /api/v1/configuration` |
| **File Attachments** | Multiple detail views | Attachment | `POST /api/v1/attachments`, `DELETE /api/v1/attachments/:id` |

---

## 3. Global Conventions & Response Envelopes

### 3.1 Monetary Representation
* To avoid floating-point inaccuracies, all monetary values are transferred as **integers in minor units (cents)** along with the three-letter ISO 4217 currency code.
* Example: `$200.00 USD` is represented as:
  ```json
  {
    "amount": 20000,
    "currency": "USD"
  }
  ```

### 3.2 Standard Success Envelope
```json
{
  "success": true,
  "data": { ... },
  "message": "Operation completed successfully"
}
```

### 3.3 Standard Paginated Envelope
```json
{
  "success": true,
  "data": [ ... ],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total_records": 142,
    "total_pages": 8
  }
}
```

### 3.4 Standard Error Envelope
```json
{
  "success": false,
  "error": {
    "code": "INVALID_CREDENTIALS",
    "message": "The email or password you entered is incorrect",
    "details": [
      {
        "field": "password",
        "issue": "must be at least 8 characters"
      }
    ]
  }
}
```

---

## 4. API Summary (45 Endpoints)

| # | Group | Method | Path | Summary |
| :- | :--- | :--- | :--- | :--- |
| 1 | Auth | `POST` | `/api/v1/auth/register` | Register new user & trigger OTP |
| 2 | Auth | `POST` | `/api/v1/auth/verify-otp` | Verify 6-digit registration OTP |
| 3 | Auth | `POST` | `/api/v1/auth/resend-otp` | Resend OTP code |
| 4 | Auth | `POST` | `/api/v1/auth/login` | Email/password login |
| 5 | Auth | `POST` | `/api/v1/auth/social-login` | OAuth login (Google/Apple/Facebook) |
| 6 | Auth | `POST` | `/api/v1/auth/refresh` | Exchange refresh token for new access token |
| 7 | Auth | `POST` | `/api/v1/auth/forgot-password` | Request password reset code |
| 8 | Auth | `POST` | `/api/v1/auth/reset-password` | Set new password with reset code |
| 9 | Auth | `POST` | `/api/v1/auth/logout` | Revoke current user session |
| 10 | User | `GET` | `/api/v1/users/me` | Fetch authenticated user profile |
| 11 | User | `PATCH` | `/api/v1/users/me` | Update name and email address |
| 12 | User | `PUT` | `/api/v1/users/me/password` | Change user password |
| 13 | User | `POST` | `/api/v1/users/me/avatar` | Upload and update profile photo |
| 14 | Accounts | `GET` | `/api/v1/accounts` | List all accounts with balances |
| 15 | Accounts | `POST` | `/api/v1/accounts` | Create a new financial account |
| 16 | Accounts | `GET` | `/api/v1/accounts/:id` | Get account details by ID |
| 17 | Accounts | `PATCH` | `/api/v1/accounts/:id` | Update account details |
| 18 | Accounts | `DELETE` | `/api/v1/accounts/:id` | Delete / archive account |
| 19 | Accounts | `GET` | `/api/v1/accounts/summary` | Get aggregated balance totals (Cash, Cards, Net) |
| 20 | Categories | `GET` | `/api/v1/categories` | List transaction categories (Income/Expense) |
| 21 | Categories | `POST` | `/api/v1/categories` | Create custom user category |
| 22 | Categories | `PATCH` | `/api/v1/categories/:id` | Update custom category |
| 23 | Categories | `DELETE` | `/api/v1/categories/:id` | Delete custom category |
| 24 | Transactions | `GET` | `/api/v1/transactions` | Query transactions (filters: date, type, category) |
| 25 | Transactions | `POST` | `/api/v1/transactions` | Create income/expense transaction |
| 26 | Transactions | `GET` | `/api/v1/transactions/:id` | Get transaction details and attachments |
| 27 | Transactions | `PATCH` | `/api/v1/transactions/:id` | Update transaction entry |
| 28 | Transactions | `DELETE` | `/api/v1/transactions/:id` | Delete transaction entry |
| 29 | Transactions | `GET` | `/api/v1/transactions/search` | Quick keyword/group search across entries |
| 30 | Transactions | `POST` | `/api/v1/transactions/transfer` | Execute money transfer between accounts |
| 31 | Budgets | `GET` | `/api/v1/budgets` | List budgets with progress and limits |
| 32 | Budgets | `POST` | `/api/v1/budgets` | Create budget allocation |
| 33 | Budgets | `GET` | `/api/v1/budgets/:id` | Get budget details by ID |
| 34 | Budgets | `PATCH` | `/api/v1/budgets/:id` | Update budget allocation |
| 35 | Budgets | `DELETE` | `/api/v1/budgets/:id` | Delete budget allocation |
| 36 | Analytics | `GET` | `/api/v1/stats/breakdown` | Get category breakdown for donut chart |
| 37 | Analytics | `GET` | `/api/v1/stats/overview` | Get monthly high-level metrics |
| 38 | Reports | `POST` | `/api/v1/reports/compare` | Compare 2 periods/months and category breakdown |
| 39 | Attachments | `POST` | `/api/v1/attachments` | Upload receipt or proof image |
| 40 | Attachments | `GET` | `/api/v1/attachments/:id` | Download / view attachment metadata |
| 41 | Attachments | `DELETE` | `/api/v1/attachments/:id` | Remove attachment |
| 42 | Configuration | `GET` | `/api/v1/configuration` | Get user preferences & default currency |
| 43 | Configuration | `PUT` | `/api/v1/configuration` | Update default currency & settings |
| 44 | Feedback | `POST` | `/api/v1/feedback` | Submit user feedback |
| 45 | Terms | `GET` | `/api/v1/terms/latest` | Retrieve current Terms & Conditions |
| 46 | Terms | `POST` | `/api/v1/terms/acknowledge` | Record user acceptance of Terms version |

---

## 5. Detailed Endpoint Specifications

### 5.1 Authentication & User Domain

#### 1. Register User
* **Endpoint:** `POST /api/v1/auth/register`
* **Auth Required:** No
* **Request Body:**
```json
{
  "email": "user@example.com",
  "password": "Password123!",
  "confirm_password": "Password123!"
}
```
* **Response `201 Created`:**
```json
{
  "success": true,
  "data": {
    "user_id": "usr_94bfa28d",
    "email": "user@example.com",
    "otp_sent": true,
    "otp_expires_in_seconds": 300
  },
  "message": "Verification OTP sent to your email"
}
```

#### 2. Verify OTP
* **Endpoint:** `POST /api/v1/auth/verify-otp`
* **Auth Required:** No
* **Request Body:**
```json
{
  "email": "user@example.com",
  "otp": "128333"
}
```
* **Response `200 OK`:**
```json
{
  "success": true,
  "data": {
    "user": {
      "id": "usr_94bfa28d",
      "email": "user@example.com"
    },
    "tokens": {
      "access_token": "eyJhbGciOiJIUzI1Ni...",
      "refresh_token": "def502000c...",
      "expires_in": 3600
    }
  },
  "message": "OTP Verified"
}
```

#### 3. Resend OTP
* **Endpoint:** `POST /api/v1/auth/resend-otp`
* **Auth Required:** No
* **Request Body:**
```json
{
  "email": "user@example.com"
}
```
* **Response `200 OK`:**
```json
{
  "success": true,
  "message": "New verification code has been dispatched"
}
```

#### 4. Login
* **Endpoint:** `POST /api/v1/auth/login`
* **Auth Required:** No
* **Request Body:**
```json
{
  "email": "user@example.com",
  "password": "Password123!",
  "remember_me": true
}
```
* **Response `200 OK`:**
```json
{
  "success": true,
  "data": {
    "user": {
      "id": "usr_94bfa28d",
      "email": "user@example.com",
      "full_name": "Thom Haye",
      "avatar_url": "https://assets.fundly.com/avatars/usr_1.jpg"
    },
    "tokens": {
      "access_token": "eyJhbGciOiJIUzI1Ni...",
      "refresh_token": "def502000c...",
      "expires_in": 86400
    }
  }
}
```

#### 5. Social Login (OAuth)
* **Endpoint:** `POST /api/v1/auth/social-login`
* **Auth Required:** No
* **Request Body:**
```json
{
  "provider": "google", // "google" | "apple" | "facebook"
  "id_token": "eyJhbGciOiJSUzI1NiIs..."
}
```
* **Response `200 OK`:** (Returns user profile and auth tokens)

#### 6. Refresh Token
* **Endpoint:** `POST /api/v1/auth/refresh`
* **Auth Required:** No
* **Request Body:**
```json
{
  "refresh_token": "def502000c..."
}
```
* **Response `200 OK`:** Returns new `access_token` and `refresh_token`.

#### 7. Forgot Password Request
* **Endpoint:** `POST /api/v1/auth/forgot-password`
* **Request Body:**
```json
{
  "email": "user@example.com"
}
```
* **Response `200 OK`:** Dispatches reset code/token.

#### 8. Reset Password Confirm
* **Endpoint:** `POST /api/v1/auth/reset-password`
* **Request Body:**
```json
{
  "email": "user@example.com",
  "reset_code": "489211",
  "new_password": "NewSecretPassword123!"
}
```
* **Response `200 OK`:** Success message.

#### 9. Logout
* **Endpoint:** `POST /api/v1/auth/logout`
* **Auth Required:** Bearer JWT
* **Request Body:**
```json
{
  "refresh_token": "def502000c..."
}
```
* **Response `200 OK`:** Session invalidated.

#### 10. Get Current User Profile
* **Endpoint:** `GET /api/v1/users/me`
* **Auth Required:** Bearer JWT
* **Response `200 OK`:**
```json
{
  "success": true,
  "data": {
    "id": "usr_94bfa28d",
    "email": "michelle.rivera@example.com",
    "full_name": "Thom Haye",
    "avatar_url": "https://assets.fundly.com/avatars/usr_1.jpg",
    "default_currency": "USD",
    "created_at": "2026-01-15T08:30:00Z"
  }
}
```

#### 11. Update Profile Details
* **Endpoint:** `PATCH /api/v1/users/me`
* **Auth Required:** Bearer JWT
* **Request Body:**
```json
{
  "full_name": "Thom Haye",
  "email": "michelle.rivera@example.com"
}
```
* **Response `200 OK`:** Updated user object.

#### 12. Change Password
* **Endpoint:** `PUT /api/v1/users/me/password`
* **Auth Required:** Bearer JWT
* **Request Body:**
```json
{
  "current_password": "OldPassword123!",
  "new_password": "NewPassword456!",
  "confirm_password": "NewPassword456!"
}
```
* **Response `200 OK`:** Password changed confirmation.

#### 13. Upload User Avatar
* **Endpoint:** `POST /api/v1/users/me/avatar`
* **Auth Required:** Bearer JWT
* **Content-Type:** `multipart/form-data`
* **Form Field:** `avatar` (image file: png, jpg, webp, max 5MB)
* **Response `200 OK`:** Returns new `avatar_url`.

---

### 5.2 Accounts Domain

#### 14. List User Accounts
* **Endpoint:** `GET /api/v1/accounts`
* **Auth Required:** Bearer JWT
* **Query Parameters:** `group` (optional filter: `cash`, `card`, `savings`, `investment`, `loan`, `insurance`)
* **Response `200 OK`:**
```json
{
  "success": true,
  "data": [
    {
      "id": "acc_01",
      "name": "Cash",
      "group": "cash",
      "balance": 20000,
      "currency": "USD",
      "description": "Daily wallet cash",
      "created_at": "2026-08-01T10:00:00Z"
    },
    {
      "id": "acc_02",
      "name": "Debit Card",
      "group": "card",
      "balance": 154050,
      "currency": "USD",
      "description": "Primary checking card",
      "created_at": "2026-08-05T12:00:00Z"
    }
  ]
}
```

#### 15. Create New Account
* **Endpoint:** `POST /api/v1/accounts`
* **Auth Required:** Bearer JWT
* **Request Body:**
```json
{
  "group": "cash", // "cash" | "card" | "savings" | "investment" | "loan" | "insurance" | "checking"
  "name": "Cash Wallet",
  "initial_balance": 20000, // cents ($200.00)
  "currency": "USD",
  "description": "Physical wallet cash"
}
```
* **Response `201 Created`:** Returns created account entity.

#### 16. Get Account Details
* **Endpoint:** `GET /api/v1/accounts/:id`
* **Auth Required:** Bearer JWT
* **Response `200 OK`:** Detailed account view.

#### 17. Update Account
* **Endpoint:** `PATCH /api/v1/accounts/:id`
* **Auth Required:** Bearer JWT
* **Request Body:**
```json
{
  "name": "Primary Cash Wallet",
  "description": "Updated description"
}
```
* **Response `200 OK`:** Updated account object.

#### 18. Delete Account
* **Endpoint:** `DELETE /api/v1/accounts/:id`
* **Auth Required:** Bearer JWT
* **Response `200 OK`:** Account deleted message.

#### 19. Get Accounts Balance Summary
* **Endpoint:** `GET /api/v1/accounts/summary`
* **Auth Required:** Bearer JWT
* **Response `200 OK`:**
```json
{
  "success": true,
  "data": {
    "total_balance": 20000,
    "total_cash": 20000,
    "total_card": 154050,
    "total_savings": 500000,
    "currency": "USD"
  }
}
```

---

### 5.3 Categories Domain

#### 20. List Categories
* **Endpoint:** `GET /api/v1/categories`
* **Auth Required:** Bearer JWT
* **Query Parameters:** `type` (`income` | `expense` | omit for all)
* **Response `200 OK`:**
```json
{
  "success": true,
  "data": [
    {
      "id": "cat_food",
      "name": "Food",
      "type": "expense",
      "icon": "bowl-food",
      "color": "#FF8A65"
    },
    {
      "id": "cat_salary",
      "name": "Salary",
      "type": "income",
      "icon": "wallet",
      "color": "#4CAF50"
    },
    {
      "id": "cat_kpi",
      "name": "KPI Bonus",
      "type": "income",
      "icon": "trophy",
      "color": "#2196F3"
    }
  ]
}
```

#### 21. Create Category
* **Endpoint:** `POST /api/v1/categories`
* **Auth Required:** Bearer JWT
* **Request Body:**
```json
{
  "name": "Medicine",
  "type": "expense",
  "icon": "pills",
  "color": "#E91E63"
}
```
* **Response `201 Created`:** Category object.

#### 22. Update Category
* **Endpoint:** `PATCH /api/v1/categories/:id`
* **Auth Required:** Bearer JWT
* **Request Body:**
```json
{
  "name": "Pharmacy & Medicine",
  "color": "#C2185B"
}
```

#### 23. Delete Category
* **Endpoint:** `DELETE /api/v1/categories/:id`
* **Auth Required:** Bearer JWT
* **Response `200 OK`:** Deletion confirmation.

---

### 5.4 Transactions Domain

#### 24. List / Filter Transactions
* **Endpoint:** `GET /api/v1/transactions`
* **Auth Required:** Bearer JWT
* **Query Parameters:**
  * `type`: `INCOME` | `EXPENSE` | `TRANSFER`
  * `category_id`: string filter
  * `account_id`: string filter
  * `date`: string (`YYYY-MM-DD`, e.g. `2026-09-16`)
  * `start_date`, `end_date`: range filters
  * `page`: integer (default: `1`)
  * `limit`: integer (default: `20`)
* **Response `200 OK`:**
```json
{
  "success": true,
  "data": [
    {
      "id": "tx_101",
      "date": "2026-09-05T00:00:00Z",
      "type": "EXPENSE",
      "category": {
        "id": "cat_food",
        "name": "Food",
        "icon": "bowl-food"
      },
      "account": {
        "id": "acc_01",
        "name": "Cash",
        "group": "cash"
      },
      "amount": 20000,
      "currency": "USD",
      "note": "Dinner with colleagues",
      "attachment_count": 2,
      "created_at": "2026-09-05T19:30:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total_records": 1,
    "total_pages": 1
  }
}
```

#### 25. Create Transaction Entry
* **Endpoint:** `POST /api/v1/transactions`
* **Auth Required:** Bearer JWT
* **Request Body:**
```json
{
  "type": "INCOME", // "INCOME" | "EXPENSE"
  "date": "2026-09-05",
  "amount": 5000, // cents ($50.00)
  "currency": "USD",
  "category_id": "cat_kpi",
  "account_id": "acc_01",
  "note": "Quarterly bonus",
  "description": "Performance evaluation bonus for Q3",
  "attachment_ids": ["att_83dfa", "att_94eba"]
}
```
* **Response `201 Created`:** Returns created transaction.

#### 26. Get Transaction Details
* **Endpoint:** `GET /api/v1/transactions/:id`
* **Auth Required:** Bearer JWT
* **Response `200 OK`:**
```json
{
  "success": true,
  "data": {
    "id": "tx_101",
    "date": "2026-09-05",
    "type": "EXPENSE",
    "expenditure_amount": 5000,
    "earning_amount": 0,
    "currency": "USD",
    "category": {
      "id": "cat_food",
      "name": "Food"
    },
    "account": {
      "id": "acc_01",
      "name": "Cash"
    },
    "note": "Business lunch",
    "description": "Lunch meeting with the design agency",
    "attachments": [
      {
        "id": "att_83dfa",
        "file_name": "Foodimg.png",
        "url": "https://assets.fundly.com/receipts/food_1.png",
        "file_size": 204850
      }
    ],
    "created_at": "2026-09-05T12:00:00Z",
    "updated_at": "2026-09-05T12:00:00Z"
  }
}
```

#### 27. Update Transaction
* **Endpoint:** `PATCH /api/v1/transactions/:id`
* **Auth Required:** Bearer JWT
* **Request Body:**
```json
{
  "date": "2026-09-05",
  "amount": 6000,
  "category_id": "cat_food",
  "account_id": "acc_01",
  "note": "Updated lunch total",
  "description": "Includes tip",
  "attachment_ids": ["att_83dfa"]
}
```
* **Response `200 OK`:** Updated transaction object.

#### 28. Delete Transaction
* **Endpoint:** `DELETE /api/v1/transactions/:id`
* **Auth Required:** Bearer JWT
* **Response `200 OK`:** Transaction deleted confirmation.

#### 29. Search Transactions
* **Endpoint:** `GET /api/v1/transactions/search`
* **Auth Required:** Bearer JWT
* **Query Parameters:** `q` (keyword query), `limit` (default: 20)
* **Response `200 OK`:** List of matching transactions, accounts, and categories.

#### 30. Transfer Between Accounts
* **Endpoint:** `POST /api/v1/transactions/transfer`
* **Auth Required:** Bearer JWT
* **Request Body:**
```json
{
  "from_account_id": "acc_01",
  "to_account_id": "acc_02",
  "amount": 10000, // cents ($100.00)
  "note": "ATM deposit"
}
```
* **Response `201 Created`:** Transfer transaction details.

---

### 5.5 Budgets Domain

#### 31. List Budgets for Period
* **Endpoint:** `GET /api/v1/budgets`
* **Auth Required:** Bearer JWT
* **Query Parameters:** `month` (`YYYY-MM`, e.g. `2026-09`), `category_id` (optional)
* **Response `200 OK`:**
```json
{
  "success": true,
  "data": [
    {
      "id": "bg_01",
      "category": {
        "id": "cat_food",
        "name": "Food",
        "icon": "bowl-food"
      },
      "month": "2026-09",
      "limit_amount": 80000, // $800.00
      "expenditure": 45480, // $454.80
      "remaining_amount": 34520, // $345.20
      "usage_percentage": 56.85,
      "remaining_percentage": 43.15,
      "currency": "USD",
      "created_at": "2026-09-01T00:00:00Z"
    }
  ]
}
```

#### 32. Create Budget Allocation
* **Endpoint:** `POST /api/v1/budgets`
* **Auth Required:** Bearer JWT
* **Request Body:**
```json
{
  "month": "2026-09",
  "limit_amount": 62075, // $620.75
  "category_id": "cat_food",
  "currency": "USD",
  "note": "Monthly food allowance",
  "description": "Dining out and groceries budget for September",
  "attachment_ids": []
}
```
* **Response `201 Created`:** Created budget entity.

#### 33. Get Budget Details
* **Endpoint:** `GET /api/v1/budgets/:id`
* **Auth Required:** Bearer JWT
* **Response `200 OK`:** Detailed budget view with real-time calculated expenditure and remaining percentage.

#### 34. Update Budget
* **Endpoint:** `PATCH /api/v1/budgets/:id`
* **Auth Required:** Bearer JWT
* **Request Body:**
```json
{
  "limit_amount": 70000,
  "note": "Increased limit for holidays",
  "description": "Updated notes"
}
```

#### 35. Delete Budget
* **Endpoint:** `DELETE /api/v1/budgets/:id`
* **Auth Required:** Bearer JWT
* **Response `200 OK`:** Budget deleted confirmation.

---

### 5.6 Stats & Analytics Domain

#### 36. Category Breakdown (Donut Chart)
* **Endpoint:** `GET /api/v1/stats/breakdown`
* **Auth Required:** Bearer JWT
* **Query Parameters:**
  * `type`: `INCOME` | `EXPENSE` (required)
  * `month`: `YYYY-MM` (e.g. `2026-08`, defaults to current month)
  * `account_id`: optional account filter
* **Response `200 OK`:**
```json
{
  "success": true,
  "data": {
    "type": "INCOME",
    "month": "2026-08",
    "total_amount": 20000, // $200.00
    "currency": "USD",
    "breakdown": [
      {
        "category_id": "cat_salary",
        "category_name": "Salary",
        "amount": 5600,
        "percentage": 28.0,
        "color": "#4CAF50"
      },
      {
        "category_id": "cat_kpi",
        "category_name": "KPI Bonus",
        "amount": 5600,
        "percentage": 28.0,
        "color": "#2196F3"
      },
      {
        "category_id": "cat_other",
        "category_name": "Other",
        "amount": 8800,
        "percentage": 44.0,
        "color": "#9E9E9E"
      }
    ]
  }
}
```

#### 37. Overview Metrics
* **Endpoint:** `GET /api/v1/stats/overview`
* **Auth Required:** Bearer JWT
* **Query Parameters:** `month` (`YYYY-MM`)
* **Response `200 OK`:** Returns total income, total expense, net balance, and budget utilization.

---

### 5.7 Comparison & Reporting Domain

#### 38. Period Comparison
* **Endpoint:** `POST /api/v1/reports/compare`
* **Auth Required:** Bearer JWT
* **Request Body:**
```json
{
  "type": "INCOME", // "INCOME" | "EXPENSE"
  "primary_period": {
    "month": "2026-08"
  },
  "comparison_period": {
    "month": "2026-07"
  },
  "category_ids": ["cat_allowance", "cat_salary", "cat_donation", "cat_kpi"]
}
```
* **Response `200 OK`:**
```json
{
  "success": true,
  "data": {
    "compared_balance": 20000, // +$200.00 delta
    "currency": "USD",
    "primary_total": 24300,
    "comparison_total": 4300,
    "categories": [
      {
        "category_id": "cat_salary",
        "name": "Salary",
        "primary_amount": 24300,
        "comparison_amount": 20000,
        "difference": 4300
      },
      {
        "category_id": "cat_allowance",
        "name": "Allowance",
        "primary_amount": 24300,
        "comparison_amount": 24300,
        "difference": 0
      }
    ]
  }
}
```

---

### 5.8 Attachments Domain

#### 39. Upload Receipt / Attachment
* **Endpoint:** `POST /api/v1/attachments`
* **Auth Required:** Bearer JWT
* **Content-Type:** `multipart/form-data`
* **Form Field:** `file` (image/pdf, max 10MB)
* **Response `201 Created`:**
```json
{
  "success": true,
  "data": {
    "id": "att_83dfa",
    "file_name": "Foodimg.png",
    "content_type": "image/png",
    "file_size": 154200,
    "url": "https://assets.fundly.com/attachments/Foodimg.png",
    "created_at": "2026-09-16T10:00:00Z"
  }
}
```

#### 40. Get Attachment
* **Endpoint:** `GET /api/v1/attachments/:id`
* **Auth Required:** Bearer JWT
* **Response `200 OK`:** Metadata and presigned download URL.

#### 41. Delete Attachment
* **Endpoint:** `DELETE /api/v1/attachments/:id`
* **Auth Required:** Bearer JWT
* **Response `200 OK`:** Deletion confirmation.

---

### 5.9 User Configuration Domain

#### 42. Get User Configuration
* **Endpoint:** `GET /api/v1/configuration`
* **Auth Required:** Bearer JWT
* **Response `200 OK`:**
```json
{
  "success": true,
  "data": {
    "default_currency": "USD",
    "language": "en",
    "theme": "dark",
    "updated_at": "2026-01-10T00:00:00Z"
  }
}
```

#### 43. Update User Configuration
* **Endpoint:** `PUT /api/v1/configuration`
* **Auth Required:** Bearer JWT
* **Request Body:**
```json
{
  "default_currency": "USD"
}
```
* **Response `200 OK`:** Updated configuration object.

---

### 5.10 Feedback Domain

#### 44. Submit Feedback
* **Endpoint:** `POST /api/v1/feedback`
* **Auth Required:** Bearer JWT
* **Request Body:**
```json
{
  "subject": "App Performance Feedback",
  "description": "Transactions sync immediately and the UI is intuitive!"
}
```
* **Response `201 Created`:**
```json
{
  "success": true,
  "message": "Thank you for your feedback! Our team has received your submission."
}
```

---

### 5.11 Terms & Conditions Domain

#### 45. Get Latest Terms
* **Endpoint:** `GET /api/v1/terms/latest`
* **Auth Required:** Optional (or Public)
* **Response `200 OK`:**
```json
{
  "success": true,
  "data": {
    "version": "2.1",
    "effective_date": "2026-01-01",
    "title": "Terms & Conditions",
    "clauses": [
      {
        "clause_number": "01",
        "title": "Clause 01",
        "content": "Lorem ipsum dolor sit amet consectetur. In faucibus ante sit nisl justo..."
      },
      {
        "clause_number": "02",
        "title": "Clause 02",
        "content": "Lorem ipsum dolor sit amet consectetur. In faucibus ante sit nisl justo..."
      }
    ]
  }
}
```

#### 46. Acknowledge Terms
* **Endpoint:** `POST /api/v1/terms/acknowledge`
* **Auth Required:** Bearer JWT
* **Request Body:**
```json
{
  "version": "2.1"
}
```
* **Response `200 OK`:**
```json
{
  "success": true,
  "message": "Terms and conditions acknowledged"
}
```

---

## 6. Mapping to DDD Go Architecture

Each API endpoint maps directly into our project's DDD structure:

```
                  ┌──────────────────────────────────────────────┐
                  │          HTTP Request (API Client)           │
                  └──────────────────────┬───────────────────────┘
                                         ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│  INFRASTRUCTURE LAYER (Primary / Driving Adapters)                          │
│  internal/infrastructure/http/                                              │
│  - auth_handler.go        - account_handler.go      - transaction_handler.go│
│  - budget_handler.go      - stats_handler.go        - feedback_handler.go   │
└──────────────────────────────────────┬──────────────────────────────────────┘
                                       ▼ DTOs & Commands
┌─────────────────────────────────────────────────────────────────────────────┐
│  APPLICATION LAYER (Use Cases & Orchestration)                              │
│  internal/application/                                                      │
│  - transfer/service.go    - auth/service.go         - transaction/service.go│
│  - budget/service.go      - stats/service.go        - account/service.go    │
└──────────────────────────────────────┬──────────────────────────────────────┘
                                       ▼ Calls Domain Logic & Interfaces
┌─────────────────────────────────────────────────────────────────────────────┐
│  DOMAIN LAYER (Core Business Entities & Interfaces)                         │
│  internal/domain/                                                           │
│  - user/                  - account/                - transaction/          │
│  - terms/                 - stat/                   (pure Go, 0 outer deps) │
└──────────────────────────────────────▲──────────────────────────────────────┘
                                       │ Implements Interfaces
┌──────────────────────────────────────┴──────────────────────────────────────┐
│  INFRASTRUCTURE LAYER (Secondary / Driven Persistence Adapters)             │
│  internal/infrastructure/persistence/postgres/                              │
│  - user_repo.go           - account_repo.go         - tx_repo.go            │
└─────────────────────────────────────────────────────────────────────────────┘
```

* **Handler Responsibilities:**
  1. Parse URL parameters, query strings, and JSON body.
  2. Perform request validation (e.g. required fields, valid email format).
  3. Invoke the corresponding Application Service.
  4. Write out the standard JSON response envelope (`success`, `data`, `error`).
* **Application Service Responsibilities:**
  1. Transaction boundaries (e.g. money transfer debit + credit + history).
  2. Fetching domain models via repository interfaces.
  3. Calling domain business methods (e.g. `account.Withdraw()`).
  4. Persisting updated aggregates.
* **Domain Responsibilities:**
  1. Enforce invariant rules (e.g. non-negative balance, matching currency).
  2. Domain events and error definitions.
