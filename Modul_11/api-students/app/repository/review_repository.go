package repository

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"api-students/app/model"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var (
	ErrInvalidID    = errors.New("format id tidak valid")
	ErrNotOwner     = errors.New("bukan pemilik data")
	ErrOwnReview    = errors.New("tidak boleh menilai ulasan sendiri")
	ErrAlreadyVoted = errors.New("sudah pernah menilai ulasan ini")
	ErrReplyLimit   = errors.New("jumlah balasan sudah mencapai batas")
)

const MaxRepliesPerReview = 50

type ReviewRepository interface {
	EnsureIndexes(ctx context.Context) error
	Create(ctx context.Context, r model.Review) (model.Review, error)
	FindByID(ctx context.Context, id string) (model.Review, error)
	FindByCourse(ctx context.Context, q model.ReviewQuery) ([]model.Review, error)
	UpdateOwn(ctx context.Context, id string, userID int, req model.UpdateReviewRequest) (model.Review, error)
	Delete(ctx context.Context, id string, userID int, anyOwner bool) error
	AddReply(ctx context.Context, id string, reply model.Reply) (model.Review, error)
	MarkHelpful(ctx context.Context, id string, userID int) (model.Review, error)
	Summary(ctx context.Context, courseID int) (model.ReviewSummary, error)
}

type reviewDocument struct {
	ID           bson.ObjectID   `bson:"_id,omitempty"`
	CourseID     int             `bson:"course_id"`
	UserID       int             `bson:"user_id"`
	Username     string          `bson:"username"`
	Rating       int             `bson:"rating"`
	Comment      string          `bson:"comment"`
	Tags         []string        `bson:"tags"`
	Replies      []replyDocument `bson:"replies"`
	HelpfulCount int             `bson:"helpful_count"`
	HelpfulBy    []int           `bson:"helpful_by"`
	CreatedAt    time.Time       `bson:"created_at"`
	UpdatedAt    time.Time       `bson:"updated_at"`
}

type replyDocument struct {
	ID        bson.ObjectID `bson:"_id"`
	UserID    int           `bson:"user_id"`
	Username  string        `bson:"username"`
	Body      string        `bson:"body"`
	CreatedAt time.Time     `bson:"created_at"`
}

func (d reviewDocument) toModel() model.Review {
	replies := make([]model.Reply, 0, len(d.Replies))
	for _, r := range d.Replies {
		replies = append(replies, model.Reply{
			ID: r.ID.Hex(), UserID: r.UserID, Username: r.Username,
			Body: r.Body, CreatedAt: r.CreatedAt,
		})
	}

	tags := d.Tags
	if tags == nil {
		tags = []string{}
	}
	return model.Review{
		ID: d.ID.Hex(), CourseID: d.CourseID, UserID: d.UserID, Username: d.Username,
		Rating: d.Rating, Comment: d.Comment, Tags: tags, Replies: replies,
		HelpfulCount: d.HelpfulCount, CreatedAt: d.CreatedAt, UpdatedAt: d.UpdatedAt,
	}
}

type reviewMongoRepository struct {
	coll *mongo.Collection
}

func NewReviewRepository(db *mongo.Database) ReviewRepository {
	return &reviewMongoRepository{coll: db.Collection("course_reviews")}
}

func now() time.Time { return time.Now().UTC().Truncate(time.Millisecond) }

func parseID(id string) (bson.ObjectID, error) {
	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return bson.ObjectID{}, ErrInvalidID
	}
	return oid, nil
}

func (r *reviewMongoRepository) EnsureIndexes(ctx context.Context) error {
	_, err := r.coll.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "course_id", Value: 1}, {Key: "user_id", Value: 1}},
			Options: options.Index().SetUnique(true).SetName("course_user_unique"),
		},
		{
			Keys:    bson.D{{Key: "course_id", Value: 1}, {Key: "_id", Value: -1}},
			Options: options.Index().SetName("course_newest"),
		},
		{
			Keys:    bson.D{{Key: "course_id", Value: 1}, {Key: "tags", Value: 1}},
			Options: options.Index().SetName("course_tags"),
		},
	})
	if err != nil {
		return fmt.Errorf("membuat index course_reviews: %w", err)
	}
	return nil
}

func (r *reviewMongoRepository) Create(ctx context.Context, rv model.Review) (model.Review, error) {
	t := now()
	doc := reviewDocument{
		ID:           bson.NewObjectID(),
		CourseID:     rv.CourseID,
		UserID:       rv.UserID,
		Username:     rv.Username,
		Rating:       rv.Rating,
		Comment:      rv.Comment,
		Tags:         rv.Tags,
		Replies:      []replyDocument{},
		HelpfulBy:    []int{},
		CreatedAt:    t,
		UpdatedAt:    t,
	}
	if doc.Tags == nil {
		doc.Tags = []string{}
	}
	if _, err := r.coll.InsertOne(ctx, doc); err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return model.Review{}, ErrDuplicate
		}
		return model.Review{}, fmt.Errorf("menyimpan ulasan: %w", err)
	}
	return doc.toModel(), nil
}

func (r *reviewMongoRepository) FindByID(ctx context.Context, id string) (model.Review, error) {
	oid, err := parseID(id)
	if err != nil {
		return model.Review{}, err
	}
	var doc reviewDocument
	if err := r.coll.FindOne(ctx, bson.M{"_id": oid}).Decode(&doc); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return model.Review{}, ErrNotFound
		}
		return model.Review{}, fmt.Errorf("mengambil ulasan: %w", err)
	}
	return doc.toModel(), nil
}

func (r *reviewMongoRepository) FindByCourse(ctx context.Context, q model.ReviewQuery) ([]model.Review, error) {
	filter := bson.M{"course_id": q.CourseID}
	if q.Tag != "" {
		filter["tags"] = q.Tag
	}
	if q.Before != "" {
		before, err := parseID(q.Before)
		if err != nil {
			return nil, err
		}
		filter["_id"] = bson.M{"$lt": before}
	}
	opts := options.Find().
		SetSort(bson.D{{Key: "_id", Value: -1}}).
		SetLimit(int64(q.Limit)).
		SetProjection(bson.M{"helpful_by": 0})

	cursor, err := r.coll.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("mengambil daftar ulasan: %w", err)
	}
	defer cursor.Close(ctx)

	var docs []reviewDocument
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, fmt.Errorf("membaca daftar ulasan: %w", err)
	}
	result := make([]model.Review, 0, len(docs))
	for _, d := range docs {
		result = append(result, d.toModel())
	}
	return result, nil
}

func (r *reviewMongoRepository) UpdateOwn(ctx context.Context, id string, userID int, req model.UpdateReviewRequest) (model.Review, error) {
	oid, err := parseID(id)
	if err != nil {
		return model.Review{}, err
	}
	set := bson.M{"updated_at": now()}
	if req.Rating != nil {
		set["rating"] = *req.Rating
	}
	if req.Comment != nil {
		set["comment"] = *req.Comment
	}
	if req.Tags != nil {
		set["tags"] = *req.Tags
	}
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var doc reviewDocument
	err = r.coll.FindOneAndUpdate(ctx,
		bson.M{"_id": oid, "user_id": userID},
		bson.M{"$set": set},
		opts,
	).Decode(&doc)
	if err == nil {
		return doc.toModel(), nil
	}
	if !errors.Is(err, mongo.ErrNoDocuments) {
		return model.Review{}, fmt.Errorf("mengubah ulasan: %w", err)
	}
	return model.Review{}, r.whyNoMatch(ctx, oid)
}

func (r *reviewMongoRepository) whyNoMatch(ctx context.Context, oid bson.ObjectID) error {
	count, err := r.coll.CountDocuments(ctx, bson.M{"_id": oid})
	if err != nil {
		return fmt.Errorf("memeriksa ulasan: %w", err)
	}
	if count == 0 {
		return ErrNotFound
	}
	return ErrNotOwner
}

func (r *reviewMongoRepository) Delete(ctx context.Context, id string, userID int, anyOwner bool) error {
	oid, err := parseID(id)
	if err != nil {
		return err
	}
	filter := bson.M{"_id": oid}
	if !anyOwner {
		filter["user_id"] = userID
	}
	res, err := r.coll.DeleteOne(ctx, filter)
	if err != nil {
		return fmt.Errorf("menghapus ulasan: %w", err)
	}
	if res.DeletedCount == 0 {
		return r.whyNoMatch(ctx, oid)
	}
	return nil
}

func (r *reviewMongoRepository) AddReply(ctx context.Context, id string, reply model.Reply) (model.Review, error) {
	oid, err := parseID(id)
	if err != nil {
		return model.Review{}, err
	}
	doc := replyDocument{
		ID:        bson.NewObjectID(),
		UserID:    reply.UserID,
		Username:  reply.Username,
		Body:      reply.Body,
		CreatedAt: now(),
	}
	limitField := "replies." + strconv.Itoa(MaxRepliesPerReview-1)
	var updated reviewDocument
	err = r.coll.FindOneAndUpdate(ctx,
		bson.M{"_id": oid, limitField: bson.M{"$exists": false}},
		bson.M{"$push": bson.M{"replies": doc}},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	).Decode(&updated)
	if err == nil {
		return updated.toModel(), nil
	}
	if !errors.Is(err, mongo.ErrNoDocuments) {
		return model.Review{}, fmt.Errorf("menambah balasan: %w", err)
	}
	count, err := r.coll.CountDocuments(ctx, bson.M{"_id": oid})
	if err != nil {
		return model.Review{}, fmt.Errorf("memeriksa ulasan: %w", err)
	}
	if count == 0 {
		return model.Review{}, ErrNotFound
	}
	return model.Review{}, ErrReplyLimit
}

func (r *reviewMongoRepository) MarkHelpful(ctx context.Context, id string, userID int) (model.Review, error) {
	oid, err := parseID(id)
	if err != nil {
		return model.Review{}, err
	}
	var updated reviewDocument
	err = r.coll.FindOneAndUpdate(ctx,
		bson.M{
			"_id":        oid,
			"user_id":    bson.M{"$ne": userID},
			"helpful_by": bson.M{"$ne": userID},
		},
		bson.M{
			"$addToSet": bson.M{"helpful_by": userID},
			"$inc":      bson.M{"helpful_count": 1},
		},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	).Decode(&updated)
	if err == nil {
		return updated.toModel(), nil
	}
	if !errors.Is(err, mongo.ErrNoDocuments) {
		return model.Review{}, fmt.Errorf("menilai ulasan: %w", err)
	}
	var current reviewDocument
	if err := r.coll.FindOne(ctx, bson.M{"_id": oid}).Decode(&current); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return model.Review{}, ErrNotFound
		}
		return model.Review{}, fmt.Errorf("memeriksa ulasan: %w", err)
	}
	if current.UserID == userID {
		return model.Review{}, ErrOwnReview
	}
	return model.Review{}, ErrAlreadyVoted
}

func (r *reviewMongoRepository) Summary(ctx context.Context, courseID int) (model.ReviewSummary, error) {
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"course_id": courseID}}},
		{{Key: "$facet", Value: bson.M{
			"stats": bson.A{
				bson.M{"$group": bson.M{
					"_id":   nil,
					"count": bson.M{"$sum": 1},
					"avg":   bson.M{"$avg": "$rating"},
				}},
			},
			"distribution": bson.A{
				bson.M{"$group": bson.M{"_id": "$rating", "total": bson.M{"$sum": 1}}},
			},
			"top_tags": bson.A{
				bson.M{"$unwind": "$tags"},
				bson.M{"$group": bson.M{"_id": "$tags", "total": bson.M{"$sum": 1}}},
				bson.M{"$sort": bson.D{{Key: "total", Value: -1}, {Key: "_id", Value: 1}}},
				bson.M{"$limit": 5},
			},
		}}},
	}

	cursor, err := r.coll.Aggregate(ctx, pipeline)
	if err != nil {
		return model.ReviewSummary{}, fmt.Errorf("meringkas ulasan: %w", err)
	}
	defer cursor.Close(ctx)

	var out []struct {
		Stats []struct {
			Count int     `bson:"count"`
			Avg   float64 `bson:"avg"`
		} `bson:"stats"`
		Distribution []struct {
			Rating int `bson:"_id"`
			Total  int `bson:"total"`
		} `bson:"distribution"`
		TopTags []struct {
			Tag   string `bson:"_id"`
			Total int    `bson:"total"`
		} `bson:"top_tags"`
	}
	if err := cursor.All(ctx, &out); err != nil {
		return model.ReviewSummary{}, fmt.Errorf("membaca ringkasan: %w", err)
	}

	summary := model.ReviewSummary{
		CourseID:     courseID,
		Distribution: map[string]int{"1": 0, "2": 0, "3": 0, "4": 0, "5": 0},
		TopTags:      []model.TagCount{},
	}
	if len(out) == 0 {
		return summary, nil
	}
	if len(out[0].Stats) > 0 {
		summary.Count = out[0].Stats[0].Count
		summary.AverageRating = float64(int(out[0].Stats[0].Avg*100+0.5)) / 100
	}
	for _, d := range out[0].Distribution {
		summary.Distribution[strconv.Itoa(d.Rating)] = d.Total
	}
	for _, t := range out[0].TopTags {
		summary.TopTags = append(summary.TopTags, model.TagCount{Tag: t.Tag, Total: t.Total})
	}
	return summary, nil
}
