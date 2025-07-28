package models

import "time"

type Transaction struct {
	id   int
	from string
	to   string
	time time.Time
}
