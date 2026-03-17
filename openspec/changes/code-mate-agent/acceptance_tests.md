# Acceptance Tests: Learning Record Agent

Acceptance test cases organized by implementation phase. Run tests after each phase to verify functionality before proceeding.

---

## Phase 4: CRUD API Tests

### AT-1: Confirm creates a learning record

**Endpoint:** POST `/api/v1/chat-record/confirm`  
**Prerequisite:** Valid auth token

| Step | Action | Expected |
|------|--------|----------|
| 1 | POST confirm with valid body | 201 Created |
| 2 | Response contains record with id, input_type, user_input | ✓ |
| 3 | Query DB: record exists with correct fields | ✓ |

```json
// Request
{
  "user_input": "serendipity",
  "input_type": "word",
  "response_payload": {
    "explanation": "意外发现美好事物的运气",
    "pronunciation": "/ˌserənˈdɪpəti/",
    "example": "Finding that book was pure serendipity."
  }
}

// Expected Response (201)
{
  "id": "uuid-...",
  "input_type": "word",
  "user_input": "serendipity",
  "created_at": "2026-01-31T..."
}
```

---

### AT-2: Confirm without auth returns 401

**Endpoint:** POST `/api/v1/chat-record/confirm`  
**Prerequisite:** No auth token

| Step | Action | Expected |
|------|--------|----------|
| 1 | POST confirm without Authorization header | 401 Unauthorized |
| 2 | Query DB: no record created | ✓ |

---

### AT-3: List returns user's records with pagination

**Endpoint:** GET `/api/v1/chat-record/list`  
**Prerequisite:** User has 15 records

| Step | Action | Expected |
|------|--------|----------|
| 1 | GET /list?page=1&size=10 | 200, 10 records, total=15 |
| 2 | GET /list?page=2&size=10 | 200, 5 records, total=15 |
| 3 | Records belong to current user only | ✓ |

```json
// Expected Response
{
  "records": [
    {"id": "...", "input_type": "word", "user_input": "...", "created_at": "..."},
    ...
  ],
  "total": 15,
  "page": 1,
  "page_size": 10
}
```

---

### AT-4: List filters by type

**Endpoint:** GET `/api/v1/chat-record/list?type=word`  
**Prerequisite:** User has records of different types

| Step | Action | Expected |
|------|--------|----------|
| 1 | GET /list?type=word | 200, only word records |
| 2 | GET /list?type=question | 200, only question records |
| 3 | GET /list?type=invalid | 200, empty or all (define behavior) |

---

### AT-5: Get single record by ID

**Endpoint:** GET `/api/v1/chat-record/:id`

| Step | Action | Expected |
|------|--------|----------|
| 1 | GET /:id with valid ID owned by user | 200, full record detail |
| 2 | GET /:id with ID owned by other user | 403 or 404 |
| 3 | GET /:id with non-existent ID | 404 |

---

### AT-6: Delete soft-deletes a record

**Endpoint:** DELETE `/api/v1/chat-record/:id`

| Step | Action | Expected |
|------|--------|----------|
| 1 | DELETE /:id with valid ID | 204 No Content |
| 2 | GET /:id for same ID | 404 (soft-deleted) |
| 3 | Query DB: deleted_at is set | ✓ |
| 4 | GET /list: record not in list | ✓ |

---

### AT-7: Stats returns counts by type

**Endpoint:** GET `/api/v1/chat-record/stats`  
**Prerequisite:** User has 5 words, 3 sentences, 2 questions, 1 idea

| Step | Action | Expected |
|------|--------|----------|
| 1 | GET /stats | 200 |
| 2 | Response total = 11 | ✓ |
| 3 | Response by_type = {word: 5, sentence: 3, question: 2, idea: 1} | ✓ |

```json
// Expected Response
{
  "total": 11,
  "by_type": {
    "word": 5,
    "sentence": 3,
    "question": 2,
    "idea": 1
  },
  "streak": 3
}
```

---

## Phase 6: Submit API Tests

### AT-8: Submit word returns classification and response

**Endpoint:** POST `/api/v1/chat-record/submit`

| Step | Action | Expected |
|------|--------|----------|
| 1 | POST submit: `{"user_input": "serendipity"}` | 200 |
| 2 | Response input_type = "word" | ✓ |
| 3 | Response has explanation in Chinese | ✓ |
| 4 | Response has pronunciation | ✓ |
| 5 | Response has example | ✓ |

```json
// Expected Response
{
  "input_type": "word",
  "response_payload": {
    "explanation": "意外发现美好事物的运气或能力",
    "pronunciation": "/ˌserənˈdɪpəti/",
    "example": "Finding that rare book was pure serendipity."
  }
}
```

---

### AT-9: Submit sentence returns classification and response

**Endpoint:** POST `/api/v1/chat-record/submit`

| Step | Action | Expected |
|------|--------|----------|
| 1 | POST submit: `{"user_input": "Time flies when you're having fun."}` | 200 |
| 2 | Response input_type = "sentence" | ✓ |
| 3 | Response has explanation in Chinese | ✓ |
| 4 | Response has example | ✓ |

---

### AT-10: Submit question returns classification and answer

**Endpoint:** POST `/api/v1/chat-record/submit`

| Step | Action | Expected |
|------|--------|----------|
| 1 | POST submit: `{"user_input": "What is the capital of France?"}` | 200 |
| 2 | Response input_type = "question" | ✓ |
| 3 | Response has answer | ✓ |

```json
// Expected Response
{
  "input_type": "question",
  "response_payload": {
    "answer": "The capital of France is Paris."
  }
}
```

---

### AT-11: Submit idea returns classification and plan

**Endpoint:** POST `/api/v1/chat-record/submit`

| Step | Action | Expected |
|------|--------|----------|
| 1 | POST submit: `{"user_input": "I want to build a habit tracker app"}` | 200 |
| 2 | Response input_type = "idea" | ✓ |
| 3 | Response has plan (array of steps) | ✓ |

```json
// Expected Response
{
  "input_type": "idea",
  "response_payload": {
    "plan": [
      "1. Define core features: habit creation, tracking, reminders",
      "2. Design database schema for habits and tracking data",
      "3. Create backend API endpoints",
      "4. Build frontend UI with calendar view",
      "5. Add push notifications for reminders"
    ]
  }
}
```

---

### AT-12: Submit does NOT write to database

**Endpoint:** POST `/api/v1/chat-record/submit`

| Step | Action | Expected |
|------|--------|----------|
| 1 | Count learning_records for user | N |
| 2 | POST submit 3 times with different inputs | 200 each |
| 3 | Count learning_records for user again | Still N (unchanged) |

---

## Phase 7: Memory Tests (Optional)

### AT-13: Submit returns similar records

**Prerequisite:** User has confirmed record for "ephemeral"

| Step | Action | Expected |
|------|--------|----------|
| 1 | POST submit: `{"user_input": "ephemeral"}` (same word) | 200 |
| 2 | Response includes similar_records | ✓ |
| 3 | similar_records contains the previous record | ✓ |

---

### AT-14: Session context influences response

**Prerequisite:** Session with prior conversation

| Step | Action | Expected |
|------|--------|----------|
| 1 | POST submit: `{"user_input": "serendipity", "session_id": "abc"}` | 200 |
| 2 | POST submit: `{"user_input": "use it in a sentence", "session_id": "abc"}` | 200 |
| 3 | Response references "serendipity" from session context | ✓ |

---

## Test Execution Order

Run tests after completing each phase:

| After Phase | Run Tests | Purpose |
|-------------|-----------|---------|
| Phase 4 | AT-1 to AT-7 | Verify CRUD API |
| Phase 6 | AT-8 to AT-12 | Verify Submit + Agent |
| Phase 7 | AT-13 to AT-14 | Verify Memory features |

---

## Test Implementation Notes

**Go httptest example:**
```go
func TestConfirmCreatesRecord(t *testing.T) {
    req := httptest.NewRequest("POST", "/api/v1/chat-record/confirm", body)
    req.Header.Set("Authorization", "Bearer "+validToken)
    w := httptest.NewRecorder()
    router.ServeHTTP(w, req)
    
    assert.Equal(t, 201, w.Code)
    // verify response body and DB record
}
```

**Python pytest example:**
```python
def test_confirm_creates_record(client, auth_headers):
    resp = client.post("/api/v1/chat-record/confirm", 
                       json={"user_input": "test", "input_type": "word", ...},
                       headers=auth_headers)
    assert resp.status_code == 201
    assert "id" in resp.json()
```
