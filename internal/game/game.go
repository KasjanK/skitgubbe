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

var AllSuits = []Suit{
    SuitClubs,
    SuitDiamonds,
    SuitHearts,
    SuitSpades,
}

var AllRanks = []Rank{
	RankTwo,
    RankThree,
    RankFour,
    RankFive,
    RankSix,
    RankSeven,
    RankEight,
    RankNine,
    RankTen,
    RankJ,
    RankQ,
    RankK,
    RankA,
}

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
	Deck 		  []Card 		`json:"deck"`
	Pile 		  []Card		`json:"pile"`
}

type Room struct {
	ID 	 	string            `json:"id"`
	OwnerID PlayerID		  `json:"owner_id"`
	Players []PlayerState     `json:"players"`
	Ready   map[PlayerID]bool `json:"ready"`
	Started bool              `json:"started"`
	GameID string             `json:"game_id"`
}

type VisiblePlayer struct {
	ID 		 PlayerID `json:"id"`
	HandSize int 	  `json:"hand_size"`
}

type VisibleState struct {
	ID 			 string      	 `json:"id"`
	You 		 PlayerState     `json:"you"`
	Others       []VisiblePlayer `json:"others"`
	Pile 		 []Card			 `json:"pile"`
	CurrentPlayer PlayerID		 `json:"current_player"`
}

type MoveType string

const (
	MoveTypePlayCard 		 MoveType = "play_card"
	MoveTypePickUp   		 MoveType = "pickup"
	MoveTypeChance   	     MoveType = "chance"
	MoveTypePlayFaceUpCard   MoveType = "play_face_up"
	MoveTypePlayFaceDownCard MoveType = "play_face_down"
)

type Move struct {
	Move  MoveType `json:"type"`
	Card  *Card    `json:"card,omitempty"`
	Index *int     `json:"index,omitempty"` 
}
