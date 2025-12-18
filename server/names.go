package main

import (
	"fmt"
	"math/rand"
	"time"
)

var gigaChadNames = []string{
	"Apex Predator",
	"Iron Lifter",
	"Chad Thundercock",
	"Zyzz Legacy",
	"Gigachad Prime",
	"Alpha Male",
	"Sigma Grindset",
	"Based God",
	"Gym Rat",
	"Protein Shake",
	"Creatine King",
	"Bench Presser",
	"Squat Rack",
	"Deadlift Demon",
	"Muscle Mommy",
	"Swole Patrol",
	"Gains Goblin",
	"Pump Chaser",
	"Flex Flexer",
	"Ripped Rick",
	"Shredded Sam",
	"Buff Brad",
	"Jacked Jack",
	"Yoked Yoda",
	"Bicep Beast",
	"Tricep Titan",
	"Quad Zilla",
	"Calf King",
	"Abs of Steel",
	"Pectoral Prince",
	"Lat Legend",
	"Trap God",
	"Deltoid Duke",
	"Glute God",
	"Hamstring Hero",
	"Forearm Freak",
	"Neck Necromancer",
	"Wrist Warrior",
	"Ankle Anchor",
	"Toe Titan",
	"Finger Flexer",
	"Thumb Thrasher",
	"Knuckle Knight",
	"Palm Paladin",
	"Elbow Emperor",
	"Shoulder Shogun",
	"Knee Knight",
	"Hip Hero",
	"Waist Warrior",
	"Chest Champion",
}

// GenerateGigaChadName returns a random name from the list.
// If existingNames is provided, it tries to find a unique name.
// If all names are taken, it appends a number.
func GenerateGigaChadName(existingNames map[string]bool) string {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	
	// Try 10 times to find a random unique name
	for i := 0; i < 10; i++ {
		name := gigaChadNames[rng.Intn(len(gigaChadNames))]
		if !existingNames[name] {
			return name
		}
	}

	// Fallback: Pick a random name and append a number until unique
	baseName := gigaChadNames[rng.Intn(len(gigaChadNames))]
	counter := 2
	for {
		name := fmt.Sprintf("%s %d", baseName, counter)
		if !existingNames[name] {
			return name
		}
		counter++
	}
}
