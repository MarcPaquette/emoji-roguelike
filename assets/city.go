package assets

// NPCDef describes a non-player character for city placement.
// Kind matches the component.NPCKind constants (0=Dialogue, 1=Healer, 2=Shop, 3=Animal).
type NPCDef struct {
	Glyph string
	Name  string
	Kind  uint8
	Lines []string
}

// ShopEntry describes one item available in Merchant Yeva's shop.
type ShopEntry struct {
	Glyph        string
	Name         string
	Price        int
	IsConsumable bool
	BonusATK     int
	BonusDEF     int
	BonusMaxHP   int
	Slot         string // "" for consumables; "head","body","feet","onehand","offhand" for equipment
}

// CityNPCs lists the named human NPCs of Emberveil.
var CityNPCs = []NPCDef{
	{
		Glyph: "🙏",
		Name:  "Sister Maris",
		Kind:  1, // NPCKindHealer
		Lines: []string{
			"May the Eternal Flame mend your wounds.",
			"Rest here, traveler. The Flame restores all.",
			"You look hale and hearty. The Flame watches over you.",
		},
	},
	{
		Glyph: "👴",
		Name:  "Father Brennan",
		Kind:  0, // NPCKindDialogue
		Lines: []string{
			"The tower has stood since before the city was a thought. We built Emberveil around it.",
			"Those who enter seeking glory rarely speak of what they find.",
			"The Flame protects this city. Beyond its walls, you are on your own.",
		},
	},
	{
		Glyph: "🍺",
		Name:  "Ol' Rudwig",
		Kind:  0, // NPCKindDialogue
		Lines: []string{
			"Last adventurer through here drank seven pints and declared herself invincible. Haven't seen her since.",
			"You look like someone who's about to do something stupid. Drink first.",
			"The ale's fresh, the floor's clean, and the stories are mostly true.",
			"Word of advice: if something glows, don't touch it. Unless it's the ale.",
		},
	},
	{
		Glyph: "⚔️",
		Name:  "Soldier Greta",
		Kind:  0, // NPCKindDialogue
		Lines: []string{
			"Don't let the glowing ones corner you. They work in groups — smarter than they look.",
			"First rule of the tower: keep the stairs behind you.",
			"I've escorted three expeditions into that place. I'm the only one who came back all three times.",
			"The deeper you go, the stranger it gets. You'll know what I mean when you see it.",
		},
	},
	{
		Glyph: "🛍️",
		Name:  "Merchant Yeva",
		Kind:  2, // NPCKindShop
		Lines: []string{
			"Welcome to Yeva's Provisions! See anything you like?",
			"Quality goods for the discerning adventurer.",
			"Fresh stock just arrived. Take a look!",
		},
	},
	{
		Glyph: "📖",
		Name:  "Scholar Alaric",
		Kind:  0, // NPCKindDialogue
		Lines: []string{
			"The tower predates the city by at least four centuries. We built around it.",
			"I've catalogued seventeen distinct energy signatures from the lower floors. Fascinating and terrifying.",
			"The inscriptions on the walls weren't made by any civilisation I've identified.",
			"Each floor seems to occupy more physical space than the exterior would suggest. Dimensional folding.",
		},
	},
	{
		Glyph: "👦",
		Name:  "Street Urchin Pip",
		Kind:  0, // NPCKindDialogue
		Lines: []string{
			"Oi! I saw a glowing thing come out of the tower once. Just for a second. Then it went back in.",
			"The soldiers don't let anyone near after dark. I sneak close sometimes. Don't tell.",
			"I bet there's treasure in there. LOADS of treasure. You should bring me some.",
			"Are you going IN? Can I come? No? Worth asking.",
		},
	},
	{
		Glyph: "👩",
		Name:  "Townsfolk Maren",
		Kind:  0, // NPCKindDialogue
		Lines: []string{
			"My husband went in there six years ago. Still waiting.",
			"The rent's cheaper on this side of town. Wonder why.",
			"Nice day, isn't it? Well. As nice as it gets near that tower.",
			"Don't stay too long down there. The city misses its heroes.",
		},
	},
	{
		Glyph: "🎣",
		Name:  "Old Fisher Bram",
		Kind:  0, // NPCKindDialogue
		Lines: []string{
			"The fish know things. They swim in from the deep channels, and they've seen what's down there.",
			"Every answer you find will come with three new questions. That's the tower's way.",
			"Things come up sometimes. Not just adventurers. Things.",
			"I fished here before the tower was famous. The fish were quieter then.",
		},
	},
	{
		Glyph: "🕊️",
		Name:  "Sister Lena",
		Kind:  0, // NPCKindDialogue
		Lines: []string{
			"The birds still nest on the tower walls. Nothing drives them away. I find that reassuring.",
			"Peace is rare in this world. Emberveil has some of it. Cherish it before you go.",
			"Travel safely, traveler. And come back. People forget to come back.",
			"The Flame burns for everyone, even those in the dark.",
		},
	},
}

// CityAnimals lists the animals of Emberveil.
// The Pigeon entry is placed multiple times in the city floor layout.
var CityAnimals = []NPCDef{
	{
		Glyph: "🐕",
		Name:  "Stray Dog",
		Kind:  3, // NPCKindAnimal
		Lines: []string{
			"A scruffy terrier trots up, tail wagging furiously. It sniffs your equipment and wanders off, apparently satisfied.",
			"The dog gives you a long, serious look, then sits down and scratches its ear.",
			"The dog rolls onto its back, paws in the air, clearly expecting attention.",
		},
	},
	{
		Glyph: "🐈",
		Name:  "Town Cat",
		Kind:  3, // NPCKindAnimal
		Lines: []string{
			"The cat stares at you with the absolute conviction that you are beneath its notice.",
			"It blinks slowly, which apparently means it trusts you. You feel oddly honoured.",
			"The cat is sitting exactly where you need to walk. It is not moving.",
		},
	},
	{
		Glyph: "🕊️",
		Name:  "Pigeon",
		Kind:  3, // NPCKindAnimal
		Lines: []string{
			"The pigeon eyes your pack suspiciously, finds nothing edible, and goes back to pecking at cobblestones.",
			"It coos at you. You coo back. An understanding is reached.",
			"The pigeon waddles sideways in a way that seems almost deliberate.",
		},
	},
	{
		Glyph: "🐓",
		Name:  "Market Hen",
		Kind:  3, // NPCKindAnimal
		Lines: []string{
			"The hen regards you with one orange eye, declares your presence unacceptable, and strides away with great dignity.",
			"It clucks loudly, as if announcing your arrival to everyone in the market.",
			"The hen is busy with something very important in the corner. You are not invited.",
		},
	},
}

// ShopCatalogue defines the items Merchant Yeva sells.
var ShopCatalogue = []ShopEntry{
	{Glyph: "🧪", Name: "Hyperflask", Price: 20, IsConsumable: true},
	{Glyph: "💎", Name: "Prism Shard", Price: 15, IsConsumable: true},
	{Glyph: "📜", Name: "Memory Scroll", Price: 18, IsConsumable: true},
	{Glyph: "🍵", Name: "Spore Draught", Price: 12, IsConsumable: true},
	{Glyph: "🪖", Name: "Crystal Helm", Price: 55, IsConsumable: false, BonusATK: 1, Slot: "head"},
	{Glyph: "🧥", Name: "Frost Weave", Price: 60, IsConsumable: false, BonusDEF: 2, Slot: "body"},
	{Glyph: "⚔️", Name: "Shard Blade", Price: 70, IsConsumable: false, BonusATK: 2, Slot: "onehand"},
	{Glyph: "🔋", Name: "Power Cell", Price: 50, IsConsumable: false, BonusDEF: 1, Slot: "offhand"},
}

// NPCScheduleEntryDef describes one period in an NPC's daily schedule.
// Behavior: 0=Stationary, 1=Wander, 2=Path.
type NPCScheduleEntryDef struct {
	StartTick                           int
	Behavior                            uint8
	BoundsX1, BoundsY1, BoundsX2, BoundsY2 int      // Wander
	Waypoints                           [][2]int     // Path: {x, y} pairs
	StandX, StandY                      int          // Stationary
}

// NPCScheduleDef holds the full schedule and movement speed for one NPC.
type NPCScheduleDef struct {
	MoveInterval int // ticks between moves (10-20 typical)
	Entries      []NPCScheduleEntryDef
}

// CityNPCSchedules maps NPC name → daily schedule. NPCs not in this map are fully static.
var CityNPCSchedules = map[string]NPCScheduleDef{
	"Merchant Yeva": {
		MoveInterval: 15,
		Entries: []NPCScheduleEntryDef{
			{StartTick: 0, Behavior: 0, StandX: 40, StandY: 47},                                          // Night: home south
			{StartTick: 1500, Behavior: 2, Waypoints: [][2]int{{40, 44}, {44, 44}, {52, 44}, {49, 47}}},   // Morning: path to shop
			{StartTick: 2500, Behavior: 1, BoundsX1: 46, BoundsY1: 45, BoundsX2: 53, BoundsY2: 49},       // Day: wander shop
			{StartTick: 5000, Behavior: 2, Waypoints: [][2]int{{52, 44}, {44, 44}, {40, 44}, {40, 47}}},   // Evening: path home
		},
	},
	"Ol' Rudwig": {
		MoveInterval: 20,
		Entries: []NPCScheduleEntryDef{
			{StartTick: 0, Behavior: 0, StandX: 4, StandY: 7},                                       // Night: tavern seat
			{StartTick: 1500, Behavior: 1, BoundsX1: 2, BoundsY1: 3, BoundsX2: 6, BoundsY2: 11},     // Morning: wander tavern
			{StartTick: 2500, Behavior: 1, BoundsX1: 2, BoundsY1: 3, BoundsX2: 6, BoundsY2: 11},     // Day: wander tavern
			{StartTick: 5000, Behavior: 1, BoundsX1: 2, BoundsY1: 3, BoundsX2: 6, BoundsY2: 11},     // Evening: wander tavern
		},
	},
	"Soldier Greta": {
		MoveInterval: 14,
		Entries: []NPCScheduleEntryDef{
			{StartTick: 0, Behavior: 0, StandX: 51, StandY: 34},                                     // Night: guard post
			{StartTick: 1500, Behavior: 2, Waypoints: [][2]int{{52, 33}}},                            // Morning: path to tower door
			{StartTick: 2500, Behavior: 1, BoundsX1: 48, BoundsY1: 32, BoundsX2: 56, BoundsY2: 36},  // Day: patrol square
			{StartTick: 5000, Behavior: 2, Waypoints: [][2]int{{51, 34}}},                            // Evening: return to post
		},
	},
	"Scholar Alaric": {
		MoveInterval: 18,
		Entries: []NPCScheduleEntryDef{
			{StartTick: 0, Behavior: 0, StandX: 60, StandY: 30},                                            // Night: square east
			{StartTick: 1500, Behavior: 2, Waypoints: [][2]int{{60, 22}, {47, 22}, {47, 14}, {47, 7}}},     // Morning: path to church
			{StartTick: 2500, Behavior: 1, BoundsX1: 36, BoundsY1: 3, BoundsX2: 57, BoundsY2: 8},          // Day: wander nave
			{StartTick: 5000, Behavior: 2, Waypoints: [][2]int{{47, 14}, {47, 22}, {60, 22}, {60, 30}}},    // Evening: path home
		},
	},
	"Father Brennan": {
		MoveInterval: 20,
		Entries: []NPCScheduleEntryDef{
			{StartTick: 0, Behavior: 0, StandX: 50, StandY: 4},                                     // Night: altar
			{StartTick: 1500, Behavior: 1, BoundsX1: 36, BoundsY1: 3, BoundsX2: 57, BoundsY2: 8},   // Morning: wander nave
			{StartTick: 2500, Behavior: 1, BoundsX1: 36, BoundsY1: 3, BoundsX2: 57, BoundsY2: 8},   // Day: wander nave
			{StartTick: 5000, Behavior: 1, BoundsX1: 36, BoundsY1: 3, BoundsX2: 57, BoundsY2: 8},   // Evening: wander nave
		},
	},
	"Sister Maris": {
		MoveInterval: 18,
		Entries: []NPCScheduleEntryDef{
			{StartTick: 0, Behavior: 0, StandX: 45, StandY: 5},                                     // Night: nave seat
			{StartTick: 1500, Behavior: 1, BoundsX1: 36, BoundsY1: 3, BoundsX2: 57, BoundsY2: 8},   // Morning: wander nave
			{StartTick: 2500, Behavior: 1, BoundsX1: 36, BoundsY1: 3, BoundsX2: 57, BoundsY2: 8},   // Day: wander nave
			{StartTick: 5000, Behavior: 1, BoundsX1: 36, BoundsY1: 3, BoundsX2: 57, BoundsY2: 8},   // Evening: wander nave
		},
	},
	"Sister Lena": {
		MoveInterval: 18,
		Entries: []NPCScheduleEntryDef{
			{StartTick: 0, Behavior: 0, StandX: 56, StandY: 11},                                     // Night: vestry
			{StartTick: 1500, Behavior: 1, BoundsX1: 54, BoundsY1: 10, BoundsX2: 57, BoundsY2: 13},  // Morning: wander vestry
			{StartTick: 2500, Behavior: 1, BoundsX1: 54, BoundsY1: 10, BoundsX2: 57, BoundsY2: 13},  // Day: wander vestry
			{StartTick: 5000, Behavior: 1, BoundsX1: 54, BoundsY1: 10, BoundsX2: 57, BoundsY2: 13},  // Evening: wander vestry
		},
	},
	"Townsfolk Maren": {
		MoveInterval: 16,
		Entries: []NPCScheduleEntryDef{
			{StartTick: 0, Behavior: 0, StandX: 40, StandY: 47},                                                    // Night: home south
			{StartTick: 1500, Behavior: 1, BoundsX1: 38, BoundsY1: 45, BoundsX2: 42, BoundsY2: 49},                // Morning: wander home
			{StartTick: 2500, Behavior: 2, Waypoints: [][2]int{{40, 44}, {40, 36}, {50, 30}}},                      // Day: path to square
			{StartTick: 5000, Behavior: 2, Waypoints: [][2]int{{40, 36}, {40, 44}, {40, 47}}},                      // Evening: path home
		},
	},
	"Street Urchin Pip": {
		MoveInterval: 12,
		Entries: []NPCScheduleEntryDef{
			{StartTick: 0, Behavior: 0, StandX: 4, StandY: 47},                                              // Night: market
			{StartTick: 1500, Behavior: 2, Waypoints: [][2]int{{6, 44}, {13, 43}, {13, 36}, {40, 30}}},      // Morning: path to square
			{StartTick: 2500, Behavior: 1, BoundsX1: 28, BoundsY1: 23, BoundsX2: 72, BoundsY2: 36},         // Day: wander square
			{StartTick: 5000, Behavior: 2, Waypoints: [][2]int{{13, 36}, {13, 43}, {6, 44}, {4, 47}}},       // Evening: path to market
		},
	},
	"Old Fisher Bram": {
		MoveInterval: 20,
		Entries: []NPCScheduleEntryDef{
			{StartTick: 0, Behavior: 0, StandX: 4, StandY: 49},                                     // Night: market stall
			{StartTick: 1500, Behavior: 1, BoundsX1: 2, BoundsY1: 45, BoundsX2: 11, BoundsY2: 49},  // Morning: wander market
			{StartTick: 2500, Behavior: 1, BoundsX1: 2, BoundsY1: 45, BoundsX2: 11, BoundsY2: 49},  // Day: wander market
			{StartTick: 5000, Behavior: 1, BoundsX1: 2, BoundsY1: 45, BoundsX2: 11, BoundsY2: 49},  // Evening: wander market
		},
	},
	"Stray Dog": {
		MoveInterval: 10,
		Entries: []NPCScheduleEntryDef{
			{StartTick: 0, Behavior: 1, BoundsX1: 28, BoundsY1: 23, BoundsX2: 72, BoundsY2: 36}, // Always wander square
		},
	},
	"Town Cat": {
		MoveInterval: 14,
		Entries: []NPCScheduleEntryDef{
			{StartTick: 0, Behavior: 0, StandX: 60, StandY: 22},                                     // Night: stationary
			{StartTick: 1500, Behavior: 1, BoundsX1: 28, BoundsY1: 20, BoundsX2: 72, BoundsY2: 24},  // Morning: wander streets
			{StartTick: 2500, Behavior: 1, BoundsX1: 28, BoundsY1: 20, BoundsX2: 72, BoundsY2: 24},  // Day: wander streets
			{StartTick: 5000, Behavior: 0, StandX: 60, StandY: 22},                                   // Evening: stationary
		},
	},
	"Pigeon": {
		MoveInterval: 8,
		Entries: []NPCScheduleEntryDef{
			{StartTick: 0, Behavior: 0, StandX: -1, StandY: -1},                                      // Night: stationary (stand pos set per-instance)
			{StartTick: 1500, Behavior: 1, BoundsX1: 28, BoundsY1: 23, BoundsX2: 72, BoundsY2: 36},  // Morning: wander square
			{StartTick: 2500, Behavior: 1, BoundsX1: 28, BoundsY1: 23, BoundsX2: 72, BoundsY2: 36},  // Day: wander square
			{StartTick: 5000, Behavior: 0, StandX: -1, StandY: -1},                                    // Evening: stationary
		},
	},
	"Market Hen": {
		MoveInterval: 12,
		Entries: []NPCScheduleEntryDef{
			{StartTick: 0, Behavior: 0, StandX: 8, StandY: 43},                                     // Night: stationary
			{StartTick: 1500, Behavior: 1, BoundsX1: 2, BoundsY1: 43, BoundsX2: 11, BoundsY2: 49},  // Morning: wander market
			{StartTick: 2500, Behavior: 1, BoundsX1: 2, BoundsY1: 43, BoundsX2: 11, BoundsY2: 49},  // Day: wander market
			{StartTick: 5000, Behavior: 1, BoundsX1: 2, BoundsY1: 43, BoundsX2: 11, BoundsY2: 49},  // Evening: wander market
		},
	},
}

// CityInscriptions is the pool of inscription texts placed on city walls and signs.
var CityInscriptions = []string{
	"The Tower of Emberveil stands where it has always stood, older than the city itself.",
	"NOTICE: Citizens are advised to avoid the tower entrance after sundown. — Captain Holt",
	"Wanted: Brave adventurers to explore the tower depths. Generous compensation. Apply within.",
	"Fresh supplies daily at Yeva's Provisions — SE quarter.",
	"Travelers welcome at The Sunken Flagon. Warmth, ale, and terrible advice.",
	"In memory of those who descended and did not return. Their courage is not forgotten.",
}
