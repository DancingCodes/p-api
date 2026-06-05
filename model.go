package main

import "time"

type Image struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"type:varchar(255);not null" json:"name"`
	Url       string    `gorm:"type:varchar(1024)" json:"url"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}
