package util

import (
	"fmt"
	"time"
)

func TimeFormat() (string, error) {
    now := time.Now()
    
    moscowTZ, err := time.LoadLocation("Europe/Moscow")
    if err != nil {
        return "", fmt.Errorf("failed to load Moscow timezone: %w", err)
    }
    
    formatted := now.In(moscowTZ).Format("2006-01-02T15:04:05-07:00")
    
    return formatted, nil
}

func TimeNow() time.Time {
	return time.Now()
}

//new Date('2025-12-07T12:00:00Z');

func DifferenceTime(timeCreation time.Time) string { // hh : mm : ss
	diff := time.Since(timeCreation)
	result := diff.Round(time.Second)

	h := result / time.Hour
	result -= h * time.Hour
	m := result / time.Minute
	result -= m * time.Minute
	s := result / time.Second

	return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
}
