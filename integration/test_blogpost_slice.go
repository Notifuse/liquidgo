package integration

import "time"

// BlogPost represents a blog post with various properties for testing slice type compatibility.
type BlogPost struct {
	Title         string
	Author        string
	Content       string
	CreatedAt     time.Time
	URL           string
	CommentsCount int
	Published     bool
	Rating        float64
}

// NewBlogPost creates a new BlogPost with the given values.
func NewBlogPost(title, author, content, url string, createdAt time.Time, commentsCount int, published bool, rating float64) BlogPost {
	return BlogPost{
		Title:         title,
		Author:        author,
		Content:       content,
		CreatedAt:     createdAt,
		URL:           url,
		CommentsCount: commentsCount,
		Published:     published,
		Rating:        rating,
	}
}

// SampleBlogPosts returns a slice of sample BlogPost instances for testing.
func SampleBlogPosts() []BlogPost {
	baseTime := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	return []BlogPost{
		NewBlogPost(
			"Getting Started with Go",
			"Alice",
			"This is a comprehensive guide to getting started with Go programming.",
			"/blog/getting-started-go",
			baseTime,
			5,
			true,
			4.5,
		),
		NewBlogPost(
			"Advanced Liquid Templates",
			"Bob",
			"Learn advanced techniques for working with Liquid templates.",
			"/blog/advanced-liquid",
			baseTime.AddDate(0, 0, 1),
			12,
			true,
			4.8,
		),
		NewBlogPost(
			"Understanding Reflection",
			"Charlie",
			"A deep dive into Go's reflection capabilities.",
			"/blog/understanding-reflection",
			baseTime.AddDate(0, 0, 2),
			8,
			false,
			4.2,
		),
		NewBlogPost(
			"Testing Best Practices",
			"Alice",
			"Best practices for writing effective tests in Go.",
			"/blog/testing-best-practices",
			baseTime.AddDate(0, 0, 3),
			15,
			true,
			4.9,
		),
		NewBlogPost(
			"Performance Optimization",
			"Bob",
			"Tips and tricks for optimizing Go application performance.",
			"/blog/performance-optimization",
			baseTime.AddDate(0, 0, 4),
			3,
			true,
			4.0,
		),
	}
}

// BlogPostsToInterfaceSlice converts a []BlogPost to []interface{} for comparison tests.
func BlogPostsToInterfaceSlice(posts []BlogPost) []interface{} {
	result := make([]interface{}, len(posts))
	for i, post := range posts {
		result[i] = post
	}
	return result
}
