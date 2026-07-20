package solitaire

import "time"

// dailyEpoch is the reference date for daily-deal numbering: Daily #0 is
// 2026-01-01. This is FROZEN FOREVER — changing it renumbers and reseeds every
// past daily, breaking everyone's streak and any shared links.
var dailyEpoch = time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

// DayNumber is the daily-deal index for now's local calendar date: the count
// of whole days from the epoch. The date is read in the player's local zone
// (so the puzzle rolls over at local midnight), then measured in UTC to keep
// the difference free of daylight-saving jumps.
func DayNumber(now time.Time) int {
	y, m, d := now.Date()
	today := time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
	return int(today.Sub(dailyEpoch).Hours()) / 24
}

// DailySeed derives the shared deal seed for a variant on a given day. Every
// player gets the same deal for a (variant, day) pair. The mixing is a frozen
// FNV-1a over the variant id and day followed by a splitmix64 avalanche —
// committed golden tests guard it against drift.
func DailySeed(id VariantID, day int) uint64 {
	const (
		offset = 1469598103934665603
		prime  = 1099511628211
	)
	h := uint64(offset)
	mix := func(b byte) {
		h ^= uint64(b)
		h *= prime
	}
	for i := 0; i < len(id); i++ {
		mix(id[i])
	}
	mix(':')
	// Offset the day so Daily #0 is not a trivial all-zero input.
	d := uint64(day) + 0x9E3779B97F4A7C15
	for i := 0; i < 8; i++ {
		mix(byte(d >> (8 * i)))
	}
	// splitmix64 finalizer for a good avalanche.
	h ^= h >> 30
	h *= 0xbf58476d1ce4e5b9
	h ^= h >> 27
	h *= 0x94d049bb133111eb
	h ^= h >> 31
	return h
}
