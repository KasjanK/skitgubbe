package game

type Rank int
type Suit int

const (
	SuitHearts Suit = iota
	SuitSpades
	SuitDiamonds
	SuitClubs
)

const (
	RankTwo   Rank = 2
	RankThree Rank = 3
	RankFour  Rank = 4
	RankFive  Rank = 5
	RankSix   Rank = 6
	RankSeven Rank = 7
	RankEight Rank = 8
	RankNine  Rank = 9
	RankTen   Rank = 10
	RankJ     Rank = 11
	RankQ     Rank = 12
	RankK     Rank = 13
	RankA     Rank = 14
)

type Card struct {
	Rank Rank `json:"rank"`
	Suit Suit `json:"suit"`
}

type PlayerID string

type PlayerState struct {
	ID 	               PlayerID `json:"id"`
	Hand               []Card   `json:"hand"`
	FaceupTableCards   []Card	`json:"faceup_table_cards"` 
	FacedownTableCards []Card   `json:"facedown_table_cards"`
}

type GameState struct {
	ID 			  string		`json:"id"`
	Players 	  []PlayerState `json:"players"`
	CurrentPlayer PlayerID		`json:"current_player"`
	Pile 		  []Card		`json:"pile"`
}
