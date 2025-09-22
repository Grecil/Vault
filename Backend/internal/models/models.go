package models

import (
	"math/rand"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID           string         `json:"id" gorm:"primaryKey;type:varchar(255)"`
	Role         UserRole       `json:"role" gorm:"type:varchar(20);default:user"`
	StorageQuota int64          `json:"storage_quota" gorm:"default:10485760"` // 10MB default
	StorageUsed  int64          `json:"storage_used" gorm:"default:0"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`

	UserFiles []UserFile `json:"user_files" gorm:"foreignKey:UserID"`
}

type UserRole string

const (
	UserRoleUser  UserRole = "user"
	UserRoleAdmin UserRole = "admin"
)

type FileHash struct {
	Hash           string    `json:"hash" gorm:"primaryKey;type:varchar(64)"` // SHA256 hash
	Size           int64     `json:"size"`
	MimeType       string    `json:"mime_type" gorm:"type:varchar(255)"`
	ReferenceCount int       `json:"reference_count" gorm:"default:0"`
	MinIOKey       string    `json:"minio_key" gorm:"type:varchar(255)"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`

	UserFiles []UserFile `json:"user_files" gorm:"foreignKey:FileHash"`
}

type UserFile struct {
	ID            uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID        string         `json:"user_id" gorm:"type:varchar(255);not null;index"`
	FileHash      string         `json:"file_hash" gorm:"type:varchar(64);not null;index"`
	Filename      string         `json:"filename" gorm:"type:varchar(255);not null"`
	IsPublic      bool           `json:"is_public" gorm:"default:false"`
	DownloadCount int            `json:"download_count" gorm:"default:0"`
	UploadedAt    time.Time      `json:"uploaded_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `json:"-" gorm:"index"`

	User     User     `json:"user" gorm:"foreignKey:UserID"`
	FileData FileHash `json:"file_data" gorm:"foreignKey:FileHash"`
}

func (u *UserFile) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	u.UploadedAt = time.Now()
	return nil
}

// ShareLink represents a clean shareable link for public files
type ShareLink struct {
	ID         string         `json:"id" gorm:"primaryKey;type:varchar(8)"` // Short random ID
	UserFileID uuid.UUID      `json:"user_file_id" gorm:"type:uuid;not null;index"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `json:"-" gorm:"index"`

	UserFile UserFile `json:"user_file" gorm:"foreignKey:UserFileID"`
}

func (s *ShareLink) BeforeCreate(tx *gorm.DB) error {
	if s.ID == "" {
		s.ID = GenerateRandomID(8)
	}
	s.CreatedAt = time.Now()
	return nil
}

// GenerateRandomID creates a random alphanumeric ID of specified length
func GenerateRandomID(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[rand.Intn(len(charset))]
	}
	return string(result)
}
