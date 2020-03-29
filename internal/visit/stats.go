package visit

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type GetStatsParams struct {
	DB          *mongo.Client
	Environment string
	ShortID     primitive.ObjectID
}

func (p GetStatsParams) validate() error {
	if p.DB == nil {
		return fmt.Errorf("missing db client")
	}
	if p.Environment == "" {
		return fmt.Errorf("invalid env")
	}
	if len(p.ShortID) == 0 {
		return fmt.Errorf("missing document id")
	}

	return nil
}

// GetStats assembles Stats for a short URL by aggregating Visit
// documents.
func GetStats(ctx context.Context, p GetStatsParams) (stats Stats, err error) {
	if err := p.validate(); err != nil {
		return stats, fmt.Errorf("invalid params: %v", err)
	}

	coll := p.DB.Database(p.Environment).Collection(Collection)

	// Create the aggregation.
	now := time.Now()
	oneDayAgo := now.AddDate(0, 0, -1)
	oneWeekAgo := now.AddDate(0, 0, -7)
	oneYearAgo := now.AddDate(-1, 0, 0)

	// $cond follows this format: [if,then,else].
	// We return a count of 1 for each document that was created
	// within the provided timespan. Otherwise, we return 0,
	// to ensure that the document is not counted as a visit.
	matchDay := bson.M{"$cond": []interface{}{
		bson.M{"$gte": []interface{}{"$time", oneDayAgo}},
		1,
		0,
	}}
	matchWeek := bson.M{"$cond": []interface{}{
		bson.M{"$gte": []interface{}{"$time", oneWeekAgo}},
		1,
		0,
	}}
	matchYear := bson.M{"$cond": []interface{}{
		bson.M{"$gte": []interface{}{"$time", oneYearAgo}},
		1,
		0,
	}}

	// Match documents created within the last year for this URL.
	// Then, group them by timestamp, incrementing by the value
	// specified in the above conditions.
	pipeline := []bson.M{
		{"$match": bson.M{
			"shortId": p.ShortID,
			"time":    bson.M{"$gt": oneYearAgo},
		}},
		{"$group": bson.M{
			"_id":  "stats",
			"day":  bson.M{"$sum": matchDay},
			"week": bson.M{"$sum": matchWeek},
			"year": bson.M{"$sum": matchYear},
		}},
	}

	// Execute the aggregation, and decode the results.
	aggregateContext, cancel := context.WithTimeout(ctx, aggregateTimeout)
	defer cancel()

	cursor, err := coll.Aggregate(aggregateContext, pipeline)
	if err != nil {
		return stats, fmt.Errorf("failed to aggregate stats: %v", err)
	}

	var res []Stats
	if err = cursor.All(ctx, &res); err != nil {
		return stats, fmt.Errorf("failed to decode stats: %v", err)
	}

	// If no document was returned, there were no visits.
	// Return the default Stats value, with zero visits as default.
	if len(res) == 0 {
		return stats, nil
	}

	// Otherwise, return the assembled statistics.
	// The aggregation should only return a single stats document,
	// since the final stage is $group.
	return res[0], nil
}
