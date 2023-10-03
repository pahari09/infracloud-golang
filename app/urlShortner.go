package app

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"infracloud-golang/infrastructure"
	"net/url"
)

type URLShortener struct {
	store infrastructure.Storage
}

type ScoredMember struct {
	Score  float64
	Member string
}

func NewURLShortener(store infrastructure.Storage) *URLShortener {
	return &URLShortener{store: store}
}

func (us *URLShortener) Shorten(originalURL string) (string, error) {
	hash := sha256.Sum256([]byte(originalURL))
	shortURL := hex.EncodeToString(hash[:])[:8]

	err := us.store.Set("url:"+shortURL, originalURL)
	if err != nil {
		return "", fmt.Errorf("failed to set URL in store: %w", err)
	}

	parsedURL, err := url.Parse(originalURL)
	if err != nil {
		return "", fmt.Errorf("failed to parse URL: %w", err)
	}

	domain := parsedURL.Hostname()

	err = us.store.ZIncrBy("domains", 1.0, domain)
	if err != nil {
		return "", fmt.Errorf("failed to increment domain count in store: %w", err)
	}

	return shortURL, nil
}

func (us *URLShortener) Redirect(shortURL string) (string, error) {
	return us.store.Get("url:" + shortURL)
}

func (us *URLShortener) Metrics() ([]ScoredMember, error) {
	zs, err := us.store.ZRevRangeWithScores("domains", 0, 2)
	if err != nil {
		return nil, err
	}

	var members []ScoredMember
	for _, z := range zs {
		members = append(members, ScoredMember{Score: z.Score, Member: fmt.Sprintf("%v", z.Member)})
	}
	return members, nil
}

func (us *URLShortener) ViewAll() (map[string]string, error) {
	data := make(map[string]string)
	cursor := uint64(0)

	for {
		keys, retCursor, err := us.store.Scan(cursor, "url:*", 10)
		if err != nil {
			return nil, fmt.Errorf("failed to scan keys from store: %w", err)
		}

		for _, key := range keys {
			value, err := us.store.Get(key)
			if err != nil {
				return nil, fmt.Errorf("failed to get value from store: %w", err)
			}
			data[key] = value
		}

		if retCursor == 0 {
			break
		}
		cursor = retCursor
	}

	return data, nil
}

func (us *URLShortener) DeleteAll() error {
	return us.store.FlushDB()
}
