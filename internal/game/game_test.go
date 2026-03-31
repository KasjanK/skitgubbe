package game


import (
	"testing"
)

func TestApplyMoveSpecialCards(t *testing.T) {
	// 1. Define the scenarios
	tests := []struct {
		name          string
		playedRank    Rank
		initialPile   []Card
		expectedPile  int  // expected number of cards in pile after move
		shouldError   bool
	}{
		{
			name:        "Normal card (8) on lower card (5) should work",
			playedRank:  8,
			initialPile: []Card{{Rank: 5, Suit: 2}},
			expectedPile: 2,
			shouldError: false,
		},
		{
			name:        "Low card (3) on high card (9) should fail",
			playedRank:  3,
			initialPile: []Card{{Rank: 9, Suit: 0}},
			expectedPile: 1,
			shouldError: true,
		},
		{
			name:        "Playing a 10 should clear the pile",
			playedRank:  10,
			initialPile: []Card{{Rank: 9, Suit: 0}, {Rank: 8, Suit: 1}},
			expectedPile: 0,
			shouldError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p1 := PlayerState{
				ID:   "p1",
				Hand: []Card{{Rank: tt.playedRank, Suit: 1}},
			}
			gs := &GameState{
				Players:       []PlayerState{p1},
				Pile:          tt.initialPile,
				CurrentPlayer: "p1",
				Phase:         PhasePlay, // Ensure we aren't in Setup
			}

			move := Move{
				Move: MoveTypePlayMany,
				Indices: []int{0},
			}

			err := ApplyMove(gs, "p1", move)

			if (err != nil) != tt.shouldError {
				t.Errorf("ApplyMove() error = %v, wantErr %v", err, tt.shouldError)
			}
			if len(gs.Pile) != tt.expectedPile {
				t.Errorf("Pile length = %d, want %d", len(gs.Pile), tt.expectedPile)
			}
		})
	}
}
