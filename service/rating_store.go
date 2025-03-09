package service

import "sync"

type RatingStore interface {
	Add(laptopID string, score float64) (*Rating, error)
}

type Rating struct {
	Count uint32
	Sum   float64
}

type InMemoryRatingStore struct {
	mutex   sync.RWMutex
	ratings map[string]*Rating
}

func NewInMemoryRatingStore() *InMemoryRatingStore {
	return &InMemoryRatingStore{
		ratings: make(map[string]*Rating),
	}
}

func (r *InMemoryRatingStore) Add(laptopID string, score float64) (*Rating, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	rating := r.ratings[laptopID]
	if rating == nil {
		rating = &Rating{}
		r.ratings[laptopID] = rating
	}
	rating.Count++
	rating.Sum += score
	return rating, nil
}
