package main

import (
	"fmt"
	"sync"
	"testing"

	"github.com/Kasjank/skitgubbe/internal/game"
)

func TestGameConcurrency(t *testing.T) {
    cfg := &apiConfig{
        games: make(map[string]*game.GameState),
		mu: sync.RWMutex{},
    }
    
    id := "test-game"
    cfg.games[id] = &game.GameState{ID: id}

    var wg sync.WaitGroup
    for i := 0; i < 100; i++ {
        wg.Add(1)
        go func(n int) {
            defer wg.Done()
            if n % 2 == 0 {
                _, _ = cfg.GetGame(id) 
            } else {
                cfg.mu.Lock()
                cfg.games[fmt.Sprintf("new-id-%d", n)] = &game.GameState{}
                cfg.mu.Unlock()
            }
        }(i)
    }
    wg.Wait()
}

func TestCreateAndJoinRoom(t *testing.T) {
    cfg := &apiConfig{
        rooms: make(map[string]*game.Room),
    }

    room, err := cfg.CreateRoom("user-1", "Alice")
    if err != nil {
        t.Fatalf("Expected no error, got %v", err)
    }
    if len(cfg.rooms) != 1 {
        t.Errorf("Expected 1 room, got %d", len(cfg.rooms))
    }

    err = cfg.JoinRoom(room.ID, "user-1", "Alice")
    if err == nil {
        t.Error("Expected error when joining a room twice, got nil")
    }
}

func TestRoomCleanup(t *testing.T) {
    cfg := &apiConfig{rooms: make(map[string]*game.Room)}
    room, _ := cfg.CreateRoom("p1", "Alice")

    err := cfg.LeaveRoom(room.ID, "p1")
    if err != nil {
        t.Fatalf("LeaveRoom failed: %v", err)
    }

    cfg.mu.RLock()
    _, exists := cfg.rooms[room.ID]
    cfg.mu.RUnlock()
    
    if exists {
        t.Error("Room should have been deleted after last player left")
    }
}
