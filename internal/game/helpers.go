package game

import (
	"math/rand"

	"github.com/google/uuid"
)

func NewDeck() []Card {
	var deck []Card
	for _, suit := range AllSuits {
		for _, rank := range AllRanks {
			card := Card{Suit: suit, Rank: rank}
			deck = append(deck, card)
		}
	}
	return deck
}

func ShuffleDeck(deck []Card) {
	// Fisher-Yates
	deckLength := len(deck)
	for i := 0; i < deckLength; i++ {
		r := i + rand.Intn(deckLength - i)
		deck[i], deck[r] = deck[r], deck[i]
	}
}

func DealCards(players []PlayerState, deck []Card, cardsPerPlayer int) ([]PlayerState, []Card) {
	idx := 0
	for player := range players {
		for i := 0; i < cardsPerPlayer; i++{
			players[player].Hand = append(players[player].Hand, deck[idx])
			idx++
			players[player].FacedownTableCards = append(players[player].FacedownTableCards, deck[idx])
			idx++
			players[player].FaceupTableCards = append(players[player].FaceupTableCards, deck[idx])
			idx++
		}
	}
	return players, deck[idx:]
}

func NewGame(players []PlayerState) *GameState {
	gameID := uuid.New()
	deck := NewDeck()
	ShuffleDeck(deck)
	players, remainingDeck := DealCards(players, deck, 3)

	game := &GameState {
		ID: gameID.String(),
		Players: players,
		CurrentPlayer: players[rand.Intn(len(players))].ID,
		Deck: remainingDeck,
		Pile: nil,
	}
	return game
}

func NewRoom(ownerID PlayerID) *Room {
	roomID := uuid.New()
	
	players := []PlayerState{
		{ID: ownerID},
	}

	room := &Room {
		ID:		 roomID.String(),
		OwnerID: ownerID,
		Players: players,
		Ready:   make(map[PlayerID]bool),
		Started: false,
		GameID:  "",
	}

	return room
}
