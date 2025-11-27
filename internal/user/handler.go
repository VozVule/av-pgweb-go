package user

import (
    "encoding/json"
    "fmt"
    "math/rand"
    "net/http"
    "sync"
)

// Handler exposes HTTP endpoints for working with users.
type Handler struct {
    mu    sync.RWMutex
    users map[int]User
}

// User is a simple DTO returned by the API.
type User struct {
    ID   int    `json:"id"`
    Name string `json:"name"`
}

// NewHandler seeds a demo data set and returns a Handler.
func NewHandler() *Handler {
    return &Handler{
        users: map[int]User{
            1: {ID: 1, Name: "Ada"},
            2: {ID: 2, Name: "Linus"},
        },
    }
}

// Register attaches the handler's routes to the provided mux.
func (h *Handler) Register(mux *http.ServeMux) {
    mux.HandleFunc("/users", h.usersCollection)
    mux.HandleFunc("/users/", h.userEntity)
}

func (h *Handler) usersCollection(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case http.MethodGet:
        h.listUsers(w)
    case http.MethodPost:
        h.createUser(w, r)
    default:
        w.WriteHeader(http.StatusMethodNotAllowed)
    }
}

func (h *Handler) userEntity(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        w.WriteHeader(http.StatusMethodNotAllowed)
        return
    }

    var id int
    if _, err := fmt.Sscanf(r.URL.Path, "/users/%d", &id); err != nil {
        http.NotFound(w, r)
        return
    }

    h.mu.RLock()
    user, ok := h.users[id]
    h.mu.RUnlock()
    if !ok {
        http.NotFound(w, r)
        return
    }

    writeJSON(w, http.StatusOK, user)
}

func (h *Handler) listUsers(w http.ResponseWriter) {
    h.mu.RLock()
    users := make([]User, 0, len(h.users))
    for _, u := range h.users {
        users = append(users, u)
    }
    h.mu.RUnlock()

    writeJSON(w, http.StatusOK, users)
}

func (h *Handler) createUser(w http.ResponseWriter, r *http.Request) {
    var payload struct {
        Name string `json:"name"`
    }
    if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
        http.Error(w, "invalid JSON", http.StatusBadRequest)
        return
    }
    if payload.Name == "" {
        http.Error(w, "name is required", http.StatusBadRequest)
        return
    }

    h.mu.Lock()
    id := rand.Intn(10_000)
    user := User{ID: id, Name: payload.Name}
    h.users[id] = user
    h.mu.Unlock()

    writeJSON(w, http.StatusCreated, user)
}

func writeJSON(w http.ResponseWriter, status int, data any) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    _ = json.NewEncoder(w).Encode(data)
}
