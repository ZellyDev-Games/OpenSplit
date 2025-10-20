package qusb2snes

import (
	"fmt"
	"math"
	"sync"
	"time"
)

var (
	roomIDEnum = map[string]uint32{
		"landingSite":                     0x91F8,
		"crateriaPowerBombRoom":           0x93AA,
		"westOcean":                       0x93FE,
		"elevatorToMaridia":               0x94CC,
		"crateriaMoat":                    0x95FF,
		"elevatorToCaterpillar":           0x962A,
		"gauntletETankRoom":               0x965B,
		"climb":                           0x96BA,
		"pitRoom":                         0x975C,
		"elevatorToMorphBall":             0x97B5,
		"bombTorizo":                      0x9804,
		"terminator":                      0x990D,
		"elevatorToGreenBrinstar":         0x9938,
		"greenPirateShaft":                0x99BD,
		"crateriaSupersRoom":              0x99F9,
		"theFinalMissile":                 0x9A90,
		"greenBrinstarMainShaft":          0x9AD9,
		"sporeSpawnSuper":                 0x9B5B,
		"earlySupers":                     0x9BC8,
		"brinstarReserveRoom":             0x9C07,
		"bigPink":                         0x9D19,
		"sporeSpawnKeyhunter":             0x9D9C,
		"sporeSpawn":                      0x9DC7,
		"pinkBrinstarPowerBombRoom":       0x9E11,
		"greenHills":                      0x9E52,
		"noobBridge":                      0x9FBA,
		"morphBall":                       0x9E9F,
		"blueBrinstarETankRoom":           0x9F64,
		"etecoonETankRoom":                0xA011,
		"etecoonSuperRoom":                0xA051,
		"waterway":                        0xA0D2,
		"alphaMissileRoom":                0xA107,
		"hopperETankRoom":                 0xA15B,
		"billyMays":                       0xA1D8,
		"redTower":                        0xA253,
		"xRay":                            0xA2CE,
		"caterpillar":                     0xA322,
		"betaPowerBombRoom":               0xA37C,
		"alphaPowerBombsRoom":             0xA3AE,
		"bat":                             0xA3DD,
		"spazer":                          0xA447,
		"warehouseETankRoom":              0xA4B1,
		"warehouseZeela":                  0xA471,
		"warehouseKiHunters":              0xA4DA,
		"kraidEyeDoor":                    0xA56B,
		"kraid":                           0xA59F,
		"statuesHallway":                  0xA5ED,
		"statues":                         0xA66A,
		"warehouseEntrance":               0xA6A1,
		"varia":                           0xA6E2,
		"cathedral":                       0xA788,
		"businessCenter":                  0xA7DE,
		"iceBeam":                         0xA890,
		"crumbleShaft":                    0xA8F8,
		"crocomireSpeedway":               0xA923,
		"crocomire":                       0xA98D,
		"hiJump":                          0xA9E5,
		"crocomireEscape":                 0xAA0E,
		"hiJumpShaft":                     0xAA41,
		"postCrocomirePowerBombRoom":      0xAADE,
		"cosineRoom":                      0xAB3B,
		"preGrapple":                      0xAB8F,
		"grapple":                         0xAC2B,
		"norfairReserveRoom":              0xAC5A,
		"greenBubblesRoom":                0xAC83,
		"bubbleMountain":                  0xACB3,
		"speedBoostHall":                  0xACF0,
		"speedBooster":                    0xAD1B,
		"singleChamber":                   0xAD5E, // Exit room from Lower Norfair, also on the path to Wave
		"doubleChamber":                   0xADAD,
		"waveBeam":                        0xADDE,
		"volcano":                         0xAE32,
		"kronicBoost":                     0xAE74,
		"magdolliteTunnel":                0xAEB4,
		"lowerNorfairElevator":            0xAF3F,
		"risingTide":                      0xAFA3,
		"spikyAcidSnakes":                 0xAFFB,
		"acidStatue":                      0xB1E5,
		"mainHall":                        0xB236, // First room in Lower Norfair
		"goldenTorizo":                    0xB283,
		"ridley":                          0xB32E,
		"lowerNorfairFarming":             0xB37A,
		"mickeyMouse":                     0xB40A,
		"pillars":                         0xB457,
		"writg":                           0xB4AD,
		"amphitheatre":                    0xB4E5,
		"lowerNorfairSpringMaze":          0xB510,
		"lowerNorfairEscapePowerBombRoom": 0xB55A,
		"redKiShaft":                      0xB585,
		"wasteland":                       0xB5D5,
		"metalPirates":                    0xB62B,
		"threeMusketeers":                 0xB656,
		"ridleyETankRoom":                 0xB698,
		"screwAttack":                     0xB6C1,
		"lowerNorfairFireflea":            0xB6EE,
		"bowling":                         0xC98E,
		"wreckedShipEntrance":             0xCA08,
		"attic":                           0xCA52,
		"atticWorkerRobotRoom":            0xCAAE,
		"wreckedShipMainShaft":            0xCAF6,
		"wreckedShipETankRoom":            0xCC27,
		"basement":                        0xCC6F, // Basement of Wrecked Ship
		"phantoon":                        0xCD13,
		"wreckedShipLeftSuperRoom":        0xCDA8,
		"wreckedShipRightSuperRoom":       0xCDF1,
		"gravity":                         0xCE40,
		"glassTunnel":                     0xCEFB,
		"mainStreet":                      0xCFC9,
		"mamaTurtle":                      0xD055,
		"wateringHole":                    0xD13B,
		"beach":                           0xD1DD,
		"plasmaBeam":                      0xD2AA,
		"maridiaElevator":                 0xD30B,
		"plasmaSpark":                     0xD340,
		"toiletBowl":                      0xD408,
		"oasis":                           0xD48E,
		"leftSandPit":                     0xD4EF,
		"rightSandPit":                    0xD51E,
		"aqueduct":                        0xD5A7,
		"butterflyRoom":                   0xD5EC,
		"botwoonHallway":                  0xD617,
		"springBall":                      0xD6D0,
		"precious":                        0xD78F,
		"botwoonETankRoom":                0xD7E4,
		"botwoon":                         0xD95E,
		"spaceJump":                       0xD9AA,
		"westCactusAlley":                 0xD9FE,
		"draygon":                         0xDA60,
		"tourianElevator":                 0xDAAE,
		"metroidOne":                      0xDAE1,
		"metroidTwo":                      0xDB31,
		"metroidThree":                    0xDB7D,
		"metroidFour":                     0xDBCD,
		"dustTorizo":                      0xDC65,
		"tourianHopper":                   0xDC19,
		"tourianEyeDoor":                  0xDDC4,
		"bigBoy":                          0xDCB1,
		"motherBrain":                     0xDD58,
		"rinkaShaft":                      0xDDF3,
		"tourianEscape4":                  0xDEDE,
		"ceresElevator":                   0xDF45,
		"flatRoom":                        0xE06B, // Placeholder name for the flat room in Ceres Station
		"ceresRidley":                     0xE0B5,
	}
	mapInUseEnum = map[string]uint32{
		"crateria":    0x0,
		"brinstar":    0x1,
		"norfair":     0x2,
		"wreckedShip": 0x3,
		"maridia":     0x4,
		"tourian":     0x5,
		"ceres":       0x6,
	}
	gameStateEnum = map[string]uint32{
		"normalGameplay":       0x8,
		"doorTransition":       0xB,
		"startOfCeresCutscene": 0x20,
		"preEndCutscene":       0x26, // briefly at this value during the black screen transition after the ship fades out
		"endCutscene":          0x27,
	}
	unlockFlagEnum = map[string]uint32{
		// First item byte
		"variaSuit":   0x1,
		"springBall":  0x2,
		"morphBall":   0x4,
		"screwAttack": 0x8,
		"gravSuit":    0x20,
		// Second item byte
		"hiJump":       0x1,
		"spaceJump":    0x2,
		"bomb":         0x10,
		"speedBooster": 0x20,
		"grapple":      0x40,
		"xray":         0x80,
		// Beams
		"wave":   0x1,
		"ice":    0x2,
		"spazer": 0x4,
		"plasma": 0x8,
		// Charge
		"chargeBeam": 0x10,
	}
	motherBrainMaxHPEnum = map[string]uint32{
		"phase1": 0xBB8,  // 3000
		"phase2": 0x4650, // 18000
		"phase3": 0x8CA0, // 36000
	}
	eventFlagEnum = map[string]uint32{
		"zebesAblaze": 0x40,
		"tubeBroken":  0x8,
	}
	bossFlagEnum = map[string]uint32{
		// Crateria
		"bombTorizo": 0x4,
		// Brinstar
		"sporeSpawn": 0x2,
		"kraid":      0x1,
		// Norfair
		"ridley":       0x1,
		"crocomire":    0x2,
		"goldenTorizo": 0x4,
		// Wrecked Ship
		"phantoon": 0x1,
		// Maridia
		"draygon": 0x1,
		"botwoon": 0x2,
		// Tourian
		"motherBrain": 0x2,
		// Ceres
		"ceresRidley": 0x1,
	}
)

type Settings struct {
	data map[string]struct {
		value  bool
		parent *string
	}
	modifiedAfterCreation bool
	mu                    sync.RWMutex
}

func NewSettings() *Settings {
	s := &Settings{
		data: make(map[string]struct {
			value  bool
			parent *string
		}),
		modifiedAfterCreation: false,
	}
	// Split on Missiles, Super Missiles, and Power Bombs
	s.Insert("ammoPickups", true)
	// Split on the first Missile pickup
	s.InsertWithParent("firstMissile", false, "ammoPickups")
	// Split on each Missile upgrade
	s.InsertWithParent("allMissiles", false, "ammoPickups")
	// Split on specific Missile Pack locations
	s.InsertWithParent("specificMissiles", false, "ammoPickups")
	// Split on Crateria Missile Pack locations
	s.InsertWithParent("crateriaMissiles", false, "specificMissiles")
	// Split on picking up the Missile Pack located at the bottom left of the West Ocean
	s.InsertWithParent("oceanBottomMissiles", false, "crateriaMissiles")
	// Split on picking up the Missile Pack located in the ceiling tile in West Ocean
	s.InsertWithParent("oceanTopMissiles", false, "crateriaMissiles")
	// Split on picking up the Missile Pack located in the Morphball maze section of West Ocean
	s.InsertWithParent("oceanMiddleMissiles", false, "crateriaMissiles")
	// Split on picking up the Missile Pack in The Moat, also known as The Lake
	s.InsertWithParent("moatMissiles", false, "crateriaMissiles")
	// Split on picking up the Missile Pack in the Pit Room
	s.InsertWithParent("oldTourianMissiles", false, "crateriaMissiles")
	// Split on picking up the right side Missile Pack at the end of Gauntlet(Green Pirates Shaft)
	s.InsertWithParent("gauntletRightMissiles", false, "crateriaMissiles")
	// Split on picking up the left side Missile Pack at the end of Gauntlet(Green Pirates Shaft)
	s.InsertWithParent("gauntletLeftMissiles", false, "crateriaMissiles")
	// Split on picking up the Missile Pack located in The Final Missile
	s.InsertWithParent("dentalPlan", false, "crateriaMissiles")
	// Split on Brinstar Missile Pack locations
	s.InsertWithParent("brinstarMissiles", false, "specificMissiles")
	// Split on picking up the Missile Pack located below the crumble bridge in the Early Supers Room
	s.InsertWithParent("earlySuperBridgeMissiles", false, "brinstarMissiles")
	// Split on picking up the first Missile Pack behind the Brinstar Reserve Tank
	s.InsertWithParent("greenBrinstarReserveMissiles", false, "brinstarMissiles")
	// Split on picking up the second Missile Pack behind the Brinstar Reserve Tank Room
	s.InsertWithParent("greenBrinstarExtraReserveMissiles", false, "brinstarMissiles")
	// Split on picking up the Missile Pack located left of center in Big Pink
	s.InsertWithParent("bigPinkTopMissiles", false, "brinstarMissiles")
	// Split on picking up the Missile Pack located at the bottom left of Big Pink
	s.InsertWithParent("chargeMissiles", false, "brinstarMissiles")
	// Split on picking up the Missile Pack in Green Hill Zone
	s.InsertWithParent("greenHillsMissiles", false, "brinstarMissiles")
	// Split on picking up the Missile Pack in the Blue Brinstar Energy Tank Room
	s.InsertWithParent("blueBrinstarETankMissiles", false, "brinstarMissiles")
	// Split on picking up the first Missile Pack of the game(First Missile Room)
	s.InsertWithParent("alphaMissiles", false, "brinstarMissiles")
	// Split on picking up the Missile Pack located on the pedestal in Billy Mays' Room
	s.InsertWithParent("billyMaysMissiles", false, "brinstarMissiles")
	// Split on picking up the Missile Pack located in the floor of Billy Mays' Room
	s.InsertWithParent("butWaitTheresMoreMissiles", false, "brinstarMissiles")
	// Split on picking up the Missile Pack in the Alpha Power Bombs Room
	s.InsertWithParent("redBrinstarMissiles", false, "brinstarMissiles")
	// Split on picking up the Missile Pack in the Warehouse Kihunter Room
	s.InsertWithParent("warehouseMissiles", false, "brinstarMissiles")
	// Split on Norfair Missile Pack locations
	s.InsertWithParent("norfairMissiles", false, "specificMissiles")
	// Split on picking up the Missile Pack in Cathedral
	s.InsertWithParent("cathedralMissiles", false, "norfairMissiles")
	// Split on picking up the Missile Pack in Crumble Shaft
	s.InsertWithParent("crumbleShaftMissiles", false, "norfairMissiles")
	// Split on picking up the Missile Pack in Crocomire Escape
	s.InsertWithParent("crocomireEscapeMissiles", false, "norfairMissiles")
	// Split on picking up the Missile Pack in the Hi Jump Energy Tank Room
	s.InsertWithParent("hiJumpMissiles", false, "norfairMissiles")
	// Split on picking up the Missile Pack in the Post Crocomire Missile Room, also known as Cosine Room
	s.InsertWithParent("postCrocomireMissiles", false, "norfairMissiles")
	// Split on picking up the Missile Pack in the Post Crocomire Jump Room
	s.InsertWithParent("grappleMissiles", false, "norfairMissiles")
	// Split on picking up the Missile Pack in the Norfair Reserve Tank Room
	s.InsertWithParent("norfairReserveMissiles", false, "norfairMissiles")
	// Split on picking up the Missile Pack in the Green Bubbles Missile Room
	s.InsertWithParent("greenBubblesMissiles", false, "norfairMissiles")
	// Split on picking up the Missile Pack in Bubble Mountain
	s.InsertWithParent("bubbleMountainMissiles", false, "norfairMissiles")
	// Split on picking up the Missile Pack in Speed Booster Hall
	s.InsertWithParent("speedBoostMissiles", false, "norfairMissiles")
	// Split on picking up the Wave Missile Pack in Double Chamber
	s.InsertWithParent("waveMissiles", false, "norfairMissiles")
	// Split on picking up the Missile Pack in the Golden Torizo's Room
	s.InsertWithParent("goldTorizoMissiles", false, "norfairMissiles")
	// Split on picking up the Missile Pack in the Mickey Mouse Room
	s.InsertWithParent("mickeyMouseMissiles", false, "norfairMissiles")
	// Split on picking up the Missile Pack in the Lower Norfair Springball Maze Room
	s.InsertWithParent("lowerNorfairSpringMazeMissiles", false, "norfairMissiles")
	// Split on picking up the Missile Pack in the The Musketeers' Room
	s.InsertWithParent("threeMusketeersMissiles", false, "norfairMissiles")
	// Split on Wrecked Ship Missile Pack locations
	s.InsertWithParent("wreckedShipMissiles", false, "specificMissiles")
	// Split on picking up the Missile Pack in Wrecked Ship Main Shaft
	s.InsertWithParent("wreckedShipMainShaftMissiles", false, "wreckedShipMissiles")
	// Split on picking up the Missile Pack in Bowling Alley
	s.InsertWithParent("bowlingMissiles", false, "wreckedShipMissiles")
	// Split on picking up the Missile Pack in the Wrecked Ship East Missile Room
	s.InsertWithParent("atticMissiles", false, "wreckedShipMissiles")
	// Split on Maridia Missile Pack locations
	s.InsertWithParent("maridiaMissiles", false, "specificMissiles")
	// Split on picking up the Missile Pack in Main Street
	s.InsertWithParent("mainStreetMissiles", false, "maridiaMissiles")
	// Split on picking up the Missile Pack in the Mama Turtle Room
	s.InsertWithParent("mamaTurtleMissiles", false, "maridiaMissiles")
	// Split on picking up the Missile Pack in Watering Hole
	s.InsertWithParent("wateringHoleMissiles", false, "maridiaMissiles")
	// Split on picking up the Missile Pack in the Pseudo Plasma Spark Room
	s.InsertWithParent("beachMissiles", false, "maridiaMissiles")
	// Split on picking up the Missile Pack in West Sand Hole
	s.InsertWithParent("leftSandPitMissiles", false, "maridiaMissiles")
	// Split on picking up the Missile Pack in East Sand Hole
	s.InsertWithParent("rightSandPitMissiles", false, "maridiaMissiles")
	// Split on picking up the Missile Pack in Aqueduct
	s.InsertWithParent("aqueductMissiles", false, "maridiaMissiles")
	// Split on picking up the Missile Pack in The Precious Room
	s.InsertWithParent("preDraygonMissiles", false, "maridiaMissiles")
	// Split on the first Super Missile pickup
	s.InsertWithParent("firstSuper", false, "ammoPickups")
	// Split on each Super Missile upgrade
	s.InsertWithParent("allSupers", false, "ammoPickups")
	// Split on specific Super Missile Pack locations
	s.InsertWithParent("specificSupers", false, "ammoPickups")
	// Split on picking up the Super Missile Pack in the Crateria Super Room
	s.InsertWithParent("climbSupers", false, "specificSupers")
	// Split on picking up the Super Missile Pack in the Spore Spawn Super Room (NOTE: SSTRA splits when the dialogue box disappears, not on touch. Use Spore Spawn RTA Finish for SSTRA runs.)
	s.InsertWithParent("sporeSpawnSupers", false, "specificSupers")
	// Split on picking up the Super Missile Pack in the Early Supers Room
	s.InsertWithParent("earlySupers", false, "specificSupers")
	// Split on picking up the Super Missile Pack in the Etecoon Super Room
	s.InsertWithParent("etecoonSupers", false, "specificSupers")
	// Split on picking up the Super Missile Pack in the Golden Torizo's Room
	s.InsertWithParent("goldTorizoSupers", false, "specificSupers")
	// Split on picking up the Super Missile Pack in the Wrecked Ship West Super Room
	s.InsertWithParent("wreckedShipLeftSupers", false, "specificSupers")
	// Split on picking up the Super Missile Pack in the Wrecked Ship East Super Room
	s.InsertWithParent("wreckedShipRightSupers", false, "specificSupers")
	// Split on picking up the Super Missile Pack in Main Street
	s.InsertWithParent("crabSupers", false, "specificSupers")
	// Split on picking up the Super Missile Pack in Watering Hole
	s.InsertWithParent("wateringHoleSupers", false, "specificSupers")
	// Split on picking up the Super Missile Pack in Aqueduct
	s.InsertWithParent("aqueductSupers", false, "specificSupers")
	// Split on the first Power Bomb pickup
	s.InsertWithParent("firstPowerBomb", true, "ammoPickups")
	// Split on each Power Bomb upgrade
	s.InsertWithParent("allPowerBombs", false, "ammoPickups")
	// Split on specific Power Bomb Pack locations
	s.InsertWithParent("specificBombs", false, "ammoPickups")
	// Split on picking up the Power Bomb Pack in the Crateria Power Bomb Room
	s.InsertWithParent("landingSiteBombs", false, "specificBombs")
	// Split on picking up the Power Bomb Pack in the Etecoon Room section of Green Brinstar Main Shaft
	s.InsertWithParent("etecoonBombs", false, "specificBombs")
	// Split on picking up the Power Bomb Pack in the Pink Brinstar Power Bomb Room
	s.InsertWithParent("pinkBrinstarBombs", false, "specificBombs")
	// Split on picking up the Power Bomb Pack in the Morph Ball Room
	s.InsertWithParent("blueBrinstarBombs", false, "specificBombs")
	// Split on picking up the Power Bomb Pack in the Alpha Power Bomb Room
	s.InsertWithParent("alphaBombs", false, "specificBombs")
	// Split on picking up the Power Bomb Pack in the Beta Power Bomb Room
	s.InsertWithParent("betaBombs", false, "specificBombs")
	// Split on picking up the Power Bomb Pack in the Post Crocomire Power Bomb Room
	s.InsertWithParent("crocomireBombs", false, "specificBombs")
	// Split on picking up the Power Bomb Pack in the Lower Norfair Escape Power Bomb Room
	s.InsertWithParent("lowerNorfairEscapeBombs", false, "specificBombs")
	// Split on picking up the Power Bomb Pack in Wasteland
	s.InsertWithParent("shameBombs", false, "specificBombs")
	// Split on picking up the Power Bomb Pack in East Sand Hall
	s.InsertWithParent("rightSandPitBombs", false, "specificBombs")

	// Split on Varia and Gravity pickups
	s.Insert("suitUpgrades", true)
	// Split on picking up the Varia Suit
	s.InsertWithParent("variaSuit", true, "suitUpgrades")
	// Split on picking up the Gravity Suit
	s.InsertWithParent("gravSuit", true, "suitUpgrades")

	// Split on beam upgrades
	s.Insert("beamUpgrades", true)
	// Split on picking up the Charge Beam
	s.InsertWithParent("chargeBeam", false, "beamUpgrades")
	// Split on picking up the Spazer
	s.InsertWithParent("spazer", false, "beamUpgrades")
	// Split on picking up the Wave Beam
	s.InsertWithParent("wave", true, "beamUpgrades")
	// Split on picking up the Ice Beam
	s.InsertWithParent("ice", false, "beamUpgrades")
	// Split on picking up the Plasma Beam
	s.InsertWithParent("plasma", false, "beamUpgrades")

	// Split on boot upgrades
	s.Insert("bootUpgrades", false)
	// Split on picking up the Hi-Jump Boots
	s.InsertWithParent("hiJump", false, "bootUpgrades")
	// Split on picking up Space Jump
	s.InsertWithParent("spaceJump", false, "bootUpgrades")
	// Split on picking up the Speed Booster
	s.InsertWithParent("speedBooster", false, "bootUpgrades")

	// Split on Energy Tanks and Reserve Tanks
	s.Insert("energyUpgrades", false)
	// Split on picking up the first Energy Tank
	s.InsertWithParent("firstETank", false, "energyUpgrades")
	// Split on picking up each Energy Tank
	s.InsertWithParent("allETanks", false, "energyUpgrades")
	// Split on specific Energy Tank locations
	s.InsertWithParent("specificETanks", false, "energyUpgrades")
	// Split on picking up the Energy Tank in the Gauntlet Energy Tank Room
	s.InsertWithParent("gauntletETank", false, "specificETanks")
	// Split on picking up the Energy Tank in the Terminator Room
	s.InsertWithParent("terminatorETank", false, "specificETanks")
	// Split on picking up the Energy Tank in the Blue Brinstar Energy Tank Room
	s.InsertWithParent("ceilingETank", false, "specificETanks")
	// Split on picking up the Energy Tank in the Etecoon Energy Tank Room
	s.InsertWithParent("etecoonsETank", false, "specificETanks")
	// Split on picking up the Energy Tank in Waterway
	s.InsertWithParent("waterwayETank", false, "specificETanks")
	// Split on picking up the Energy Tank in the Hopper Energy Tank Room
	s.InsertWithParent("waveGateETank", false, "specificETanks")
	// Split on picking up the Kraid Energy Tank in the Warehouse Energy Tank Room
	s.InsertWithParent("kraidETank", false, "specificETanks")
	// Split on picking up the Energy Tank in Crocomire's Room
	s.InsertWithParent("crocomireETank", false, "specificETanks")
	// Split on picking up the Energy Tank in the Hi Jump Energy Tank Room
	s.InsertWithParent("hiJumpETank", false, "specificETanks")
	// Split on picking up the Energy Tank in the Ridley Tank Room
	s.InsertWithParent("ridleyETank", false, "specificETanks")
	// Split on picking up the Energy Tank in the Lower Norfair Fireflea Room
	s.InsertWithParent("firefleaETank", false, "specificETanks")
	// Split on picking up the Energy Tank in the Wrecked Ship Energy Tank Room
	s.InsertWithParent("wreckedShipETank", false, "specificETanks")
	// Split on picking up the Energy Tank in the Mama Turtle Room
	s.InsertWithParent("tatoriETank", false, "specificETanks")
	// Split on picking up the Energy Tank in the Botwoon Energy Tank Room
	s.InsertWithParent("botwoonETank", false, "specificETanks")
	// Split on picking up each Reserve Tank
	s.InsertWithParent("reserveTanks", false, "energyUpgrades")
	// Split on specific Reserve Tank locations
	s.InsertWithParent("specificRTanks", false, "energyUpgrades")
	// Split on picking up the Reserve Tank in the Brinstar Reserve Tank Room
	s.InsertWithParent("brinstarReserve", false, "specificRTanks")
	// Split on picking up the Reserve Tank in the Norfair Reserve Tank Room
	s.InsertWithParent("norfairReserve", false, "specificRTanks")
	// Split on picking up the Reserve Tank in Bowling Alley
	s.InsertWithParent("wreckedShipReserve", false, "specificRTanks")
	// Split on picking up the Reserve Tank in West Sand Hole
	s.InsertWithParent("maridiaReserve", false, "specificRTanks")

	// Split on the miscellaneous upgrades
	s.Insert("miscUpgrades", false)
	// Split on picking up the Morphing Ball
	s.InsertWithParent("morphBall", false, "miscUpgrades")
	// Split on picking up the Bomb
	s.InsertWithParent("bomb", false, "miscUpgrades")
	// Split on picking up the Spring Ball
	s.InsertWithParent("springBall", false, "miscUpgrades")
	// Split on picking up the Screw Attack
	s.InsertWithParent("screwAttack", false, "miscUpgrades")
	// Split on picking up the Grapple Beam
	s.InsertWithParent("grapple", false, "miscUpgrades")
	// Split on picking up the X-Ray Scope
	s.InsertWithParent("xray", false, "miscUpgrades")

	// Split on transitions between areas
	s.Insert("areaTransitions", true)
	// Split on entering miniboss rooms (except Bomb Torizo)
	s.InsertWithParent("miniBossRooms", false, "areaTransitions")
	// Split on entering major boss rooms
	s.InsertWithParent("bossRooms", false, "areaTransitions")
	// Split on elevator transitions between areas (except Statue Room to Tourian)
	s.InsertWithParent("elevatorTransitions", false, "areaTransitions")
	// Split on leaving Ceres Station
	s.InsertWithParent("ceresEscape", false, "areaTransitions")
	// Split on entering the Wrecked Ship Entrance from the lower door of West Ocean
	s.InsertWithParent("wreckedShipEntrance", false, "areaTransitions")
	// Split on entering Red Tower from Noob Bridge
	s.InsertWithParent("redTowerMiddleEntrance", false, "areaTransitions")
	// Split on entering Red Tower from Skree Boost room
	s.InsertWithParent("redTowerBottomEntrance", false, "areaTransitions")
	// Split on entering Kraid's Lair
	s.InsertWithParent("kraidsLair", false, "areaTransitions")
	// Split on entering Rising Tide from Cathedral
	s.InsertWithParent("risingTideEntrance", false, "areaTransitions")
	// Split on exiting Attic
	s.InsertWithParent("atticExit", false, "areaTransitions")
	// Split on blowing up the tube to enter Maridia
	s.InsertWithParent("tubeBroken", false, "areaTransitions")
	// Split on exiting West Cacattack Alley
	s.InsertWithParent("cacExit", false, "areaTransitions")
	// Split on entering Toilet Bowl from either direction
	s.InsertWithParent("toilet", false, "areaTransitions")
	// Split on entering Kronic Boost room
	s.InsertWithParent("kronicBoost", false, "areaTransitions")
	// Split on the elevator down to Lower Norfair
	s.InsertWithParent("lowerNorfairEntrance", false, "areaTransitions")
	// Split on entering Worst Room in the Game
	s.InsertWithParent("writg", false, "areaTransitions")
	// Split on entering Red Kihunter Shaft from either Amphitheatre or Wastelands (NOTE: will split twice)
	s.InsertWithParent("redKiShaft", false, "areaTransitions")
	// Split on entering Metal Pirates Room from Wasteland
	s.InsertWithParent("metalPirates", false, "areaTransitions")
	// Split on entering Lower Norfair Springball Maze Room
	s.InsertWithParent("lowerNorfairSpringMaze", false, "areaTransitions")
	// Split on moving from the Three Musketeers' Room to the Single Chamber
	s.InsertWithParent("lowerNorfairExit", false, "areaTransitions")
	// Split on entering the Statues Room with all four major bosses defeated
	s.InsertWithParent("goldenFour", true, "areaTransitions")
	// Split on the elevator down to Tourian
	s.InsertWithParent("tourianEntrance", false, "areaTransitions")
	// Split on exiting each of the Metroid rooms in Tourian
	s.InsertWithParent("metroids", false, "areaTransitions")
	// Split on moving from the Dust Torizo Room to the Big Boy Room
	s.InsertWithParent("babyMetroidRoom", false, "areaTransitions")
	// Split on moving from Tourian Escape Room 4 to The Climb
	s.InsertWithParent("escapeClimb", false, "areaTransitions")

	// Split on defeating minibosses
	s.Insert("miniBosses", false)
	// Split on starting the Ceres Escape
	s.InsertWithParent("ceresRidley", false, "miniBosses")
	// Split on Bomb Torizo's drops appearing
	s.InsertWithParent("bombTorizo", false, "miniBosses")
	// Split on the last hit to Spore Spawn
	s.InsertWithParent("sporeSpawn", false, "miniBosses")
	// Split on Crocomire's drops appearing
	s.InsertWithParent("crocomire", false, "miniBosses")
	// Split on Botwoon's vertical column being fully destroyed
	s.InsertWithParent("botwoon", false, "miniBosses")
	// Split on Golden Torizo's drops appearing
	s.InsertWithParent("goldenTorizo", false, "miniBosses")

	// Split on defeating major bosses
	s.Insert("bosses", true)
	// Split shortly after Kraid's drops appear
	s.InsertWithParent("kraid", false, "bosses")
	// Split on Phantoon's drops appearing
	s.InsertWithParent("phantoon", false, "bosses")
	// Split on Draygon's drops appearing
	s.InsertWithParent("draygon", false, "bosses")
	// Split on Ridley's drops appearing
	s.InsertWithParent("ridley", true, "bosses")
	// Split on Mother Brain's head hitting the ground at the end of the first phase
	s.InsertWithParent("mb1", false, "bosses")
	// Split on the Baby Metroid detaching from Mother Brain's head
	s.InsertWithParent("mb2", true, "bosses")
	// Split on the start of the Zebes Escape
	s.InsertWithParent("mb3", false, "bosses")

	// Split on facing forward at the end of Zebes Escape
	s.Insert("rtaFinish", true)
	// Split on In-Game Time finalizing, when the end cutscene starts
	s.Insert("igtFinish", false)
	// Split on the end of a Spore Spawn RTA run, when the text box clears after collecting the Super Missiles
	s.Insert("sporeSpawnRTAFinish", false)
	// Split on the end of a 100 Missile RTA run, when the text box clears after collecting the hundredth missile
	s.Insert("hundredMissileRTAFinish", false)
	s.modifiedAfterCreation = false
	return s
}

func (s *Settings) Insert(name string, value bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.modifiedAfterCreation = true
	s.data[name] = struct {
		value  bool
		parent *string
	}{value: value, parent: nil}
}

func (s *Settings) InsertWithParent(name string, value bool, parent string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.modifiedAfterCreation = true
	p := parent
	s.data[name] = struct {
		value  bool
		parent *string
	}{value: value, parent: &p}
}

func (s *Settings) Contains(varName string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, ok := s.data[varName]
	return ok
}

func (s *Settings) Get(varName string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.getRecursive(varName)
}

func (s *Settings) getRecursive(varName string) bool {
	entry, ok := s.data[varName]
	if !ok {
		return false
	}
	if entry.parent == nil {
		return entry.value
	}
	return entry.value && s.getRecursive(*entry.parent)
}

func (s *Settings) Set(varName string, value bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	entry, ok := s.data[varName]
	if !ok {
		s.data[varName] = struct {
			value  bool
			parent *string
		}{value: value, parent: nil}
	} else {
		s.data[varName] = struct {
			value  bool
			parent *string
		}{value: value, parent: entry.parent}
	}
	s.modifiedAfterCreation = true
}

func (s *Settings) Roots() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var roots []string
	for k, v := range s.data {
		if v.parent == nil {
			roots = append(roots, k)
		}
	}
	return roots
}

func (s *Settings) Children(key string) []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var children []string
	for k, v := range s.data {
		if v.parent != nil && *v.parent == key {
			children = append(children, k)
		}
	}
	return children
}

func (s *Settings) Lookup(varName string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	entry, ok := s.data[varName]
	if !ok {
		panic("variable not found")
	}
	return entry.value
}

func (s *Settings) LookupMut(varName string) *bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	entry, ok := s.data[varName]
	if !ok {
		panic("variable not found")
	}
	s.modifiedAfterCreation = true
	// To mutate the value, we need to update the map entry.
	// Return a pointer to the value inside the map by re-assigning.
	// Since Go does not allow direct pointer to map values, we simulate with a helper struct.
	val := entry.value
	// parent := entry.parent
	// Create a wrapper struct to hold pointer to value
	type boolWrapper struct {
		val *bool
	}
	bw := boolWrapper{val: &val}
	// Return pointer to val, but user must call Set to update map.
	return bw.val
}

func (s *Settings) HasBeenModified() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.modifiedAfterCreation
}

func (s *Settings) SplitOnMiscUpgrades() {
	s.Set("miscUpgrades", true)
	s.Set("morphBall", true)
	s.Set("bomb", true)
	s.Set("springBall", true)
	s.Set("screwAttack", true)
	s.Set("grapple", true)
	s.Set("xray", true)
}

func (s *Settings) SplitOnHundo() {
	s.Set("ammoPickups", true)
	s.Set("allMissiles", true)
	s.Set("allSupers", true)
	s.Set("allPowerBombs", true)
	s.Set("beamUpgrades", true)
	s.Set("chargeBeam", true)
	s.Set("spazer", true)
	s.Set("wave", true)
	s.Set("ice", true)
	s.Set("plasma", true)
	s.Set("bootUpgrades", true)
	s.Set("hiJump", true)
	s.Set("spaceJump", true)
	s.Set("speedBooster", true)
	s.Set("energyUpgrades", true)
	s.Set("allETanks", true)
	s.Set("reserveTanks", true)
	s.SplitOnMiscUpgrades()
	s.Set("areaTransitions", true) // should already be true
	s.Set("tubeBroken", true)
	s.Set("ceresEscape", true)
	s.Set("bosses", true) // should already be true
	s.Set("kraid", true)
	s.Set("phantoon", true)
	s.Set("draygon", true)
	s.Set("ridley", true)
	s.Set("mb1", true)
	s.Set("mb2", true)
	s.Set("mb3", true)
	s.Set("miniBosses", true)
	s.Set("ceresRidley", true)
	s.Set("bombTorizo", true)
	s.Set("crocomire", true)
	s.Set("botwoon", true)
	s.Set("goldenTorizo", true)
	s.Set("babyMetroidRoom", true)
}

func (s *Settings) SplitOnAnyPercent() {
	s.Set("ammoPickups", true)
	s.Set("specificMissiles", true)
	s.Set("specificSupers", true)
	s.Set("wreckedShipLeftSupers", true)
	s.Set("specificPowerBombs", true)
	s.Set("firstMissile", true)
	s.Set("firstSuper", true)
	s.Set("firstPowerBomb", true)
	s.Set("brinstarMissiles", true)
	s.Set("norfairMissiles", true)
	s.Set("chargeMissiles", true)
	s.Set("waveMissiles", true)
	s.Set("beamUpgrades", true)
	s.Set("chargeBeam", true)
	s.Set("wave", true)
	s.Set("ice", true)
	s.Set("plasma", true)
	s.Set("bootUpgrades", true)
	s.Set("hiJump", true)
	s.Set("speedBooster", true)
	s.Set("specificETanks", true)
	s.Set("energyUpgrades", true)
	s.Set("terminatorETank", true)
	s.Set("hiJumpETank", true)
	s.Set("botwoonETank", true)
	s.Set("miscUpgrades", true)
	s.Set("morphBall", true)
	s.Set("spaceJump", true)
	s.Set("bomb", true)
	s.Set("areaTransitions", true) // should already be true
	s.Set("tubeBroken", true)
	s.Set("ceresEscape", true)
	s.Set("bosses", true) // should already be true
	s.Set("kraid", true)
	s.Set("phantoon", true)
	s.Set("draygon", true)
	s.Set("ridley", true)
	s.Set("mb1", true)
	s.Set("mb2", true)
	s.Set("mb3", true)
	s.Set("miniBosses", true)
	s.Set("ceresRidley", true)
	s.Set("bombTorizo", true)
	s.Set("botwoon", true)
	s.Set("goldenTorizo", true)
	s.Set("babyMetroidRoom", true)
}

// Width enum equivalent
type Width int

const (
	Byte Width = iota
	Word
)

type MemoryWatcher struct {
	address uint32
	current uint32
	old     uint32
	width   Width
}

func NewMemoryWatcher(address uint32, width Width) *MemoryWatcher {
	return &MemoryWatcher{
		address: address,
		current: 0,
		old:     0,
		width:   width,
	}
}

func (mw *MemoryWatcher) UpdateValue(memory []byte) {
	mw.old = mw.current
	switch mw.width {
	case Byte:
		mw.current = uint32(memory[mw.address])
	case Word:
		addr := mw.address
		mw.current = uint32(memory[addr]) | uint32(memory[addr+1])<<8
	}
}

func split(settings *Settings, snes *SNESState) bool {
	firstMissile := settings.Get("firstMissile") && snes.vars["maxMissiles"].old == 0 && snes.vars["maxMissiles"].current == 5
	allMissiles := settings.Get("allMissiles") && (snes.vars["maxMissiles"].old+5) == snes.vars["maxMissiles"].current
	oceanBottomMissiles := settings.Get("oceanBottomMissiles") && snes.vars["roomID"].current == roomIDEnum["westOcean"] && (snes.vars["crateriaItems"].old+2) == (snes.vars["crateriaItems"].current)
	oceanTopMissiles := settings.Get("oceanTopMissiles") && snes.vars["roomID"].current == roomIDEnum["westOcean"] && (snes.vars["crateriaItems"].old+4) == (snes.vars["crateriaItems"].current)
	oceanMiddleMissiles := settings.Get("oceanMiddleMissiles") && snes.vars["roomID"].current == roomIDEnum["westOcean"] && (snes.vars["crateriaItems"].old+8) == (snes.vars["crateriaItems"].current)
	moatMissiles := settings.Get("moatMissiles") && snes.vars["roomID"].current == roomIDEnum["crateriaMoat"] && (snes.vars["crateriaItems"].old+16) == (snes.vars["crateriaItems"].current)
	oldTourianMissiles := settings.Get("oldTourianMissiles") && snes.vars["roomID"].current == roomIDEnum["pitRoom"] && (snes.vars["crateriaItems"].old+64) == (snes.vars["crateriaItems"].current)
	gauntletRightMissiles := settings.Get("gauntletRightMissiles") && snes.vars["roomID"].current == roomIDEnum["greenPirateShaft"] && (snes.vars["brinteriaItems"].old+2) == (snes.vars["brinteriaItems"].current)
	gauntletLeftMissiles := settings.Get("gauntletLeftMissiles") && snes.vars["roomID"].current == roomIDEnum["greenPirateShaft"] && (snes.vars["brinteriaItems"].old+4) == (snes.vars["brinteriaItems"].current)
	dentalPlan := settings.Get("dentalPlan") && snes.vars["roomID"].current == roomIDEnum["theFinalMissile"] && (snes.vars["brinteriaItems"].old+16) == (snes.vars["brinteriaItems"].current)
	earlySuperBridgeMissiles := settings.Get("earlySuperBridgeMissiles") && snes.vars["roomID"].current == roomIDEnum["earlySupers"] && (snes.vars["brinteriaItems"].old+128) == (snes.vars["brinteriaItems"].current)
	greenBrinstarReserveMissiles := settings.Get("greenBrinstarReserveMissiles") && snes.vars["roomID"].current == roomIDEnum["brinstarReserveRoom"] && (snes.vars["brinstarItems2"].old+8) == (snes.vars["brinstarItems2"].current)
	greenBrinstarExtraReserveMissiles := settings.Get("greenBrinstarExtraReserveMissiles") && snes.vars["roomID"].current == roomIDEnum["brinstarReserveRoom"] && (snes.vars["brinstarItems2"].old+4) == (snes.vars["brinstarItems2"].current)
	bigPinkTopMissiles := settings.Get("bigPinkTopMissiles") && snes.vars["roomID"].current == roomIDEnum["bigPink"] && (snes.vars["brinstarItems2"].old+32) == (snes.vars["brinstarItems2"].current)
	chargeMissiles := settings.Get("chargeMissiles") && snes.vars["roomID"].current == roomIDEnum["bigPink"] && (snes.vars["brinstarItems2"].old+64) == (snes.vars["brinstarItems2"].current)
	greenHillsMissiles := settings.Get("greenHillsMissiles") && snes.vars["roomID"].current == roomIDEnum["greenHills"] && (snes.vars["brinstarItems3"].old+2) == (snes.vars["brinstarItems3"].current)
	blueBrinstarETankMissiles := settings.Get("blueBrinstarETankMissiles") && snes.vars["roomID"].current == roomIDEnum["blueBrinstarETankRoom"] && (snes.vars["brinstarItems3"].old+16) == (snes.vars["brinstarItems3"].current)
	alphaMissiles := settings.Get("alphaMissiles") && snes.vars["roomID"].current == roomIDEnum["alphaMissileRoom"] && (snes.vars["brinstarItems4"].old+4) == (snes.vars["brinstarItems4"].current)
	billyMaysMissiles := settings.Get("billyMaysMissiles") && snes.vars["roomID"].current == roomIDEnum["billyMays"] && (snes.vars["brinstarItems4"].old+16) == (snes.vars["brinstarItems4"].current)
	butWaitTheresMoreMissiles := settings.Get("butWaitTheresMoreMissiles") && snes.vars["roomID"].current == roomIDEnum["billyMays"] && (snes.vars["brinstarItems4"].old+32) == (snes.vars["brinstarItems4"].current)
	redBrinstarMissiles := settings.Get("redBrinstarMissiles") && snes.vars["roomID"].current == roomIDEnum["alphaPowerBombsRoom"] && (snes.vars["brinstarItems5"].old+2) == (snes.vars["brinstarItems5"].current)
	warehouseMissiles := settings.Get("warehouseMissiles") && snes.vars["roomID"].current == roomIDEnum["warehouseKiHunters"] && (snes.vars["brinstarItems5"].old+16) == (snes.vars["brinstarItems5"].current)
	cathedralMissiles := settings.Get("cathedralMissiles") && snes.vars["roomID"].current == roomIDEnum["cathedral"] && (snes.vars["norfairItems1"].old+2) == (snes.vars["norfairItems1"].current)
	crumbleShaftMissiles := settings.Get("crumbleShaftMissiles") && snes.vars["roomID"].current == roomIDEnum["crumbleShaft"] && (snes.vars["norfairItems1"].old+8) == (snes.vars["norfairItems1"].current)
	crocomireEscapeMissiles := settings.Get("crocomireEscapeMissiles") && snes.vars["roomID"].current == roomIDEnum["crocomireEscape"] && (snes.vars["norfairItems1"].old+64) == (snes.vars["norfairItems1"].current)
	hiJumpMissiles := settings.Get("hiJumpMissiles") && snes.vars["roomID"].current == roomIDEnum["hiJumpShaft"] && (snes.vars["norfairItems1"].old+128) == (snes.vars["norfairItems1"].current)
	postCrocomireMissiles := settings.Get("postCrocomireMissiles") && snes.vars["roomID"].current == roomIDEnum["cosineRoom"] && (snes.vars["norfairItems2"].old+4) == (snes.vars["norfairItems2"].current)
	grappleMissiles := settings.Get("grappleMissiles") && snes.vars["roomID"].current == roomIDEnum["preGrapple"] && (snes.vars["norfairItems2"].old+8) == (snes.vars["norfairItems2"].current)
	norfairReserveMissiles := settings.Get("norfairReserveMissiles") && snes.vars["roomID"].current == roomIDEnum["norfairReserveRoom"] && (snes.vars["norfairItems2"].old+64) == (snes.vars["norfairItems2"].current)
	greenBubblesMissiles := settings.Get("greenBubblesMissiles") && snes.vars["roomID"].current == roomIDEnum["greenBubblesRoom"] && (snes.vars["norfairItems2"].old+128) == (snes.vars["norfairItems2"].current)
	bubbleMountainMissiles := settings.Get("bubbleMountainMissiles") && snes.vars["roomID"].current == roomIDEnum["bubbleMountain"] && (snes.vars["norfairItems3"].old+1) == (snes.vars["norfairItems3"].current)
	speedBoostMissiles := settings.Get("speedBoostMissiles") && snes.vars["roomID"].current == roomIDEnum["speedBoostHall"] && (snes.vars["norfairItems3"].old+2) == (snes.vars["norfairItems3"].current)
	waveMissiles := settings.Get("waveMissiles") && snes.vars["roomID"].current == roomIDEnum["doubleChamber"] && (snes.vars["norfairItems3"].old+8) == (snes.vars["norfairItems3"].current)
	goldTorizoMissiles := settings.Get("goldTorizoMissiles") && snes.vars["roomID"].current == roomIDEnum["goldenTorizo"] && (snes.vars["norfairItems3"].old+64) == (snes.vars["norfairItems3"].current)
	mickeyMouseMissiles := settings.Get("mickeyMouseMissiles") && snes.vars["roomID"].current == roomIDEnum["mickeyMouse"] && (snes.vars["norfairItems4"].old+2) == (snes.vars["norfairItems4"].current)
	lowerNorfairSpringMazeMissiles := settings.Get("lowerNorfairSpringMazeMissiles") && snes.vars["roomID"].current == roomIDEnum["lowerNorfairSpringMaze"] && (snes.vars["norfairItems4"].old+4) == (snes.vars["norfairItems4"].current)
	threeMusketeersMissiles := settings.Get("threeMusketeersMissiles") && snes.vars["roomID"].current == roomIDEnum["threeMusketeers"] && (snes.vars["norfairItems4"].old+32) == (snes.vars["norfairItems4"].current)
	wreckedShipMainShaftMissiles := settings.Get("wreckedShipMainShaftMissiles") && snes.vars["roomID"].current == roomIDEnum["wreckedShipMainShaft"] && (snes.vars["wreckedShipItems"].old+1) == (snes.vars["wreckedShipItems"].current)
	bowlingMissiles := settings.Get("bowlingMissiles") && snes.vars["roomID"].current == roomIDEnum["bowling"] && (snes.vars["wreckedShipItems"].old+4) == (snes.vars["wreckedShipItems"].current)
	atticMissiles := settings.Get("atticMissiles") && snes.vars["roomID"].current == roomIDEnum["atticWorkerRobotRoom"] && (snes.vars["wreckedShipItems"].old+8) == (snes.vars["wreckedShipItems"].current)
	mainStreetMissiles := settings.Get("mainStreetMissiles") && snes.vars["roomID"].current == roomIDEnum["mainStreet"] && (snes.vars["maridiaItems1"].old+1) == (snes.vars["maridiaItems1"].current)
	mamaTurtleMissiles := settings.Get("mamaTurtleMissiles") && snes.vars["roomID"].current == roomIDEnum["mamaTurtle"] && (snes.vars["maridiaItems1"].old+8) == (snes.vars["maridiaItems1"].current)
	wateringHoleMissiles := settings.Get("wateringHoleMissiles") && snes.vars["roomID"].current == roomIDEnum["wateringHole"] && (snes.vars["maridiaItems1"].old+32) == (snes.vars["maridiaItems1"].current)
	beachMissiles := settings.Get("beachMissiles") && snes.vars["roomID"].current == roomIDEnum["beach"] && (snes.vars["maridiaItems1"].old+64) == (snes.vars["maridiaItems1"].current)
	leftSandPitMissiles := settings.Get("leftSandPitMissiles") && snes.vars["roomID"].current == roomIDEnum["leftSandPit"] && (snes.vars["maridiaItems2"].old+1) == (snes.vars["maridiaItems2"].current)
	rightSandPitMissiles := settings.Get("rightSandPitMissiles") && snes.vars["roomID"].current == roomIDEnum["rightSandPit"] && (snes.vars["maridiaItems2"].old+4) == (snes.vars["maridiaItems2"].current)
	aqueductMissiles := settings.Get("aqueductMissiles") && snes.vars["roomID"].current == roomIDEnum["aqueduct"] && (snes.vars["maridiaItems2"].old+16) == (snes.vars["maridiaItems2"].current)
	preDraygonMissiles := settings.Get("preDraygonMissiles") && snes.vars["roomID"].current == roomIDEnum["precious"] && (snes.vars["maridiaItems2"].old+128) == (snes.vars["maridiaItems2"].current)
	firstSuper := settings.Get("firstSuper") && snes.vars["maxSupers"].old == 0 && snes.vars["maxSupers"].current == 5
	allSupers := settings.Get("allSupers") && (snes.vars["maxSupers"].old+5) == (snes.vars["maxSupers"].current)
	climbSupers := settings.Get("climbSupers") && snes.vars["roomID"].current == roomIDEnum["crateriaSupersRoom"] && (snes.vars["brinteriaItems"].old+8) == (snes.vars["brinteriaItems"].current)
	sporeSpawnSupers := settings.Get("sporeSpawnSupers") && snes.vars["roomID"].current == roomIDEnum["sporeSpawnSuper"] && (snes.vars["brinteriaItems"].old+64) == (snes.vars["brinteriaItems"].current)
	earlySupers := settings.Get("earlySupers") && snes.vars["roomID"].current == roomIDEnum["earlySupers"] && (snes.vars["brinstarItems2"].old+1) == (snes.vars["brinstarItems2"].current)
	etecoonSupers := (settings.Get("etecoonSupers") || settings.Get("etacoonSupers")) && snes.vars["roomID"].current == roomIDEnum["etecoonSuperRoom"] && (snes.vars["brinstarItems3"].old+128) == (snes.vars["brinstarItems3"].current)
	goldTorizoSupers := settings.Get("goldTorizoSupers") && snes.vars["roomID"].current == roomIDEnum["goldenTorizo"] && (snes.vars["norfairItems3"].old+128) == (snes.vars["norfairItems3"].current)
	wreckedShipLeftSupers := settings.Get("wreckedShipLeftSupers") && snes.vars["roomID"].current == roomIDEnum["wreckedShipLeftSuperRoom"] && (snes.vars["wreckedShipItems"].old+32) == (snes.vars["wreckedShipItems"].current)
	wreckedShipRightSupers := settings.Get("wreckedShipRightSupers") && snes.vars["roomID"].current == roomIDEnum["wreckedShipRightSuperRoom"] && (snes.vars["wreckedShipItems"].old+64) == (snes.vars["wreckedShipItems"].current)
	crabSupers := settings.Get("crabSupers") && snes.vars["roomID"].current == roomIDEnum["mainStreet"] && (snes.vars["maridiaItems1"].old+2) == (snes.vars["maridiaItems1"].current)
	wateringHoleSupers := settings.Get("wateringHoleSupers") && snes.vars["roomID"].current == roomIDEnum["wateringHole"] && (snes.vars["maridiaItems1"].old+16) == (snes.vars["maridiaItems1"].current)
	aqueductSupers := settings.Get("aqueductSupers") && snes.vars["roomID"].current == roomIDEnum["aqueduct"] && (snes.vars["maridiaItems2"].old+32) == (snes.vars["maridiaItems2"].current)
	firstPowerBomb := settings.Get("firstPowerBomb") && snes.vars["maxPowerBombs"].old == 0 && snes.vars["maxPowerBombs"].current == 5
	allPowerBombs := settings.Get("allPowerBombs") && (snes.vars["maxPowerBombs"].old+5) == (snes.vars["maxPowerBombs"].current)
	landingSiteBombs := settings.Get("landingSiteBombs") && snes.vars["roomID"].current == roomIDEnum["crateriaPowerBombRoom"] && (snes.vars["crateriaItems"].old+1) == (snes.vars["crateriaItems"].current)
	etecoonBombs := (settings.Get("etecoonBombs") || settings.Get("etacoonBombs")) && snes.vars["roomID"].current == roomIDEnum["greenBrinstarMainShaft"] && (snes.vars["brinteriaItems"].old+32) == (snes.vars["brinteriaItems"].current)
	pinkBrinstarBombs := settings.Get("pinkBrinstarBombs") && snes.vars["roomID"].current == roomIDEnum["pinkBrinstarPowerBombRoom"] && (snes.vars["brinstarItems3"].old+1) == (snes.vars["brinstarItems3"].current)
	blueBrinstarBombs := settings.Get("blueBrinstarBombs") && snes.vars["roomID"].current == roomIDEnum["morphBall"] && (snes.vars["brinstarItems3"].old+8) == (snes.vars["brinstarItems3"].current)
	alphaBombs := settings.Get("alphaBombs") && snes.vars["roomID"].current == roomIDEnum["alphaPowerBombsRoom"] && (snes.vars["brinstarItems5"].old+1) == (snes.vars["brinstarItems5"].current)
	betaBombs := settings.Get("betaBombs") && snes.vars["roomID"].current == roomIDEnum["betaPowerBombRoom"] && (snes.vars["brinstarItems4"].old+128) == (snes.vars["brinstarItems4"].current)
	crocomireBombs := settings.Get("crocomireBombs") && snes.vars["roomID"].current == roomIDEnum["postCrocomirePowerBombRoom"] && (snes.vars["norfairItems2"].old+2) == (snes.vars["norfairItems2"].current)
	lowerNorfairEscapeBombs := settings.Get("lowerNorfairEscapeBombs") && snes.vars["roomID"].current == roomIDEnum["lowerNorfairEscapePowerBombRoom"] && (snes.vars["norfairItems4"].old+8) == (snes.vars["norfairItems4"].current)
	shameBombs := settings.Get("shameBombs") && snes.vars["roomID"].current == roomIDEnum["wasteland"] && (snes.vars["norfairItems4"].old+16) == (snes.vars["norfairItems4"].current)
	rightSandPitBombs := settings.Get("rightSandPitBombs") && snes.vars["roomID"].current == roomIDEnum["rightSandPit"] && (snes.vars["maridiaItems2"].old+8) == (snes.vars["maridiaItems2"].current)
	pickup := firstMissile || allMissiles || oceanBottomMissiles || oceanTopMissiles || oceanMiddleMissiles || moatMissiles || oldTourianMissiles || gauntletRightMissiles || gauntletLeftMissiles || dentalPlan || earlySuperBridgeMissiles || greenBrinstarReserveMissiles || greenBrinstarExtraReserveMissiles || bigPinkTopMissiles || chargeMissiles || greenHillsMissiles || blueBrinstarETankMissiles || alphaMissiles || billyMaysMissiles || butWaitTheresMoreMissiles || redBrinstarMissiles || warehouseMissiles || cathedralMissiles || crumbleShaftMissiles || crocomireEscapeMissiles || hiJumpMissiles || postCrocomireMissiles || grappleMissiles || norfairReserveMissiles || greenBubblesMissiles || bubbleMountainMissiles || speedBoostMissiles || waveMissiles || goldTorizoMissiles || mickeyMouseMissiles || lowerNorfairSpringMazeMissiles || threeMusketeersMissiles || wreckedShipMainShaftMissiles || bowlingMissiles || atticMissiles || mainStreetMissiles || mamaTurtleMissiles || wateringHoleMissiles || beachMissiles || leftSandPitMissiles || rightSandPitMissiles || aqueductMissiles || preDraygonMissiles || firstSuper || allSupers || climbSupers || sporeSpawnSupers || earlySupers || etecoonSupers || goldTorizoSupers || wreckedShipLeftSupers || wreckedShipRightSupers || crabSupers || wateringHoleSupers || aqueductSupers || firstPowerBomb || allPowerBombs || landingSiteBombs || etecoonBombs || pinkBrinstarBombs || blueBrinstarBombs || alphaBombs || betaBombs || crocomireBombs || lowerNorfairEscapeBombs || shameBombs || rightSandPitBombs

	// Item unlock section
	varia := settings.Get("variaSuit") && snes.vars["roomID"].current == roomIDEnum["varia"] && (snes.vars["unlockedEquips2"].old&unlockFlagEnum["variaSuit"]) == 0 && (snes.vars["unlockedEquips2"].current&unlockFlagEnum["variaSuit"]) > 0
	springBall := settings.Get("springBall") && snes.vars["roomID"].current == roomIDEnum["springBall"] && (snes.vars["unlockedEquips2"].old&unlockFlagEnum["springBall"]) == 0 && (snes.vars["unlockedEquips2"].current&unlockFlagEnum["springBall"]) > 0
	morphBall := settings.Get("morphBall") && snes.vars["roomID"].current == roomIDEnum["morphBall"] && (snes.vars["unlockedEquips2"].old&unlockFlagEnum["morphBall"]) == 0 && (snes.vars["unlockedEquips2"].current&unlockFlagEnum["morphBall"]) > 0
	screwAttack := settings.Get("screwAttack") && snes.vars["roomID"].current == roomIDEnum["screwAttack"] && (snes.vars["unlockedEquips2"].old&unlockFlagEnum["screwAttack"]) == 0 && (snes.vars["unlockedEquips2"].current&unlockFlagEnum["screwAttack"]) > 0
	gravSuit := settings.Get("gravSuit") && snes.vars["roomID"].current == roomIDEnum["gravity"] && (snes.vars["unlockedEquips2"].old&unlockFlagEnum["gravSuit"]) == 0 && (snes.vars["unlockedEquips2"].current&unlockFlagEnum["gravSuit"]) > 0
	hiJump := settings.Get("hiJump") && snes.vars["roomID"].current == roomIDEnum["hiJump"] && (snes.vars["unlockedEquips"].old&unlockFlagEnum["hiJump"]) == 0 && (snes.vars["unlockedEquips"].current&unlockFlagEnum["hiJump"]) > 0
	spaceJump := settings.Get("spaceJump") && snes.vars["roomID"].current == roomIDEnum["spaceJump"] && (snes.vars["unlockedEquips"].old&unlockFlagEnum["spaceJump"]) == 0 && (snes.vars["unlockedEquips"].current&unlockFlagEnum["spaceJump"]) > 0
	bomb := settings.Get("bomb") && snes.vars["roomID"].current == roomIDEnum["bombTorizo"] && (snes.vars["unlockedEquips"].old&unlockFlagEnum["bomb"]) == 0 && (snes.vars["unlockedEquips"].current&unlockFlagEnum["bomb"]) > 0
	speedBooster := settings.Get("speedBooster") && snes.vars["roomID"].current == roomIDEnum["speedBooster"] && (snes.vars["unlockedEquips"].old&unlockFlagEnum["speedBooster"]) == 0 && (snes.vars["unlockedEquips"].current&unlockFlagEnum["speedBooster"]) > 0
	grapple := settings.Get("grapple") && snes.vars["roomID"].current == roomIDEnum["grapple"] && (snes.vars["unlockedEquips"].old&unlockFlagEnum["grapple"]) == 0 && (snes.vars["unlockedEquips"].current&unlockFlagEnum["grapple"]) > 0
	xray := settings.Get("xray") && snes.vars["roomID"].current == roomIDEnum["xRay"] && (snes.vars["unlockedEquips"].old&unlockFlagEnum["xray"]) == 0 && (snes.vars["unlockedEquips"].current&unlockFlagEnum["xray"]) > 0
	unlock := varia || springBall || morphBall || screwAttack || gravSuit || hiJump || spaceJump || bomb || speedBooster || grapple || xray

	// Beam unlock section
	wave := settings.Get("wave") && snes.vars["roomID"].current == roomIDEnum["waveBeam"] && (snes.vars["unlockedBeams"].old&unlockFlagEnum["wave"]) == 0 && (snes.vars["unlockedBeams"].current&unlockFlagEnum["wave"]) > 0
	ice := settings.Get("ice") && snes.vars["roomID"].current == roomIDEnum["iceBeam"] && (snes.vars["unlockedBeams"].old&unlockFlagEnum["ice"]) == 0 && (snes.vars["unlockedBeams"].current&unlockFlagEnum["ice"]) > 0
	spazer := settings.Get("spazer") && snes.vars["roomID"].current == roomIDEnum["spazer"] && (snes.vars["unlockedBeams"].old&unlockFlagEnum["spazer"]) == 0 && (snes.vars["unlockedBeams"].current&unlockFlagEnum["spazer"]) > 0
	plasma := settings.Get("plasma") && snes.vars["roomID"].current == roomIDEnum["plasmaBeam"] && (snes.vars["unlockedBeams"].old&unlockFlagEnum["plasma"]) == 0 && (snes.vars["unlockedBeams"].current&unlockFlagEnum["plasma"]) > 0
	chargeBeam := settings.Get("chargeBeam") && snes.vars["roomID"].current == roomIDEnum["bigPink"] && (snes.vars["unlockedCharge"].old&unlockFlagEnum["chargeBeam"]) == 0 && (snes.vars["unlockedCharge"].current&unlockFlagEnum["chargeBeam"]) > 0
	beam := wave || ice || spazer || plasma || chargeBeam

	// E-tanks and reserve tanks
	firstETank := settings.Get("firstETank") && snes.vars["maxEnergy"].old == 99 && snes.vars["maxEnergy"].current == 199
	allETanks := settings.Get("allETanks") && (snes.vars["maxEnergy"].old+100) == (snes.vars["maxEnergy"].current)
	gauntletETank := settings.Get("gauntletETank") && snes.vars["roomID"].current == roomIDEnum["gauntletETankRoom"] && (snes.vars["crateriaItems"].old+32) == (snes.vars["crateriaItems"].current)
	terminatorETank := settings.Get("terminatorETank") && snes.vars["roomID"].current == roomIDEnum["terminator"] && (snes.vars["brinteriaItems"].old+1) == (snes.vars["brinteriaItems"].current)
	ceilingETank := settings.Get("ceilingETank") && snes.vars["roomID"].current == roomIDEnum["blueBrinstarETankRoom"] && (snes.vars["brinstarItems3"].old+32) == (snes.vars["brinstarItems3"].current)
	etecoonsETank := (settings.Get("etecoonsETank") || settings.Get("etacoonsETank")) && snes.vars["roomID"].current == roomIDEnum["etecoonETankRoom"] && (snes.vars["brinstarItems3"].old+64) == (snes.vars["brinstarItems3"].current)
	waterwayETank := settings.Get("waterwayETank") && snes.vars["roomID"].current == roomIDEnum["waterway"] && (snes.vars["brinstarItems4"].old+2) == (snes.vars["brinstarItems4"].current)
	waveGateETank := settings.Get("waveGateETank") && snes.vars["roomID"].current == roomIDEnum["hopperETankRoom"] && (snes.vars["brinstarItems4"].old+8) == (snes.vars["brinstarItems4"].current)
	kraidETank := settings.Get("kraidETank") && snes.vars["roomID"].current == roomIDEnum["warehouseETankRoom"] && (snes.vars["brinstarItems5"].old+8) == (snes.vars["brinstarItems5"].current)
	crocomireETank := settings.Get("crocomireETank") && snes.vars["roomID"].current == roomIDEnum["crocomire"] && (snes.vars["norfairItems1"].old+16) == (snes.vars["norfairItems1"].current)
	hiJumpETank := settings.Get("hiJumpETank") && snes.vars["roomID"].current == roomIDEnum["hiJumpShaft"] && (snes.vars["norfairItems2"].old+1) == (snes.vars["norfairItems2"].current)
	ridleyETank := settings.Get("ridleyETank") && snes.vars["roomID"].current == roomIDEnum["ridleyETankRoom"] && (snes.vars["norfairItems4"].old+64) == (snes.vars["norfairItems4"].current)
	firefleaETank := settings.Get("firefleaETank") && snes.vars["roomID"].current == roomIDEnum["lowerNorfairFireflea"] && (snes.vars["norfairItems5"].old+1) == (snes.vars["norfairItems5"].current)
	wreckedShipETank := settings.Get("wreckedShipETank") && snes.vars["roomID"].current == roomIDEnum["wreckedShipETankRoom"] && (snes.vars["wreckedShipItems"].old+16) == (snes.vars["wreckedShipItems"].current)
	tatoriETank := settings.Get("tatoriETank") && snes.vars["roomID"].current == roomIDEnum["mamaTurtle"] && (snes.vars["maridiaItems1"].old+4) == (snes.vars["maridiaItems1"].current)
	botwoonETank := settings.Get("botwoonETank") && snes.vars["roomID"].current == roomIDEnum["botwoonETankRoom"] && (snes.vars["maridiaItems3"].old+1) == (snes.vars["maridiaItems3"].current)
	reserveTanks := settings.Get("reserveTanks") && (snes.vars["maxReserve"].old+100) == (snes.vars["maxReserve"].current)
	brinstarReserve := settings.Get("brinstarReserve") && snes.vars["roomID"].current == roomIDEnum["brinstarReserveRoom"] && (snes.vars["brinstarItems2"].old+2) == (snes.vars["brinstarItems2"].current)
	norfairReserve := settings.Get("norfairReserve") && snes.vars["roomID"].current == roomIDEnum["norfairReserveRoom"] && (snes.vars["norfairItems2"].old+32) == (snes.vars["norfairItems2"].current)
	wreckedShipReserve := settings.Get("wreckedShipReserve") && snes.vars["roomID"].current == roomIDEnum["bowling"] && (snes.vars["wreckedShipItems"].old+2) == (snes.vars["wreckedShipItems"].current)
	maridiaReserve := settings.Get("maridiaReserve") && snes.vars["roomID"].current == roomIDEnum["leftSandPit"] && (snes.vars["maridiaItems2"].old+2) == (snes.vars["maridiaItems2"].current)
	energyUpgrade := firstETank || allETanks || gauntletETank || terminatorETank || ceilingETank || etecoonsETank || waterwayETank || waveGateETank || kraidETank || crocomireETank || hiJumpETank || ridleyETank || firefleaETank || wreckedShipETank || tatoriETank || botwoonETank || reserveTanks || brinstarReserve || norfairReserve || wreckedShipReserve || maridiaReserve

	// Miniboss room transitions
	miniBossRooms := false
	if settings.Get("miniBossRooms") {
		ceresRidleyRoom := snes.vars["roomID"].old == roomIDEnum["flatRoom"] && snes.vars["roomID"].current == roomIDEnum["ceresRidley"]
		sporeSpawnRoom := snes.vars["roomID"].old == roomIDEnum["sporeSpawnKeyhunter"] && snes.vars["roomID"].current == roomIDEnum["sporeSpawn"]
		crocomireRoom := snes.vars["roomID"].old == roomIDEnum["crocomireSpeedway"] && snes.vars["roomID"].current == roomIDEnum["crocomire"]
		botwoonRoom := snes.vars["roomID"].old == roomIDEnum["botwoonHallway"] && snes.vars["roomID"].current == roomIDEnum["botwoon"]
		// Allow either vanilla or GGG entry
		goldenTorizoRoom := (snes.vars["roomID"].old == roomIDEnum["acidStatue"] || snes.vars["roomID"].old == roomIDEnum["screwAttack"]) && snes.vars["roomID"].current == roomIDEnum["goldenTorizo"]
		miniBossRooms = ceresRidleyRoom || sporeSpawnRoom || crocomireRoom || botwoonRoom || goldenTorizoRoom
	}

	// Boss room transitions
	bossRooms := false
	if settings.Get("bossRooms") {
		kraidRoom := snes.vars["roomID"].old == roomIDEnum["kraidEyeDoor"] && snes.vars["roomID"].current == roomIDEnum["kraid"]
		phantoonRoom := snes.vars["roomID"].old == roomIDEnum["basement"] && snes.vars["roomID"].current == roomIDEnum["phantoon"]
		draygonRoom := snes.vars["roomID"].old == roomIDEnum["precious"] && snes.vars["roomID"].current == roomIDEnum["draygon"]
		ridleyRoom := snes.vars["roomID"].old == roomIDEnum["lowerNorfairFarming"] && snes.vars["roomID"].current == roomIDEnum["ridley"]
		motherBrainRoom := snes.vars["roomID"].old == roomIDEnum["rinkaShaft"] && snes.vars["roomID"].current == roomIDEnum["motherBrain"]
		bossRooms = kraidRoom || phantoonRoom || draygonRoom || ridleyRoom || motherBrainRoom
	}

	// Elevator transitions between areas
	elevatorTransitions := false
	if settings.Get("elevatorTransitions") {
		blueBrinstar := (snes.vars["roomID"].old == roomIDEnum["elevatorToMorphBall"] && snes.vars["roomID"].current == roomIDEnum["morphBall"]) || (snes.vars["roomID"].old == roomIDEnum["morphBall"] && snes.vars["roomID"].current == roomIDEnum["elevatorToMorphBall"])
		greenBrinstar := (snes.vars["roomID"].old == roomIDEnum["elevatorToGreenBrinstar"] && snes.vars["roomID"].current == roomIDEnum["greenBrinstarMainShaft"]) || (snes.vars["roomID"].old == roomIDEnum["greenBrinstarMainShaft"] && snes.vars["roomID"].current == roomIDEnum["elevatorToGreenBrinstar"])
		businessCenter := (snes.vars["roomID"].old == roomIDEnum["warehouseEntrance"] && snes.vars["roomID"].current == roomIDEnum["businessCenter"]) || (snes.vars["roomID"].old == roomIDEnum["businessCenter"] && snes.vars["roomID"].current == roomIDEnum["warehouseEntrance"])
		caterpillar := (snes.vars["roomID"].old == roomIDEnum["elevatorToCaterpillar"] && snes.vars["roomID"].current == roomIDEnum["caterpillar"]) || (snes.vars["roomID"].old == roomIDEnum["caterpillar"] && snes.vars["roomID"].current == roomIDEnum["elevatorToCaterpillar"])
		maridiaElevator := (snes.vars["roomID"].old == roomIDEnum["elevatorToMaridia"] && snes.vars["roomID"].current == roomIDEnum["maridiaElevator"]) || (snes.vars["roomID"].old == roomIDEnum["maridiaElevator"] && snes.vars["roomID"].current == roomIDEnum["elevatorToMaridia"])
		elevatorTransitions = blueBrinstar || greenBrinstar || businessCenter || caterpillar || maridiaElevator
	}

	// Room transitions
	ceresEscape := settings.Get("ceresEscape") && snes.vars["roomID"].current == roomIDEnum["ceresElevator"] && snes.vars["gameState"].old == gameStateEnum["normalGameplay"] && snes.vars["gameState"].current == gameStateEnum["startOfCeresCutscene"]
	wreckedShipEntrance := settings.Get("wreckedShipEntrance") && snes.vars["roomID"].old == roomIDEnum["westOcean"] && snes.vars["roomID"].current == roomIDEnum["wreckedShipEntrance"]
	redTowerMiddleEntrance := settings.Get("redTowerMiddleEntrance") && snes.vars["roomID"].old == roomIDEnum["noobBridge"] && snes.vars["roomID"].current == roomIDEnum["redTower"]
	redTowerBottomEntrance := settings.Get("redTowerBottomEntrance") && snes.vars["roomID"].old == roomIDEnum["bat"] && snes.vars["roomID"].current == roomIDEnum["redTower"]
	kraidsLair := settings.Get("kraidsLair") && snes.vars["roomID"].old == roomIDEnum["warehouseEntrance"] && snes.vars["roomID"].current == roomIDEnum["warehouseZeela"]
	risingTideEntrance := settings.Get("risingTideEntrance") && snes.vars["roomID"].old == roomIDEnum["cathedral"] && snes.vars["roomID"].current == roomIDEnum["risingTide"]
	atticExit := settings.Get("atticExit") && snes.vars["roomID"].old == roomIDEnum["attic"] && snes.vars["roomID"].current == roomIDEnum["westOcean"]
	tubeBroken := settings.Get("tubeBroken") && snes.vars["roomID"].current == roomIDEnum["glassTunnel"] && (snes.vars["eventFlags"].old&eventFlagEnum["tubeBroken"]) == 0 && (snes.vars["eventFlags"].current&eventFlagEnum["tubeBroken"]) > 0
	cacExit := settings.Get("cacExit") && snes.vars["roomID"].old == roomIDEnum["westCactusAlley"] && snes.vars["roomID"].current == roomIDEnum["butterflyRoom"]
	toilet := settings.Get("toilet") && (snes.vars["roomID"].old == roomIDEnum["plasmaSpark"] && snes.vars["roomID"].current == roomIDEnum["toiletBowl"] || snes.vars["roomID"].old == roomIDEnum["oasis"] && snes.vars["roomID"].current == roomIDEnum["toiletBowl"])
	kronicBoost := settings.Get("kronicBoost") && (snes.vars["roomID"].old == roomIDEnum["magdolliteTunnel"] && snes.vars["roomID"].current == roomIDEnum["kronicBoost"] || snes.vars["roomID"].old == roomIDEnum["spikyAcidSnakes"] && snes.vars["roomID"].current == roomIDEnum["kronicBoost"] || snes.vars["roomID"].old == roomIDEnum["volcano"] && snes.vars["roomID"].current == roomIDEnum["kronicBoost"])
	lowerNorfairEntrance := settings.Get("lowerNorfairEntrance") && snes.vars["roomID"].old == roomIDEnum["lowerNorfairElevator"] && snes.vars["roomID"].current == roomIDEnum["mainHall"]
	writg := settings.Get("writg") && snes.vars["roomID"].old == roomIDEnum["pillars"] && snes.vars["roomID"].current == roomIDEnum["writg"]
	redKiShaft := settings.Get("redKiShaft") && (snes.vars["roomID"].old == roomIDEnum["amphitheatre"] && snes.vars["roomID"].current == roomIDEnum["redKiShaft"] || snes.vars["roomID"].old == roomIDEnum["wasteland"] && snes.vars["roomID"].current == roomIDEnum["redKiShaft"])
	metalPirates := settings.Get("metalPirates") && snes.vars["roomID"].old == roomIDEnum["wasteland"] && snes.vars["roomID"].current == roomIDEnum["metalPirates"]
	lowerNorfairSpringMaze := settings.Get("lowerNorfairSpringMaze") && snes.vars["roomID"].old == roomIDEnum["lowerNorfairFireflea"] && snes.vars["roomID"].current == roomIDEnum["lowerNorfairSpringMaze"]
	lowerNorfairExit := settings.Get("lowerNorfairExit") && snes.vars["roomID"].old == roomIDEnum["threeMusketeers"] && snes.vars["roomID"].current == roomIDEnum["singleChamber"]
	allBossesFinished := (snes.vars["brinstarBosses"].current&bossFlagEnum["kraid"]) > 0 && (snes.vars["wreckedShipBosses"].current&bossFlagEnum["phantoon"]) > 0 && (snes.vars["maridiaBosses"].current&bossFlagEnum["draygon"]) > 0 && (snes.vars["norfairBosses"].current&bossFlagEnum["ridley"]) > 0
	goldenFour := settings.Get("goldenFour") && snes.vars["roomID"].old == roomIDEnum["statuesHallway"] && snes.vars["roomID"].current == roomIDEnum["statues"] && allBossesFinished
	tourianEntrance := settings.Get("tourianEntrance") && snes.vars["roomID"].old == roomIDEnum["statues"] && snes.vars["roomID"].current == roomIDEnum["tourianElevator"]
	metroids := settings.Get("metroids") && (snes.vars["roomID"].old == roomIDEnum["metroidOne"] && snes.vars["roomID"].current == roomIDEnum["metroidTwo"] || snes.vars["roomID"].old == roomIDEnum["metroidTwo"] && snes.vars["roomID"].current == roomIDEnum["metroidThree"] || snes.vars["roomID"].old == roomIDEnum["metroidThree"] && snes.vars["roomID"].current == roomIDEnum["metroidFour"] || snes.vars["roomID"].old == roomIDEnum["metroidFour"] && snes.vars["roomID"].current == roomIDEnum["tourianHopper"])
	babyMetroidRoom := settings.Get("babyMetroidRoom") && snes.vars["roomID"].old == roomIDEnum["dustTorizo"] && snes.vars["roomID"].current == roomIDEnum["bigBoy"]
	escapeClimb := settings.Get("escapeClimb") && snes.vars["roomID"].old == roomIDEnum["tourianEscape4"] && snes.vars["roomID"].current == roomIDEnum["climb"]
	roomTransitions := miniBossRooms || bossRooms || elevatorTransitions || ceresEscape || wreckedShipEntrance || redTowerMiddleEntrance || redTowerBottomEntrance || kraidsLair || risingTideEntrance || atticExit || tubeBroken || cacExit || toilet || kronicBoost || lowerNorfairEntrance || writg || redKiShaft || metalPirates || lowerNorfairSpringMaze || lowerNorfairExit || tourianEntrance || goldenFour || metroids || babyMetroidRoom || escapeClimb

	// Minibosses
	ceresRidley := settings.Get("ceresRidley") && (snes.vars["ceresBosses"].old&bossFlagEnum["ceresRidley"]) == 0 && (snes.vars["ceresBosses"].current&bossFlagEnum["ceresRidley"]) > 0 && snes.vars["roomID"].current == roomIDEnum["ceresRidley"]
	bombTorizo := settings.Get("bombTorizo") && (snes.vars["crateriaBosses"].old&bossFlagEnum["bombTorizo"]) == 0 && (snes.vars["crateriaBosses"].current&bossFlagEnum["bombTorizo"]) > 0 && snes.vars["roomID"].current == roomIDEnum["bombTorizo"]
	sporeSpawn := settings.Get("sporeSpawn") && (snes.vars["brinstarBosses"].old&bossFlagEnum["sporeSpawn"]) == 0 && (snes.vars["brinstarBosses"].current&bossFlagEnum["sporeSpawn"]) > 0 && snes.vars["roomID"].current == roomIDEnum["sporeSpawn"]
	crocomire := settings.Get("crocomire") && (snes.vars["norfairBosses"].old&bossFlagEnum["crocomire"]) == 0 && (snes.vars["norfairBosses"].current&bossFlagEnum["crocomire"]) > 0 && snes.vars["roomID"].current == roomIDEnum["crocomire"]
	botwoon := settings.Get("botwoon") && (snes.vars["maridiaBosses"].old&bossFlagEnum["botwoon"]) == 0 && (snes.vars["maridiaBosses"].current&bossFlagEnum["botwoon"]) > 0 && snes.vars["roomID"].current == roomIDEnum["botwoon"]
	goldenTorizo := settings.Get("goldenTorizo") && (snes.vars["norfairBosses"].old&bossFlagEnum["goldenTorizo"]) == 0 && (snes.vars["norfairBosses"].current&bossFlagEnum["goldenTorizo"]) > 0 && snes.vars["roomID"].current == roomIDEnum["goldenTorizo"]
	minibossDefeat := ceresRidley || bombTorizo || sporeSpawn || crocomire || botwoon || goldenTorizo

	// Bosses
	kraid := settings.Get("kraid") && (snes.vars["brinstarBosses"].old&bossFlagEnum["kraid"]) == 0 && (snes.vars["brinstarBosses"].current&bossFlagEnum["kraid"]) > 0 && snes.vars["roomID"].current == roomIDEnum["kraid"]
	if kraid {
		fmt.Println("Split due to kraid defeat")
	}
	phantoon := settings.Get("phantoon") && (snes.vars["wreckedShipBosses"].old&bossFlagEnum["phantoon"]) == 0 && (snes.vars["wreckedShipBosses"].current&bossFlagEnum["phantoon"]) > 0 && snes.vars["roomID"].current == roomIDEnum["phantoon"]
	if phantoon {
		fmt.Println("Split due to phantoon defeat")
	}
	draygon := settings.Get("draygon") && (snes.vars["maridiaBosses"].old&bossFlagEnum["draygon"]) == 0 && (snes.vars["maridiaBosses"].current&bossFlagEnum["draygon"]) > 0 && snes.vars["roomID"].current == roomIDEnum["draygon"]
	if draygon {
		fmt.Println("Split due to draygon defeat")
	}
	ridley := settings.Get("ridley") && (snes.vars["norfairBosses"].old&bossFlagEnum["ridley"]) == 0 && (snes.vars["norfairBosses"].current&bossFlagEnum["ridley"]) > 0 && snes.vars["roomID"].current == roomIDEnum["ridley"]
	if ridley {
		fmt.Println("Split due to ridley defeat")
	}
	// Mother Brain phases
	inMotherBrainRoom := snes.vars["roomID"].current == roomIDEnum["motherBrain"]
	mb1 := settings.Get("mb1") && inMotherBrainRoom && snes.vars["gameState"].current == gameStateEnum["normalGameplay"] && snes.vars["motherBrainHP"].old == 0 && snes.vars["motherBrainHP"].current == (motherBrainMaxHPEnum["phase2"])
	if mb1 {
		fmt.Println("Split due to mb1 defeat")
	}
	mb2 := settings.Get("mb2") && inMotherBrainRoom && snes.vars["gameState"].current == gameStateEnum["normalGameplay"] && snes.vars["motherBrainHP"].old == 0 && snes.vars["motherBrainHP"].current == (motherBrainMaxHPEnum["phase3"])
	if mb2 {
		fmt.Println("Split due to mb2 defeat")
	}
	mb3 := settings.Get("mb3") && inMotherBrainRoom && (snes.vars["tourianBosses"].old&bossFlagEnum["motherBrain"]) == 0 && (snes.vars["tourianBosses"].current&bossFlagEnum["motherBrain"]) > 0
	if mb3 {
		fmt.Println("Split due to mb3 defeat")
	}
	bossDefeat := kraid || phantoon || draygon || ridley || mb1 || mb2 || mb3

	// Run-ending splits
	escape := settings.Get("rtaFinish") && (snes.vars["eventFlags"].current&eventFlagEnum["zebesAblaze"]) > 0 && snes.vars["shipAI"].old != 0xaa4f && snes.vars["shipAI"].current == 0xaa4f

	takeoff := settings.Get("igtFinish") && snes.vars["roomID"].current == roomIDEnum["landingSite"] && snes.vars["gameState"].old == gameStateEnum["preEndCutscene"] && snes.vars["gameState"].current == gameStateEnum["endCutscene"]

	sporeSpawnRTAFinish := false
	if settings.Get("sporeSpawnRTAFinish") {
		if snes.pickedUpSporeSpawnSuper {
			if snes.vars["igtFrames"].old != snes.vars["igtFrames"].current {
				sporeSpawnRTAFinish = true
				snes.pickedUpSporeSpawnSuper = false
			}
		} else {
			snes.pickedUpSporeSpawnSuper = snes.vars["roomID"].current == roomIDEnum["sporeSpawnSuper"] && (snes.vars["maxSupers"].old+5) == (snes.vars["maxSupers"].current) && (snes.vars["brinstarBosses"].current&bossFlagEnum["sporeSpawn"]) > 0
		}
	}

	hundredMissileRTAFinish := false
	if settings.Get("hundredMissileRTAFinish") {
		if snes.pickedUpHundredthMissile {
			if snes.vars["igtFrames"].old != snes.vars["igtFrames"].current {
				hundredMissileRTAFinish = true
				snes.pickedUpHundredthMissile = false
			}
		} else {
			snes.pickedUpHundredthMissile = snes.vars["maxMissiles"].old == 95 && snes.vars["maxMissiles"].current == 100
		}
	}

	nonStandardCategoryFinish := sporeSpawnRTAFinish || hundredMissileRTAFinish

	if pickup {
		fmt.Println("Split due to pickup")
	}

	if unlock {
		fmt.Println("Split due to unlock")
	}

	if beam {
		fmt.Println("Split due to beam upgrade")
	}

	if energyUpgrade {
		fmt.Println("Split due to energy upgrade")
	}

	if roomTransitions {
		fmt.Println("Split due to room transition")
	}

	if minibossDefeat {
		fmt.Println("Split due to miniboss defeat")
	}

	// individual boss defeat conditions already covered above
	if escape {
		fmt.Println("Split due to escape")
	}

	if takeoff {
		fmt.Println("Split due to takeoff")
	}

	if nonStandardCategoryFinish {
		fmt.Println("Split due to non standard category finish")
	}

	return pickup || unlock || beam || energyUpgrade || roomTransitions || minibossDefeat || bossDefeat || escape || takeoff || nonStandardCategoryFinish
}

const NUM_LATENCY_SAMPLES = 10

type SNESState struct {
	vars                     map[string]*MemoryWatcher
	pickedUpHundredthMissile bool
	pickedUpSporeSpawnSuper  bool
	latencySamples           []uint128
	data                     []byte
	doExtraUpdate            bool
	mu                       sync.Mutex
}

type uint128 struct {
	hi uint64
	lo uint64
}

func (a uint128) Add(b uint128) uint128 {
	lo := a.lo + b.lo
	hi := a.hi + b.hi
	if lo < a.lo {
		hi++
	}
	return uint128{hi: hi, lo: lo}
}

func (a uint128) Sub(b uint128) uint128 {
	lo := a.lo - b.lo
	hi := a.hi - b.hi
	if a.lo < b.lo {
		hi--
	}
	return uint128{hi: hi, lo: lo}
}

func (a uint128) ToFloat64() float64 {
	return float64(a.hi)*math.Pow(2, 64) + float64(a.lo)
}

func NewSNESState() *SNESState {
	data := make([]byte, 0x10000)
	vars := map[string]*MemoryWatcher{
		// Word
		"controller":    NewMemoryWatcher(0x008B, Word),
		"roomID":        NewMemoryWatcher(0x079B, Word),
		"enemyHP":       NewMemoryWatcher(0x0F8C, Word),
		"shipAI":        NewMemoryWatcher(0x0FB2, Word),
		"motherBrainHP": NewMemoryWatcher(0x0FCC, Word),
		// Byte
		"mapInUse":          NewMemoryWatcher(0x079F, Byte),
		"gameState":         NewMemoryWatcher(0x0998, Byte),
		"unlockedEquips2":   NewMemoryWatcher(0x09A4, Byte),
		"unlockedEquips":    NewMemoryWatcher(0x09A5, Byte),
		"unlockedBeams":     NewMemoryWatcher(0x09A8, Byte),
		"unlockedCharge":    NewMemoryWatcher(0x09A9, Byte),
		"maxEnergy":         NewMemoryWatcher(0x09C4, Word),
		"maxMissiles":       NewMemoryWatcher(0x09C8, Byte),
		"maxSupers":         NewMemoryWatcher(0x09CC, Byte),
		"maxPowerBombs":     NewMemoryWatcher(0x09D0, Byte),
		"maxReserve":        NewMemoryWatcher(0x09D4, Word),
		"igtFrames":         NewMemoryWatcher(0x09DA, Byte),
		"igtSeconds":        NewMemoryWatcher(0x09DC, Byte),
		"igtMinutes":        NewMemoryWatcher(0x09DE, Byte),
		"igtHours":          NewMemoryWatcher(0x09E0, Byte),
		"playerState":       NewMemoryWatcher(0x0A28, Byte),
		"eventFlags":        NewMemoryWatcher(0xD821, Byte),
		"crateriaBosses":    NewMemoryWatcher(0xD828, Byte),
		"brinstarBosses":    NewMemoryWatcher(0xD829, Byte),
		"norfairBosses":     NewMemoryWatcher(0xD82A, Byte),
		"wreckedShipBosses": NewMemoryWatcher(0xD82B, Byte),
		"maridiaBosses":     NewMemoryWatcher(0xD82C, Byte),
		"tourianBosses":     NewMemoryWatcher(0xD82D, Byte),
		"ceresBosses":       NewMemoryWatcher(0xD82E, Byte),
		"crateriaItems":     NewMemoryWatcher(0xD870, Byte),
		"brinteriaItems":    NewMemoryWatcher(0xD871, Byte),
		"brinstarItems2":    NewMemoryWatcher(0xD872, Byte),
		"brinstarItems3":    NewMemoryWatcher(0xD873, Byte),
		"brinstarItems4":    NewMemoryWatcher(0xD874, Byte),
		"brinstarItems5":    NewMemoryWatcher(0xD875, Byte),
		"norfairItems1":     NewMemoryWatcher(0xD876, Byte),
		"norfairItems2":     NewMemoryWatcher(0xD877, Byte),
		"norfairItems3":     NewMemoryWatcher(0xD878, Byte),
		"norfairItems4":     NewMemoryWatcher(0xD879, Byte),
		"norfairItems5":     NewMemoryWatcher(0xD87A, Byte),
		"wreckedShipItems":  NewMemoryWatcher(0xD880, Byte),
		"maridiaItems1":     NewMemoryWatcher(0xD881, Byte),
		"maridiaItems2":     NewMemoryWatcher(0xD882, Byte),
		"maridiaItems3":     NewMemoryWatcher(0xD883, Byte),
	}
	return &SNESState{
		doExtraUpdate:            true,
		data:                     data,
		latencySamples:           make([]uint128, 0),
		pickedUpHundredthMissile: false,
		pickedUpSporeSpawnSuper:  false,
		vars:                     vars,
	}
}

func (mw MemoryWatcher) ptr() *MemoryWatcher {
	return &mw
}

func (s *SNESState) update() {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, watcher := range s.vars {
		if s.doExtraUpdate {
			watcher.UpdateValue(s.data)
			s.doExtraUpdate = false
		}
		watcher.UpdateValue(s.data)
	}
}

type SNESSummary struct {
	LatencyAverage float64
	LatencyStddev  float64
	Start          bool
	Reset          bool
	Split          bool
}

func (s *SNESState) FetchAll(client SyncClient, settings *Settings) (*SNESSummary, error) {
	startTime := time.Now()
	addresses := [][2]int{
		{0xF5008B, 2},  // Controller 1 Input
		{0xF5079B, 3},  // ROOM ID + ROOM # for region + Region Number
		{0xF50998, 1},  // GAME STATE
		{0xF509A4, 61}, // ITEMS
		{0xF50A28, 1},
		{0xF50F8C, 66},
		{0xF5D821, 14},
		{0xF5D870, 20},
	}
	snesData, err := client.getAddresses(addresses)
	if err != nil {
		return nil, err
	}

	copy(s.data[0x008B:0x008B+2], snesData[0])
	copy(s.data[0x079B:0x079B+3], snesData[1])
	s.data[0x0998] = snesData[2][0]
	copy(s.data[0x09A4:0x09A4+61], snesData[3])
	s.data[0x0A28] = snesData[4][0]
	copy(s.data[0x0F8C:0x0F8C+66], snesData[5])
	copy(s.data[0xD821:0xD821+14], snesData[6])
	copy(s.data[0xD870:0xD870+20], snesData[7])

	s.update()

	start := s.start()
	reset := s.reset()
	split := split(settings, s)

	elapsed := time.Since(startTime).Milliseconds()

	if len(s.latencySamples) == NUM_LATENCY_SAMPLES {
		s.latencySamples = s.latencySamples[1:]
	}
	s.latencySamples = append(s.latencySamples, uint128FromInt(elapsed))

	averageLatency := averageUint128Slice(s.latencySamples)

	var sdevSum float64
	for _, x := range s.latencySamples {
		diff := x.ToFloat64() - averageLatency
		sdevSum += diff * diff
	}
	stddev := math.Sqrt(sdevSum / float64(len(s.latencySamples)-1))

	return &SNESSummary{
		LatencyAverage: averageLatency,
		LatencyStddev:  stddev,
		Start:          start,
		Reset:          reset,
		Split:          split,
	}, nil
}

func uint128FromInt(i int64) uint128 {
	if i < 0 {
		return uint128{hi: math.MaxUint64, lo: uint64(i)}
	}
	return uint128{hi: 0, lo: uint64(i)}
}

func averageUint128Slice(arr []uint128) float64 {
	var sum uint128
	for _, v := range arr {
		sum = sum.Add(v)
	}
	return sum.ToFloat64() / float64(len(arr))
}

func (s *SNESState) start() bool {
	normalStart := s.vars["gameState"].old == 2 && s.vars["gameState"].current == 0x1f
	cutsceneEnded := s.vars["gameState"].old == 0x1E && s.vars["gameState"].current == 0x1F
	zebesStart := s.vars["gameState"].old == 5 && s.vars["gameState"].current == 6
	return normalStart || cutsceneEnded || zebesStart
}

func (s *SNESState) reset() bool {
	return s.vars["roomID"].old != 0 && s.vars["roomID"].current == 0
}

type TimeSpan struct {
	seconds float64
}

func (t TimeSpan) Seconds() float64 {
	return t.seconds
}

func TimeSpanFromSeconds(seconds float64) TimeSpan {
	return TimeSpan{seconds: seconds}
}

func (s *SNESState) gametimeToSeconds() TimeSpan {
	hours := float64(s.vars["igtHours"].current)
	minutes := float64(s.vars["igtMinutes"].current)
	seconds := float64(s.vars["igtSeconds"].current)

	totalSeconds := hours*3600 + minutes*60 + seconds
	return TimeSpanFromSeconds(totalSeconds)
}

type SuperMetroidAutoSplitter struct {
	snes         *SNESState
	settings     *sync.RWMutex
	settingsData *Settings
}

func NewSuperMetroidAutoSplitter(settings *sync.RWMutex, settingsData *Settings) *SuperMetroidAutoSplitter {
	return &SuperMetroidAutoSplitter{
		snes:         NewSNESState(),
		settings:     settings,
		settingsData: settingsData,
	}
}

func (a *SuperMetroidAutoSplitter) Update(client SyncClient) (*SNESSummary, error) {
	return a.snes.FetchAll(client, a.settingsData)
}

func (a *SuperMetroidAutoSplitter) GametimeToSeconds() *TimeSpan {
	t := a.snes.gametimeToSeconds()
	return &t
}

func (a *SuperMetroidAutoSplitter) ResetGameTracking() {
	a.snes = NewSNESState()
}
