package models

var ValidTrucks = []string{"Tulip", "Watson", "Libby", "Andre350", "Magnolia"}

var ValidTeams = []string{
	"urban_trees", "beltline", "neighborwoods", "forest_restoration",
	"education", "admin", "volunteer_services", "workforce_development",
	"downtown_planting", "floaters",
}

func IsValidTruck(name string) bool {
	for _, t := range ValidTrucks {
		if name == t {
			return true
		}
	}
	return false
}

func IsValidTeam(name string) bool {
	for _, t := range ValidTeams {
		if name == t {
			return true
		}
	}
	return false
}
