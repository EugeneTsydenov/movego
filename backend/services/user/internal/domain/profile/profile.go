package profile

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

const tagUpdateCooldown = 30 * 24 * time.Hour

var (
	ErrTagChangeCooldown = errors.New("profile: tag can only be changed once every 30 days")
	ErrProfileNotFound   = errors.New("profile: not found")
)

type Profile struct {
	accountID    uuid.UUID
	tag          Tag
	tagUpdatedAt time.Time
	displayName  DisplayName
	avatarURL    AvatarURL
	country      CountryCode
	createdAt    time.Time
	updatedAt    time.Time
}

func New(accountID uuid.UUID, tag Tag, displayName DisplayName, avatarURL AvatarURL, country CountryCode) *Profile {
	now := time.Now().UTC()
	return &Profile{
		accountID:    accountID,
		tag:          tag,
		tagUpdatedAt: now,
		displayName:  displayName,
		avatarURL:    avatarURL,
		country:      country,
		createdAt:    now,
		updatedAt:    now,
	}
}

func (p *Profile) AccountID() uuid.UUID     { return p.accountID }
func (p *Profile) Tag() Tag                 { return p.tag }
func (p *Profile) DisplayName() DisplayName { return p.displayName }
func (p *Profile) AvatarURL() AvatarURL     { return p.avatarURL }
func (p *Profile) Country() CountryCode     { return p.country }
func (p *Profile) CreatedAt() time.Time     { return p.createdAt }
func (p *Profile) UpdatedAt() time.Time     { return p.updatedAt }

func (p *Profile) UpdateTag(tagStr string) error {
	if time.Since(p.tagUpdatedAt) < tagUpdateCooldown {
		return ErrTagChangeCooldown
	}

	tag, err := NewTag(tagStr)
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	p.tag = tag
	p.tagUpdatedAt = now
	p.updatedAt = now
	return nil
}

func (p *Profile) CanUpdateTagAt() time.Time {
	return p.tagUpdatedAt.Add(tagUpdateCooldown)
}

func (p *Profile) UpdateDisplayName(name string) error {
	displayName, err := NewDisplayName(name)
	if err != nil {
		return err
	}

	p.displayName = displayName
	p.updatedAt = time.Now().UTC()
	return nil
}

func (p *Profile) UpdateAvatarURL(url string) error {
	avatarURL, err := NewAvatarURL(url)
	if err != nil {
		return err
	}
	p.avatarURL = avatarURL
	p.updatedAt = time.Now().UTC()
	return nil
}

func (p *Profile) UpdateCountry(country string) error {
	countryCode, err := NewCountryCode(country)
	if err != nil {
		return err
	}
	p.country = countryCode
	p.updatedAt = time.Now().UTC()
	return nil
}
