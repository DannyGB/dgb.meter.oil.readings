package database

import (
	"fmt"
	"strconv"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PageParams struct {
	Skip          int
	Take          int
	SortDirection string
	Filter        string
}

func (params *PageParams) getOrFilters() []bson.M {
	orFilter := []bson.M{}

	if params.Filter != "" {

		orFilter = []bson.M{
			{"rate": bson.D{{
				Key: "$regex", Value: primitive.Regex{Pattern: fmt.Sprintf("%s.*", params.Filter), Options: "i"},
			}}},
			{"note": bson.D{{
				Key: "$regex", Value: primitive.Regex{Pattern: fmt.Sprintf("%s.*", params.Filter), Options: "i"},
			}}},
			{"readingdate": bson.D{{
				Key: "$regex", Value: primitive.Regex{Pattern: fmt.Sprintf("%s.*", params.Filter), Options: "i"},
			}}},
		}
	}

	filterConverted, convErr := strconv.Atoi(params.Filter)

	if convErr == nil {
		return append(orFilter, bson.M{"reading": filterConverted})
	}

	return orFilter
}

func (params *PageParams) GetFilters() bson.M {
	if params.Filter != "" {
		return bson.M{
			"$or": params.getOrFilters(),
		}
	}

	return bson.M{}
}

func (params *PageParams) GetSortDirection() int {
	if params.SortDirection == "desc" {
		return -1
	}

	return 1
}
