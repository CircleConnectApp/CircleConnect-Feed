package controllers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/CircleConnectApp/feed-service/database"
	"github.com/CircleConnectApp/feed-service/models"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type FeedController struct {
	mongoDB *mongo.Database
	pgDB    *sql.DB
	config  struct {
		UserServiceURL      string
		PostServiceURL      string
		CommunityServiceURL string
	}
}

// NewFeedController creates a new instance of FeedController
func NewFeedController(mongoDB *mongo.Database, pgDB *sql.DB, userServiceURL, postServiceURL, communityServiceURL string) *FeedController {
	return &FeedController{
		mongoDB: mongoDB,
		pgDB:    pgDB,
		config: struct {
			UserServiceURL      string
			PostServiceURL      string
			CommunityServiceURL string
		}{
			UserServiceURL:      userServiceURL,
			PostServiceURL:      postServiceURL,
			CommunityServiceURL: communityServiceURL,
		},
	}
}

// GetFeed retrieves the personalized feed for the authenticated user
func (fc *FeedController) GetFeed(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var query models.FeedQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set default values
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.Limit <= 0 {
		query.Limit = 20
	}

	// Get user preferences
	pref, err := fc.getUserPreferences(userID.(int))
	if err != nil && err != mongo.ErrNoDocuments {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user preferences"})
		return
	}

	// If no sort method specified, use the one from user preferences
	if query.SortBy == "" && pref != nil {
		query.SortBy = pref.FeedSortMethod
	}

	// Default to date sorting if no preference is set
	if query.SortBy == "" {
		query.SortBy = "date"
	}

	// Get joined communities for the user
	joinedCommunities, err := fc.getJoinedCommunities(userID.(int))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get joined communities"})
		return
	}

	// Build the feed items based on user preferences and joined communities
	feed, err := fc.buildFeed(userID.(int), joinedCommunities, pref, query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to build feed"})
		return
	}

	c.JSON(http.StatusOK, feed)
}

// GetRecommendedPosts retrieves recommended posts for the user
func (fc *FeedController) GetRecommendedPosts(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var query models.FeedQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set default values
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.Limit <= 0 {
		query.Limit = 20
	}

	// Get user preferences and demographic info
	pref, err := fc.getUserPreferences(userID.(int))
	if err != nil && err != mongo.ErrNoDocuments {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user preferences"})
		return
	}

	// Default to relevance for recommendations
	query.SortBy = "relevance"

	// Get user's demographics from profile service
	userInfo, err := fc.getUserInfo(userID.(int))
	if err != nil {
		log.Printf("Warning: Failed to get user info: %v", err)
		// Continue without user info
	}

	// Build recommended feed based on user preferences, demographics, and post popularity
	feed, err := fc.buildRecommendedFeed(userID.(int), pref, userInfo, query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to build recommended feed"})
		return
	}

	c.JSON(http.StatusOK, feed)
}

// GetUserPreferences retrieves the feed preferences for the authenticated user
func (fc *FeedController) GetUserPreferences(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	pref, err := fc.getUserPreferences(userID.(int))
	if err != nil {
		if err == mongo.ErrNoDocuments {
			// Return default preferences if not found
			pref = &models.UserPreference{
				UserID:         userID.(int),
				FeedSortMethod: "date",
				UpdatedAt:      time.Now(),
			}
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user preferences"})
			return
		}
	}

	c.JSON(http.StatusOK, pref)
}

// UpdateUserPreferences updates the feed preferences for the authenticated user
func (fc *FeedController) UpdateUserPreferences(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req models.UpdatePreferenceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get current preferences
	pref, err := fc.getUserPreferences(userID.(int))
	if err != nil && err != mongo.ErrNoDocuments {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user preferences"})
		return
	}

	// Create new preferences if not found
	if err == mongo.ErrNoDocuments || pref == nil {
		pref = &models.UserPreference{
			UserID:         userID.(int),
			FeedSortMethod: "date",
			UpdatedAt:      time.Now(),
		}
	}

	// Update fields if provided
	if req.FeedSortMethod != "" {
		pref.FeedSortMethod = req.FeedSortMethod
	}
	if req.PreferedTags != nil {
		pref.PreferedTags = req.PreferedTags
	}
	if req.ExcludedTags != nil {
		pref.ExcludedTags = req.ExcludedTags
	}
	if req.PreferedCommunities != nil {
		pref.PreferedCommunities = req.PreferedCommunities
	}
	pref.UpdatedAt = time.Now()

	// Save preferences
	err = fc.saveUserPreferences(pref)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save user preferences"})
		return
	}

	c.JSON(http.StatusOK, pref)
}

// Helper functions

// getUserPreferences retrieves a user's feed preferences
func (fc *FeedController) getUserPreferences(userID int) (*models.UserPreference, error) {
	collection := fc.mongoDB.Collection(database.PreferencesCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var pref models.UserPreference
	err := collection.FindOne(ctx, bson.M{"user_id": userID}).Decode(&pref)
	if err != nil {
		return nil, err
	}

	return &pref, nil
}

// saveUserPreferences saves a user's feed preferences
func (fc *FeedController) saveUserPreferences(pref *models.UserPreference) error {
	collection := fc.mongoDB.Collection(database.PreferencesCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if pref.ID.IsZero() {
		pref.ID = primitive.NewObjectID()
		_, err := collection.InsertOne(ctx, pref)
		return err
	}

	_, err := collection.ReplaceOne(ctx, bson.M{"_id": pref.ID}, pref)
	return err
}

// getJoinedCommunities retrieves the communities joined by a user
func (fc *FeedController) getJoinedCommunities(userID int) ([]int, error) {
	url := fmt.Sprintf("%s/user/%d/communities", fc.config.CommunityServiceURL, userID)

	// Make HTTP request to community service
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get joined communities: status %d", resp.StatusCode)
	}

	var result struct {
		Communities []struct {
			ID int `json:"id"`
		} `json:"communities"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	var communityIDs []int
	for _, community := range result.Communities {
		communityIDs = append(communityIDs, community.ID)
	}

	return communityIDs, nil
}

// getUserInfo retrieves user demographic information
func (fc *FeedController) getUserInfo(userID int) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/users/%d", fc.config.UserServiceURL, userID)

	// Make HTTP request to user service
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get user info: status %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}

// buildFeed constructs a personalized feed for the user
func (fc *FeedController) buildFeed(userID int, communities []int, pref *models.UserPreference, query models.FeedQuery) (*models.Feed, error) {
	// Get posts from the post service
	postsURL := fmt.Sprintf("%s/posts", fc.config.PostServiceURL)

	// Add query parameters
	if query.CommunityID > 0 {
		postsURL += fmt.Sprintf("?community_id=%d", query.CommunityID)
	} else if len(communities) > 0 {
		// If community ID not specified, use all user's joined communities
		postsURL += "?community_id=" + strconv.Itoa(communities[0])
		for i := 1; i < len(communities); i++ {
			postsURL += "," + strconv.Itoa(communities[i])
		}
	}

	// Add pagination
	if query.Page > 0 && query.Limit > 0 {
		if postsURL == fc.config.PostServiceURL+"/posts" {
			postsURL += "?"
		} else {
			postsURL += "&"
		}
		postsURL += fmt.Sprintf("page=%d&limit=%d", query.Page, query.Limit)
	}

	// Make HTTP request to post service
	resp, err := http.Get(postsURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get posts: status %d", resp.StatusCode)
	}

	// Parse response
	var postsResp struct {
		Posts []struct {
			ID          string    `json:"id"`
			UserID      int       `json:"user_id"`
			CommunityID int       `json:"community_id"`
			Title       string    `json:"title"`
			Content     string    `json:"content"`
			MediaURLs   []string  `json:"media_urls"`
			Tags        []string  `json:"tags"`
			CreatedAt   time.Time `json:"created_at"`
			LikeCount   int       `json:"like_count"`
			UserName    string    `json:"user_name"`
			UserPic     string    `json:"user_pic"`
		} `json:"posts"`
		Total int `json:"total"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&postsResp); err != nil {
		return nil, err
	}

	// Convert to feed items
	feedItems := make([]models.FeedItem, 0, len(postsResp.Posts))
	for _, post := range postsResp.Posts {
		postID, err := primitive.ObjectIDFromHex(post.ID)
		if err != nil {
			// Skip invalid post IDs
			continue
		}

		// Calculate relevance score
		relevance := calculateRelevance(post.LikeCount, post.CreatedAt, pref)

		feedItem := models.FeedItem{
			ID:           primitive.NewObjectID(),
			PostID:       postID,
			UserID:       post.UserID,
			CommunityID:  post.CommunityID,
			Title:        post.Title,
			Content:      post.Content,
			LikeCount:    post.LikeCount,
			MediaURLs:    post.MediaURLs,
			Tags:         post.Tags,
			CreatedAt:    post.CreatedAt,
			Relevance:    relevance,
			AuthorName:   post.UserName,
			AuthorAvatar: post.UserPic,
		}
		feedItems = append(feedItems, feedItem)
	}

	// Sort feed items based on query
	sortFeedItems(feedItems, query.SortBy)

	return &models.Feed{
		Items: feedItems,
		Total: postsResp.Total,
		Page:  query.Page,
		Limit: query.Limit,
	}, nil
}

// buildRecommendedFeed constructs a recommended feed for the user
func (fc *FeedController) buildRecommendedFeed(userID int, pref *models.UserPreference, userInfo map[string]interface{}, query models.FeedQuery) (*models.Feed, error) {
	// Get popular posts
	postsURL := fmt.Sprintf("%s/posts?sort=popular&page=%d&limit=%d", fc.config.PostServiceURL, query.Page, query.Limit)

	// Make HTTP request to post service
	resp, err := http.Get(postsURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get posts: status %d", resp.StatusCode)
	}

	// Parse response
	var postsResp struct {
		Posts []struct {
			ID          string    `json:"id"`
			UserID      int       `json:"user_id"`
			CommunityID int       `json:"community_id"`
			Title       string    `json:"title"`
			Content     string    `json:"content"`
			MediaURLs   []string  `json:"media_urls"`
			Tags        []string  `json:"tags"`
			CreatedAt   time.Time `json:"created_at"`
			LikeCount   int       `json:"like_count"`
			UserName    string    `json:"user_name"`
			UserPic     string    `json:"user_pic"`
		} `json:"posts"`
		Total int `json:"total"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&postsResp); err != nil {
		return nil, err
	}

	// Convert to feed items
	feedItems := make([]models.FeedItem, 0, len(postsResp.Posts))
	for _, post := range postsResp.Posts {
		postID, err := primitive.ObjectIDFromHex(post.ID)
		if err != nil {
			// Skip invalid post IDs
			continue
		}

		// Calculate relevance score with user demographics factored in
		relevance := calculateRecommendationScore(post, userInfo, pref)

		feedItem := models.FeedItem{
			ID:           primitive.NewObjectID(),
			PostID:       postID,
			UserID:       post.UserID,
			CommunityID:  post.CommunityID,
			Title:        post.Title,
			Content:      post.Content,
			LikeCount:    post.LikeCount,
			MediaURLs:    post.MediaURLs,
			Tags:         post.Tags,
			CreatedAt:    post.CreatedAt,
			Relevance:    relevance,
			AuthorName:   post.UserName,
			AuthorAvatar: post.UserPic,
		}
		feedItems = append(feedItems, feedItem)
	}

	// Sort by relevance
	sortFeedItems(feedItems, "relevance")

	return &models.Feed{
		Items: feedItems,
		Total: postsResp.Total,
		Page:  query.Page,
		Limit: query.Limit,
	}, nil
}

// calculateRelevance calculates a post's relevance score
func calculateRelevance(likeCount int, createdAt time.Time, pref *models.UserPreference) float64 {
	// Simple algorithm based on popularity and recency
	recencyFactor := 1.0 / (1.0 + float64(time.Since(createdAt).Hours()/24.0)) // Higher value for more recent posts
	popularityFactor := float64(likeCount) / 100.0                             // Normalize like count

	// Combine factors with weights
	relevanceScore := (0.7 * recencyFactor) + (0.3 * popularityFactor)

	return relevanceScore
}

// calculateRecommendationScore calculates a recommendation score for a post
func calculateRecommendationScore(post struct {
	ID          string    `json:"id"`
	UserID      int       `json:"user_id"`
	CommunityID int       `json:"community_id"`
	Title       string    `json:"title"`
	Content     string    `json:"content"`
	MediaURLs   []string  `json:"media_urls"`
	Tags        []string  `json:"tags"`
	CreatedAt   time.Time `json:"created_at"`
	LikeCount   int       `json:"like_count"`
	UserName    string    `json:"user_name"`
	UserPic     string    `json:"user_pic"`
}, userInfo map[string]interface{}, pref *models.UserPreference) float64 {
	// Calculate base relevance
	baseRelevance := calculateRelevance(post.LikeCount, post.CreatedAt, pref)

	// Add additional factors based on user preferences
	additionalScore := 0.0

	// Preferred tags match
	if pref != nil && len(pref.PreferedTags) > 0 {
		for _, tag := range post.Tags {
			for _, prefTag := range pref.PreferedTags {
				if tag == prefTag {
					additionalScore += 0.2 // Boost for each matching preferred tag
					break
				}
			}
		}
	}

	// Preferred communities match
	if pref != nil && len(pref.PreferedCommunities) > 0 {
		for _, prefCommunity := range pref.PreferedCommunities {
			if post.CommunityID == prefCommunity {
				additionalScore += 0.3 // Boost for preferred community
				break
			}
		}
	}

	// User demographics (gender, interests, etc.)
	if userInfo != nil {
		// Here you would implement logic to boost posts based on demographic matching
		// This is a simplified placeholder - in a real system, this would be more sophisticated
		if gender, ok := userInfo["gender"].(string); ok {
			// Example logic: if the post has tags that typically appeal to this gender
			// This is a very simplistic example and should be replaced with a more nuanced approach
			for _, tag := range post.Tags {
				if (gender == "male" && (tag == "sports" || tag == "gaming")) ||
					(gender == "female" && (tag == "fashion" || tag == "beauty")) {
					additionalScore += 0.1
				}
			}
		}
	}

	return baseRelevance + additionalScore
}

// sortFeedItems sorts feed items based on the specified method
func sortFeedItems(items []models.FeedItem, sortBy string) {
	switch sortBy {
	case "date":
		// Sort by creation date (newest first)
		options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	case "popular":
		// Sort by like count (highest first)
		options.Find().SetSort(bson.D{{Key: "like_count", Value: -1}})
	case "relevance":
		// Sort by relevance score (highest first)
		options.Find().SetSort(bson.D{{Key: "relevance", Value: -1}})
	default:
		// Default to date sorting
		options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	}
}
