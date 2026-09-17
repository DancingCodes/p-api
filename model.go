package main

import "time"

type Image struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"type:varchar(255)" json:"name"`
	Category  string    `gorm:"type:varchar(32);not null;default:'';index" json:"category"`
	Url       string    `gorm:"type:varchar(1024);not null" json:"url"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

type Video struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"type:varchar(255)" json:"name"`
	Category  string    `gorm:"type:varchar(32);not null;default:'';index" json:"category"`
	Url       string    `gorm:"type:varchar(1024);not null" json:"url"`
	CoverUrl  string    `gorm:"column:cover_url;type:varchar(1024);not null;default:''" json:"cover_url"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}
