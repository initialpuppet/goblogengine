package model

import (
	"time"

	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
)

// Statistics holds key statistics about the application.
type Statistics struct {
	PostCount     int
	DraftCount    int
	VersionCount  int
	AuthorCount   int
	CategoryCount int
	ImageCount    int

	Generated time.Time
}

const statisticsKind = "Statistics"

// generateStatistics runs counts on all the entities in the application, and
// caches the output in the datastore.
func generateStatistics(ctx context.Context) (*Statistics, error) {
	postCount, err := GetBlogPostCount(ctx)
	if err != nil {
		return nil, err
	}

	draftCount, err := GetBlogPostDraftCount(ctx)
	if err != nil {
		return nil, err
	}

	versionCount, err := GetBlogPostVersionCount(ctx)
	if err != nil {
		return nil, err
	}

	authorCount, err := GetAuthorCount(ctx)
	if err != nil {
		return nil, err
	}

	categoryCount, err := GetCategoryCount(ctx)
	if err != nil {
		return nil, err
	}

	imageCount, err := GetImageCount(ctx)
	if err != nil {
		return nil, err
	}

	s := Statistics{
		PostCount:     postCount,
		DraftCount:    draftCount,
		VersionCount:  versionCount,
		AuthorCount:   authorCount,
		CategoryCount: categoryCount,
		ImageCount:    imageCount,
		Generated:     time.Now(),
	}

	return &s, nil
}

// GetStatistics retrieves cached statistics and regenerates them if necessary.
func GetStatistics(ctx context.Context) (*Statistics, error) {
	q := datastore.NewQuery(statisticsKind).
		Order("-Generated").
		Limit(1)

	var stats []Statistics
	_, err := q.GetAll(ctx, &stats)
	if err != nil {
		return nil, err
	}

	if len(stats) < 1 || stats[0].Generated.Add(1*time.Minute).Before(time.Now()) {
		stat, err := generateStatistics(ctx)
		if err != nil {
			return nil, err
		}
		go stat.Save(ctx)
		return stat, nil
	}

	return &stats[0], nil
}

// Save saves the statistics to the datastore.
func (s *Statistics) Save(ctx context.Context) error {
	k := datastore.NewIncompleteKey(ctx, statisticsKind, blogRootKey(ctx))
	_, err := datastore.Put(ctx, k, s)
	return err
}
