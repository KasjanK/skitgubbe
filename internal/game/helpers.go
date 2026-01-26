package game

import (
	"fmt"
	"log"
	"math/rand"
	"slices"

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
	for _, player := range game.Players {
		fmt.Printf("PlayerID: %s\n", player.ID)
		fmt.Printf("Hand:\n")
		for _, card := range player.Hand {
			fmt.Printf("Rank: %d, Suit: %d\n", card.Rank, card.Suit)
		}
	}

	fmt.Printf("All cards are dealed. Cards left in deck: %v", len(game.Deck))

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

func VisibleStateFor(gs *GameState, viewer PlayerID) VisibleState {
	var you PlayerState
	others := make([]VisiblePlayer, 0, len(gs.Players) - 1)

	for _, player := range gs.Players {
		if player.ID == viewer {
			you = player
		} else {
			others = append(others, VisiblePlayer{
				ID: 	player.ID,
				HandSize: len(player.Hand),
			})
		}
	}

	return VisibleState{
		ID: 		   gs.ID,
		You: 		   you,
		Others:		   others,
		Pile: 		   gs.Pile,
		CurrentPlayer: gs.CurrentPlayer,
	}
}

func ApplyMove(gs *GameState, playerID PlayerID, move Move) error {
	if gs.CurrentPlayer != playerID {
		return fmt.Errorf("not your turn")
	}
	var player *PlayerState
	
	for i  := range gs.Players {
		if gs.Players[i].ID == playerID {
			player = &gs.Players[i]
			break
		}
	}

	if player == nil {
		return fmt.Errorf("Player not in game")
	}

	if move.Move == MoveTypePlayCard {
		// check for card in hand
		idx := -1
		for i, card := range player.Hand {
			if card.Suit == move.Card.Suit && card.Rank == move.Card.Rank {
				idx = i
				break
			}
		}
		if idx == -1 {
			return fmt.Errorf("Card not in hand")
		}

		// check pile 
		if len(gs.Pile) == 0 {
			log.Printf("Pile is empty, card %v played", *move.Card)
		} else {
			top := gs.Pile[len(gs.Pile) - 1]
			if move.Card.Rank < top.Rank {
				return fmt.Errorf("Card too low")
			}
		}
		
		// apply move
		player.Hand = slices.Delete(player.Hand, idx, idx + 1)
		gs.Pile = append(gs.Pile, *move.Card)
		log.Printf("Card %v played by %v", *move.Card, player.ID)

		// draw card after move
		if len(player.Hand) != 3 {
			for i := 0; i < (3 - len(player.Hand)); i++ {
				if len(gs.Deck) == 0 {
					break 
				}
				player.Hand = append(player.Hand, gs.Deck[len(gs.Deck) - 1])
				gs.Deck = slices.Delete(gs.Deck, len(gs.Deck) - 1, len(gs.Deck))
				fmt.Printf("Cards left in deck: %v", len(gs.Deck))
			}
		}
	}

	if move.Move == MoveTypePickUp {
		if len(gs.Pile) == 0 {
			return fmt.Errorf("Pile is empty, nothing to pick up")
		}
		player.Hand = append(player.Hand, gs.Pile...)
		gs.Pile = nil
	}

	if move.Move == MoveTypeChance {
		if len(gs.Pile) == 0 {
			return fmt.Errorf("Pile is empty, not allowed to take a chance")
		}

		if len(gs.Deck) == 0 {
			return fmt.Errorf("Deck is empty, no cards left")
		}

		chanceCard := gs.Deck[len(gs.Deck) - 1]
		gs.Deck = slices.Delete(gs.Deck, len(gs.Deck) - 1, len(gs.Deck))		
		fmt.Printf("Chancecard taken: %v. Cards left: %v", chanceCard, len(gs.Deck))

		top := gs.Pile[len(gs.Pile) - 1]
		if chanceCard.Rank < top.Rank {
			gs.Pile = append(gs.Pile, chanceCard)
			player.Hand = append(player.Hand, gs.Pile...)
			gs.Pile = nil
			fmt.Printf("Chancecard too low, picked up pile. Cards left: %v\n", len(gs.Deck))
		} else {
			gs.Pile = append(gs.Pile, chanceCard)
			fmt.Printf("Chancecard played: %v. Cards left: %v\n", chanceCard, len(gs.Deck))
		}
	}

	if len(gs.Players) == 0 {
		return fmt.Errorf("no players left")
	}
	
	// next player
	idx := 0
	for i, player := range gs.Players {
		if player.ID == gs.CurrentPlayer {
			idx = i
			break
		}
	}

	nextPlayer := (idx + 1) % len(gs.Players)
	gs.CurrentPlayer = gs.Players[nextPlayer].ID

	return nil
}
