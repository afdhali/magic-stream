package utils

import (
	"strconv"
	"strings"

	"github.com/afdhali/magic-stream/Backend/MagicStreamServer/models"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// PaginationParams represents pagination parameters
type PaginationParams struct {
	Limit int64
	Skip  int64
}

// ParsePaginationParams parses limit and skip from query parameters
func ParsePaginationParams(limitStr, skipStr string, defaultLimit, maxLimit int) PaginationParams {
	limit := int64(defaultLimit)
	if limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 && parsed <= maxLimit {
			limit = int64(parsed)
		}
	}

	skip := int64(0)
	if skipStr != "" {
		if parsed, err := strconv.Atoi(skipStr); err == nil && parsed >= 0 {
			skip = int64(parsed)
		}
	}

	return PaginationParams{
		Limit: limit,
		Skip:  skip,
	}
}

// BuildMovieFilter builds a MongoDB filter for movie queries
func BuildMovieFilter(genreFilter, rankingFilter string) bson.M {
	filter := bson.M{}

	if genreFilter != "" {
		filter["genre.genre_name"] = bson.M{
			"$regex":   genreFilter,
			"$options": "i",
		}
	}

	if rankingFilter != "" {
		if rankingValue, err := strconv.Atoi(rankingFilter); err == nil && rankingValue > 0 {
			filter["ranking.ranking_value"] = bson.M{"$gte": rankingValue}
		}
	}

	return filter
}

// ValidateGenres checks if genres are valid (helper function)
func ValidateGenres(genres []models.Genre) bool {
	if len(genres) == 0 {
		return false
	}

	// Check for duplicate genre IDs
	seen := make(map[int]bool)
	for _, genre := range genres {
		if genre.GenreID <= 0 || seen[genre.GenreID] {
			return false
		}
		seen[genre.GenreID] = true
	}

	return true
}

// SanitizeString trims and cleans string input
func SanitizeString(s string) string {
	return strings.TrimSpace(s)
}

// IsValidEmail performs basic email validation
func IsValidEmail(email string) bool {
	return strings.Contains(email, "@") && strings.Contains(email, ".")
}

// CalculatePaginationInfo calculates pagination metadata
func CalculatePaginationInfo(totalCount, limit, skip int64) map[string]interface{} {
	totalPages := (totalCount + limit - 1) / limit
	currentPage := (skip / limit) + 1

	return map[string]interface{}{
		"total":        totalCount,
		"limit":        limit,
		"skip":         skip,
		"total_pages":  totalPages,
		"current_page": currentPage,
	}
}

// ExtractGenreIDs extracts genre IDs from a slice of genres
func ExtractGenreIDs(genres []models.Genre) []int {
	ids := make([]int, len(genres))
	for i, genre := range genres {
		ids[i] = genre.GenreID
	}
	return ids
}