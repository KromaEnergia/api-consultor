// internal/auth/refresh_model.go
package auth

import "time"

type RefreshToken struct {
  ID        uint       `gorm:"primaryKey"`
  UserID    uint       `gorm:"index"`
  FamilyID  string     `gorm:"index"`
  Hash      string     `gorm:"uniqueIndex"`
  IsAdmin   bool       // <<< novo
  ExpiresAt time.Time  `gorm:"index"`
  RevokedAt *time.Time
  CreatedAt time.Time
}
