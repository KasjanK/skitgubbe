package game

import (
	"fmt"
	"log"
	"math/rand"
	"slices"
	"sort"

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
	for i := range deckLength {
		r := i + rand.Intn(deckLength - i)
		deck[i], deck[r] = deck[r], deck[i]
	}
}

func DealCards(players []PlayerState, deck []Card, cardsPerPlayer int) ([]PlayerState, []Card) {
	idx := 0
	n := len(players)

	for  range cardsPerPlayer {
		for player := range n {
			players[player].FacedownTableCards = append(players[player].FacedownTableCards, deck[idx])
			idx++
		}
	}

	for range cardsPerPlayer {
		for player := range n {
			players[player].FaceupTableCards = append(players[player].FaceupTableCards, deck[idx])
			idx++
		}
	}

	for range cardsPerPlayer {
		for player := range n {
			players[player].Hand = append(players[player].Hand, deck[idx])
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
		ID: 		   gameID.String(),
		Players:       players,
		CurrentPlayer: players[rand.Intn(len(players))].ID,
		Deck: 		   remainingDeck,
		Pile: 		   nil,
		Phase: 		   PhaseSetup,
	}

	for _, player := range game.Players {
		fmt.Printf("PlayerID: %s\n", player.ID)
		fmt.Printf("Hand:\n")
		for _, card := range player.Hand {
			fmt.Printf("Rank: %d, Suit: %d\n", card.Rank, card.Suit)
		}
		fmt.Printf("Facedown:\n")
		for _, card := range player.FacedownTableCards {
			fmt.Printf("Rank: %d, Suit: %d\n", card.Rank, card.Suit)
		}
	}

	fmt.Printf("All cards are dealed. Cards left in deck: %v", len(game.Deck))
	fmt.Printf("Game is in %s phase", PhaseSetup)

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
				FacedownTableCardsSize: len(player.FacedownTableCards),
				OthersFaceupTableCards: player.FaceupTableCards,
			})
		}
	}

	return VisibleState{
		ID: 		   gs.ID,
		Phase:		   gs.Phase,
		You: 		   you,
		Others:		   others,
		Pile: 		   gs.Pile,
		CurrentPlayer: gs.CurrentPlayer,
	}
}

func ApplyMove(gs *GameState, playerID PlayerID, move Move) error {
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

	if gs.Phase == PhaseSetup {
		switch move.Move {
		case MoveTypeSwapFaceUp:
			return applySwapFaceUp(player, move)
		case MoveTypeReadySetup:
			return applyReadySetup(gs, player)
		default:
			return fmt.Errorf("can't do that in setup phase")
		}
	}

	if gs.CurrentPlayer != playerID {
		return fmt.Errorf("not your turn")
	}

	specialCard := false

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
			if move.Card.Rank < top.Rank && move.Card.Rank != 10 && move.Card.Rank != 2 {
				return fmt.Errorf("Card too low")
			}
		}
		
		// apply move
		player.Hand = slices.Delete(player.Hand, idx, idx + 1)
		gs.Pile = append(gs.Pile, *move.Card)
		log.Printf("Card %v played by %v", *move.Card, player.ID)

		// draw card after move
		if len(player.Hand) < 3 {
			for i := 0; i < (3 - len(player.Hand)); i++ {
				if len(gs.Deck) == 0 {
					break 
				}
				player.Hand = append(player.Hand, gs.Deck[len(gs.Deck) - 1])
				gs.Deck = slices.Delete(gs.Deck, len(gs.Deck) - 1, len(gs.Deck))
				fmt.Printf("Cards left in deck: %v", len(gs.Deck))
			}
		}

		if len(player.Hand) == 0 &&
		len(player.FaceupTableCards) == 0 &&
		len(player.FacedownTableCards) == 0 {

			last := gs.Pile[len(gs.Pile)-1] 

			isSpecialOrAce := last.Rank == 2 || last.Rank == 10 || last.Rank == 14
			if isSpecialOrAce {
				fmt.Println("You can't go out on a special card.")

				if len(gs.Pile) > 0 {
					player.Hand = append(player.Hand, gs.Pile...)
					gs.Pile = nil
				}
				return nil
			}

			fmt.Printf("Player %s won\n", player.ID)
			for i, p := range gs.Players {
				if p.ID == player.ID {
					gs.Players = slices.Delete(gs.Players, i, i+1)
					break
				}
			}
			return nil
		}	

		if move.Card.Rank == 10 || lastFourSame(gs) {
			gs.Pile = nil
			specialCard = true
		} else if move.Card.Rank == 2 {
			specialCard = true
		}
	}

	if move.Move == MoveTypePlayMany {
		// check if any cards are selected
		if len(move.Indices) == 0 {
			return fmt.Errorf("No cards selected")
		}

		idxs := append([]int(nil), move.Indices...)
		sort.Sort(sort.Reverse(sort.IntSlice(idxs)))

		var rank Rank
		// check if all cards are same rank
		for i, idx := range idxs {
			if idx < 0 || idx >= len(player.Hand) {
				return fmt.Errorf("invalid index: %d", idx)
			}

		    card := player.Hand[idx]

			if i == 0 {
				rank = card.Rank
			} else if card.Rank != rank {
				return fmt.Errorf("all cards must be the same rank")
			}
		}

		// check pile
		if len(gs.Pile) > 0 {
			top := gs.Pile[len(gs.Pile) - 1]
		    if rank < top.Rank && rank != 10 && rank != 2 {
				return fmt.Errorf("Card too low")
			}
		}

		// remove from hand and add to pile
		for _, idx := range idxs {
			card := player.Hand[idx]
			gs.Pile = append(gs.Pile, card)
			player.Hand = slices.Delete(player.Hand, idx, idx + 1)
		}

		// draw card after move
		if len(player.Hand) < 3 {
			need := 3 - len(player.Hand)
			for i := 0; i < need; i++ {
				if len(gs.Deck) == 0 {
					break 
				}
				player.Hand = append(player.Hand, gs.Deck[len(gs.Deck) - 1])
				gs.Deck = slices.Delete(gs.Deck, len(gs.Deck) - 1, len(gs.Deck))
				fmt.Printf("Cards left in deck: %v", len(gs.Deck))
			}
		}

		// sort hand
		sort.Slice(player.Hand, func(i, j int) bool {
			return player.Hand[i].Rank < player.Hand[j].Rank
		})

		// check if player won 
		if len(player.Hand) == 0 &&
		len(player.FaceupTableCards) == 0 &&
		len(player.FacedownTableCards) == 0 {

			last := gs.Pile[len(gs.Pile)-1] 

			isSpecialOrAce := last.Rank == 2 || last.Rank == 10 || last.Rank == 14
			if isSpecialOrAce {
				fmt.Println("You can't go out on a special card.")

				if len(gs.Pile) > 0 {
					player.Hand = append(player.Hand, gs.Pile...)
					gs.Pile = nil
				}
				return nil
			}

			fmt.Printf("Player %s won\n", player.ID)
			for i, p := range gs.Players {
				if p.ID == player.ID {
					gs.Players = slices.Delete(gs.Players, i, i+1)
					break
				}
			}
			return nil
		}	

		if rank == 10 || lastFourSame(gs) {
			gs.Pile = nil
			specialCard = true
		} else if rank == 2 {
			specialCard = true
		}
	}

	if move.Move == MoveTypePickUp {
		if len(gs.Pile) == 0 {
			return fmt.Errorf("Pile is empty, nothing to pick up")
		}
		player.Hand = append(player.Hand, gs.Pile...)
		gs.Pile = nil

		// sort hand
		sort.Slice(player.Hand, func(i, j int) bool {
			return player.Hand[i].Rank < player.Hand[j].Rank
		})
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
		fmt.Printf("Chancecard taken: %v. Cards left: %v\n", chanceCard, len(gs.Deck))

		top := gs.Pile[len(gs.Pile) - 1]

		if chanceCard.Rank < top.Rank && chanceCard.Rank != 10 && chanceCard.Rank != 2 {
			gs.Pile = append(gs.Pile, chanceCard)
			player.Hand = append(player.Hand, gs.Pile...)
			gs.Pile = nil

			// sort hand
			sort.Slice(player.Hand, func(i, j int) bool {
				return player.Hand[i].Rank < player.Hand[j].Rank
			})
			fmt.Printf("Chancecard too low, picked up pile. Cards left: %v\n", len(gs.Deck))
		} else {
			gs.Pile = append(gs.Pile, chanceCard)
			fmt.Printf("Chancecard %v played. Cards left: %v\n", chanceCard, len(gs.Deck))
			if chanceCard.Rank == 10 || lastFourSame(gs) {
				gs.Pile = nil
				specialCard = true
			} else if chanceCard.Rank == 2 {
				specialCard = true
			}
		}
	}

	if move.Move == MoveTypePlayFaceUpCard {
		// only available when hand and deck are empty
		if len(player.Hand) != 0 || len(gs.Deck) != 0 {
			return fmt.Errorf("You can't play face-up cards yet")
		}

		if move.Index == nil {
			return fmt.Errorf("missing index")
		}

		idx := *move.Index
		if idx < 0 || idx >= len(player.FaceupTableCards) {
			return fmt.Errorf("invalid index")
		}

		card := player.FaceupTableCards[idx]
		player.FaceupTableCards = slices.Delete(player.FaceupTableCards, idx, idx + 1)

		if len(gs.Pile) == 0 {
			gs.Pile = append(gs.Pile, card)
		} else {
			top := gs.Pile[len(gs.Pile)-1]
			if card.Rank < top.Rank && card.Rank != 10 && card.Rank != 2 {
				gs.Pile = append(gs.Pile, card)
				player.Hand = append(player.Hand, gs.Pile...)
				// sort hand
				sort.Slice(player.Hand, func(i, j int) bool {
					return player.Hand[i].Rank < player.Hand[j].Rank
				})
				gs.Pile = nil
			} else {
				gs.Pile = append(gs.Pile, card)
			}
		}

		if card.Rank == 10 || lastFourSame(gs) {
			gs.Pile = nil
			specialCard = true
		} else if card.Rank == 2 {
			specialCard = true
		}
	}

	if move.Move == MoveTypePlayFaceDownCard {
		if len(player.Hand) != 0 || len(player.FaceupTableCards) != 0 || len(gs.Deck) != 0 {
			return fmt.Errorf("You can't play those yet")
		}

		if move.Index == nil {
			return fmt.Errorf("missing index")
		}

		idx := *move.Index
		if idx < 0 || idx >= len(player.FacedownTableCards) {
			return fmt.Errorf("invalid index")
		}

		card := player.FacedownTableCards[idx]
		fmt.Printf("Facedown card %v played.", card)
		player.FacedownTableCards = slices.Delete(player.FacedownTableCards, idx, idx + 1)
		if len(gs.Pile) > 0 {
			top := gs.Pile[len(gs.Pile) - 1]
			if card.Rank < top.Rank && card.Rank != 10 && card.Rank != 2 {
				gs.Pile = append(gs.Pile, card)
				player.Hand = append(player.Hand, gs.Pile...)
				gs.Pile = nil
				// sort hand
				sort.Slice(player.Hand, func(i, j int) bool {
					return player.Hand[i].Rank < player.Hand[j].Rank
				})
				fmt.Println("Facedown card too low, picked up pile.")
			} else {
				gs.Pile = append(gs.Pile, card)
			}
		}

		if len(player.FacedownTableCards) == 0 {
			isSpecialOrAce := card.Rank == 2 || card.Rank == 10 || card.Rank == 14

			if isSpecialOrAce {
				fmt.Println("You can't go out on a special card.")
				if len(gs.Pile) > 0 {
					player.Hand = append(player.Hand, gs.Pile...)
					gs.Pile = nil
					return nil
				}
				return nil
			}

			fmt.Printf("Player %s won", player.ID)
			for i, p := range gs.Players {
				if player.ID == p.ID {
					gs.Players = slices.Delete(gs.Players, i, i + 1)
					break
				}
			}
			return nil
		}

		if card.Rank == 10 || lastFourSame(gs) {
			gs.Pile = nil
			specialCard = true
		} else if card.Rank == 2 {
			specialCard = true
		}
	}
	
	if !specialCard {
		advanceTurn(gs)
	}
	return nil
}

func advanceTurn(gs *GameState) {
	idx := 0
	for i, player := range gs.Players {
		if player.ID == gs.CurrentPlayer {
			idx = i
			break
		}
	}

	nextPlayer := (idx + 1) % len(gs.Players)
	gs.CurrentPlayer = gs.Players[nextPlayer].ID
}

func lastFourSame(gs *GameState) bool {
	if len(gs.Pile) < 4 {
		return false
	}

	last := gs.Pile[len(gs.Pile) - 1]
	for i := len(gs.Pile) - 4; i < len(gs.Pile); i++ {
		if gs.Pile[i].Rank != last.Rank {
			return false
		}
	}
	fmt.Printf("four in a row")
	return true
}

func applySwapFaceUp(player *PlayerState, move Move) error {
	if move.Index == nil || len(move.Indices) == 0 {
		return fmt.Errorf("missing indices")
	}
	handIndex := *move.Index
	tableIndex := move.Indices[0]

	if handIndex < 0 || handIndex >= len(player.Hand) {
		return fmt.Errorf("invalid hand index")
	}
	if tableIndex < 0 || tableIndex >= len(player.FaceupTableCards) {
		return fmt.Errorf("invalid table index")
	}

	player.Hand[handIndex], player.FaceupTableCards[tableIndex] =
	player.FaceupTableCards[tableIndex], player.Hand[handIndex]

	sort.Slice(player.Hand, func(i, j int) bool {
		return player.Hand[i].Rank < player.Hand[j].Rank
	})
	return nil
}

func applyReadySetup(gs *GameState, player *PlayerState) error {
	if len(player.FaceupTableCards) != 3 {
		return fmt.Errorf("you must have 3 face-up cards")
	}

	player.Ready = true

	allReady := true
	for i := range gs.Players {
		if !gs.Players[i].Ready {
			allReady = false
			break
		}
	}

	if allReady {
		gs.Phase = PhasePlay
	}
	return nil
}
