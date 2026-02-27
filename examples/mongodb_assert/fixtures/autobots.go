package fixtures

// ---------- Autobots (Heroes) ----------

// Bumblelf = Bumblebee + Elf
type Bumblelf struct {
	DnDBase        `       bson:",inline"`
	StealthLevel   uint8  `bson:"stealth_level"   json:"stealth_level"`
	FavoredTerrain string `bson:"favored_terrain" json:"favored_terrain"`
}

func NewBumblelf() *Bumblelf {
	return &Bumblelf{
		DnDBase:        DnDBase{EntityName: "Bumblelf", ClassName: "Elf Scout"},
		StealthLevel:   14,
		FavoredTerrain: "Forest",
	}
}

// OptimadinPrime = Optimus Prime + Paladin
type OptimadinPrime struct {
	DnDBase        `       bson:",inline"`
	OathType       string `bson:"oath_type"       json:"oath_type"`
	LeadershipAura string `bson:"leadership_aura" json:"leadership_aura"`
}

func NewOptimadinPrime() *OptimadinPrime {
	return &OptimadinPrime{
		DnDBase:        DnDBase{EntityName: "OptimadinPrime", ClassName: "Paladin Commander"},
		OathType:       "Oath of Devotion",
		LeadershipAura: "Courageous Presence",
	}
}

// Ironknight = Ironhide + Knight
type Ironknight struct {
	DnDBase     `       bson:",inline"`
	ShieldType  string `bson:"shield_type"  json:"shield_type"`
	ArmorWeight uint32 `bson:"armor_weight" json:"armor_weight"`
}

func NewIronknight() *Ironknight {
	return &Ironknight{
		DnDBase:     DnDBase{EntityName: "Ironknight", ClassName: "Armored Knight"},
		ShieldType:  "Tower Shield",
		ArmorWeight: 85,
	}
}

// Ratcheric = Ratchet + Cleric
type Ratcheric struct {
	DnDBase      `       bson:",inline"`
	HealingPower uint16 `bson:"healing_power" json:"healing_power"`
	Deity        string `bson:"deity"         json:"deity"`
}

func NewRatcheric() *Ratcheric {
	return &Ratcheric{
		DnDBase:      DnDBase{EntityName: "Ratcheric", ClassName: "Battle Cleric"},
		HealingPower: 120,
		Deity:        "Primus",
	}
}

// Jazogue = Jazz + Rogue
type Jazogue struct {
	DnDBase         `       bson:",inline"`
	SneakAttackDice uint8  `bson:"sneak_attack_dice" json:"sneak_attack_dice"`
	Specialty       string `bson:"specialty"         json:"specialty"`
}

func NewJazogue() *Jazogue {
	return &Jazogue{
		DnDBase:         DnDBase{EntityName: "Jazogue", ClassName: "Agile Rogue"},
		SneakAttackDice: 6,
		Specialty:       "Infiltration",
	}
}

// Wheelificer = Wheeljack + Artificer
type Wheelificer struct {
	DnDBase      `         bson:",inline"`
	Inventions   []string `bson:"inventions"    json:"inventions"`
	WorkshopName string   `bson:"workshop_name" json:"workshop_name"`
}

func NewWheelificer() *Wheelificer {
	return &Wheelificer{
		DnDBase:      DnDBase{EntityName: "Wheelificer", ClassName: "Arcane Artificer"},
		Inventions:   []string{"Energon Converter", "Dimensional Anchor", "Sonic Wrench"},
		WorkshopName: "The Spark Forge",
	}
}
