# Testing HTTP Handlers (Gin)

```go
func Test_Handler_CreateUser(t *testing.T) {
    tests := []struct {
        name       string
        body       string
        mockResp   *User
        mockErr    error
        wantStatus int
    }{
        {
            name:       "valid request",
            body:       `{"email":"user@test.com","name":"Test"}`,
            mockResp:   &User{ID: "1", Email: "user@test.com"},
            wantStatus: http.StatusCreated,
        },
        {
            name:       "missing email",
            body:       `{"name":"Test"}`,
            wantStatus: http.StatusBadRequest,
        },
        {
            name:       "invalid json",
            body:       `{invalid}`,
            wantStatus: http.StatusBadRequest,
        },
        {
            name:       "service error",
            body:       `{"email":"user@test.com","name":"Test"}`,
            mockErr:    errors.New("mock error"),
            wantStatus: http.StatusInternalServerError,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            svc := createMock{
                createResp: tt.mockResp,
                createErr:  tt.mockErr,
            }
            h := NewHandler(svc)

            w := httptest.NewRecorder()
            r, _ := http.NewRequest(http.MethodPost, "/users", strings.NewReader(tt.body))
            r.Header.Set("Content-Type", "application/json")

            mux := gin.New()
            mux.POST("/users", h.CreateUser)
            mux.ServeHTTP(w, r)

            if w.Code != tt.wantStatus {
                t.Errorf("status = %d, want %d", w.Code, tt.wantStatus)
            }
        })
    }
}
```

For query parameters:

```go
r, _ := http.NewRequest(http.MethodGet, "/users", nil)
q := r.URL.Query()
q.Set("status", "active")
q.Set("page", "1")
r.URL.RawQuery = q.Encode()
```

> For stdlib `http.Handler` (no Gin), call the handler directly: `h.CreateUser(w, r)`
