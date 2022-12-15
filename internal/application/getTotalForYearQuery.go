package application

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func getTotalForYear(result []primitive.M) float64 {
	var first, last float64 = 0, 0
	firstDay := false

	for _, val := range result {
		if !firstDay {
			first = val["reading"].(float64)
			firstDay = true
		}

		last = val["reading"].(float64)
	}

	return (last - first)
}
