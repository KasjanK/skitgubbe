package game

import "math/rand"

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
	length := len(deck)
	for i := 0; i < length; i++ {
		r := i + rand.Intn(length - i)
		deck[i], deck[r] = deck[r], deck[i]
	}
}
