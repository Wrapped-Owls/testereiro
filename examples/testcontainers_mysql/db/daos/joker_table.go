package daos

type Joker struct {
	ID     int    `db:"id"`
	Name   string `db:"name"`
	Effect string `db:"effect"`
	Rarity string `db:"rarity"`
}
