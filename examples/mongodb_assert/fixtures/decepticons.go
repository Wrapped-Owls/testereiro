package fixtures

// ---------- Decepticons (Villains) ----------

// MegadwarfTron = Megatron + Dwarf
type MegadwarfTron struct {
	DnDBase        `       bson:",inline"`
	ArmorThickness uint32 `bson:"armor_thickness" json:"armor_thickness"`
	BattleCry      string `bson:"battle_cry"      json:"battle_cry"`
}

func NewMegadwarfTron() *MegadwarfTron {
	return &MegadwarfTron{
		DnDBase:        DnDBase{EntityName: "MegadwarfTron", ClassName: "Dwarven Warlord"},
		ArmorThickness: 200,
		BattleCry:      "Peace through tyranny!",
	}
}

// Sorcerscream = Starscream + Sorcerer
type Sorcerscream struct {
	DnDBase       `       bson:",inline"`
	SpellPower    uint16 `bson:"spell_power"     json:"spell_power"`
	SchoolOfMagic string `bson:"school_of_magic" json:"school_of_magic"`
}

func NewSorcerscream() *Sorcerscream {
	return &Sorcerscream{
		DnDBase:       DnDBase{EntityName: "Sorcerscream", ClassName: "Screaming Sorcerer"},
		SpellPower:    95,
		SchoolOfMagic: "Evocation",
	}
}

// Soundbard = Soundwave + Bard
type Soundbard struct {
	DnDBase     `       bson:",inline"`
	Instrument  string `bson:"instrument"    json:"instrument"`
	SongOfPower string `bson:"song_of_power" json:"song_of_power"`
}

func NewSoundbard() *Soundbard {
	return &Soundbard{
		DnDBase:     DnDBase{EntityName: "Soundbard", ClassName: "Shadow Bard"},
		Instrument:  "Echoing Lute",
		SongOfPower: "Dirge of Domination",
	}
}

// Shocklock = Shockwave + Warlock
type Shocklock struct {
	DnDBase        `       bson:",inline"`
	PatronName     string `bson:"patron_name"     json:"patron_name"`
	EldritchBlasts uint8  `bson:"eldritch_blasts" json:"eldritch_blasts"`
}

func NewShocklock() *Shocklock {
	return &Shocklock{
		DnDBase:        DnDBase{EntityName: "Shocklock", ClassName: "Eldritch Warlock"},
		PatronName:     "Unicron the Devourer",
		EldritchBlasts: 4,
	}
}
