package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// FeedItem represents a post in the user's feed
type FeedItem struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	PostID       primitive.ObjectID `bson:"post_id" json:"post_id"`
	UserID       int                `bson:"user_id" json:"user_id"`
	CommunityID  int                `bson:"community_id" json:"community_id"`
	Title        string             `bson:"title" json:"title"`
	Content      string             `bson:"content" json:"content"`
	LikeCount    int                `bson:"like_count" json:"like_count"`
	MediaURLs    []string           `bson:"media_urls,omitempty" json:"media_urls,omitempty"`
	Tags         []string           `bson:"tags,omitempty" json:"tags,omitempty"`
	CreatedAt    time.Time          `bson:"created_at" json:"created_at"`
	Relevance    float64            `bson:"relevance" json:"relevance"`
	AuthorName   string             `bson:"author_name" json:"author_name"`
	AuthorAvatar string             `bson:"author_avatar,omitempty" json:"author_avatar,omitempty"`
}

// UserPreference stores a user's feed preferences
type UserPreference struct {
	ID                  primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID              int                `bson:"user_id" json:"user_id"`
	FeedSortMethod      string             `bson:"feed_sort_method" json:"feed_sort_method"` // "date", "relevance", "popular"
	PreferedTags        []string           `bson:"prefered_tags,omitempty" json:"prefered_tags,omitempty"`
	ExcludedTags        []string           `bson:"excluded_tags,omitempty" json:"excluded_tags,omitempty"`
	PreferedCommunities []int              `bson:"prefered_communities,omitempty" json:"prefered_communities,omitempty"`
	UpdatedAt           time.Time          `bson:"updated_at" json:"updated_at"`
}

// Feed represents a collection of FeedItems
type Feed struct {
	Items []FeedItem `json:"items"`
	Total int        `json:"total"`
	Page  int        `json:"page"`
	Limit int        `json:"limit"`
}

// FeedQuery represents query parameters for feed retrieval
type FeedQuery struct {
	SortBy      string   `form:"sort_by" json:"sort_by"` // date, relevance, popular
	Tags        []string `form:"tags" json:"tags"`
	CommunityID int      `form:"community_id" json:"community_id"`
	Page        int      `form:"page" json:"page"`
	Limit       int      `form:"limit" json:"limit"`
}

// UpdatePreferenceRequest is used to update user preferences
type UpdatePreferenceRequest struct {
	FeedSortMethod      string   `json:"feed_sort_method,omitempty"`
	PreferedTags        []string `json:"prefered_tags,omitempty"`
	ExcludedTags        []string `json:"excluded_tags,omitempty"`
	PreferedCommunities []int    `json:"prefered_communities,omitempty"`
}
