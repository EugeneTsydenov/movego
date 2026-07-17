package profile

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newProfile(t *testing.T) *Profile {
	t.Helper()

	tag, err := NewTag("player123")
	require.NoError(t, err)

	displayName, err := NewDisplayName("Player")
	require.NoError(t, err)

	avatarURL, err := NewAvatarURL("https://example.com/avatar.png")
	require.NoError(t, err)

	country, err := NewCountryCode("RU")
	require.NoError(t, err)

	return New(uuid.New(), tag, displayName, avatarURL, country)
}

func TestNew(t *testing.T) {
	profile := newProfile(t)

	assert.Equal(t, "player123", profile.Tag().String())
	assert.Equal(t, "Player", profile.DisplayName().String())
	assert.Equal(t, "https://example.com/avatar.png", profile.AvatarURL().String())
	assert.Equal(t, "RU", profile.Country().String())
}

func TestUpdateDisplayName(t *testing.T) {
	profile := newProfile(t)

	err := profile.UpdateDisplayName("New Name")

	require.NoError(t, err)
	assert.Equal(t, "New Name", profile.DisplayName().String())
}

func TestUpdateDisplayNameInvalid(t *testing.T) {
	profile := newProfile(t)

	oldName := profile.DisplayName()

	err := profile.UpdateDisplayName("")

	assert.Error(t, err)
	assert.Equal(t, oldName, profile.DisplayName())
}

func TestUpdateAvatarURL(t *testing.T) {
	profile := newProfile(t)

	err := profile.UpdateAvatarURL("https://example.com/new.png")

	require.NoError(t, err)
	assert.Equal(t, "https://example.com/new.png", profile.AvatarURL().String())
}

func TestUpdateAvatarURLInvalid(t *testing.T) {
	profile := newProfile(t)

	oldURL := profile.AvatarURL()

	err := profile.UpdateAvatarURL("invalid-url")

	assert.Error(t, err)
	assert.Equal(t, oldURL, profile.AvatarURL())
}

func TestUpdateCountry(t *testing.T) {
	profile := newProfile(t)

	err := profile.UpdateCountry("US")

	require.NoError(t, err)
	assert.Equal(t, "US", profile.Country().String())
}

func TestUpdateCountryInvalid(t *testing.T) {
	profile := newProfile(t)

	oldCountry := profile.Country()

	err := profile.UpdateCountry("INVALID")

	assert.Error(t, err)
	assert.Equal(t, oldCountry, profile.Country())
}

func TestUpdateTagCooldown(t *testing.T) {
	profile := newProfile(t)

	err := profile.UpdateTag("newtag123")

	assert.ErrorIs(t, err, ErrTagChangeCooldown)
	assert.Equal(t, "player123", profile.Tag().String())
}

func TestUpdateTag(t *testing.T) {
	profile := newProfile(t)

	profile.tagUpdatedAt = time.Now().Add(-31 * 24 * time.Hour)

	err := profile.UpdateTag("newtag123")

	require.NoError(t, err)
	assert.Equal(t, "newtag123", profile.Tag().String())
}

func TestUpdateTagInvalid(t *testing.T) {
	profile := newProfile(t)

	profile.tagUpdatedAt = time.Now().Add(-31 * 24 * time.Hour)

	oldTag := profile.Tag()

	err := profile.UpdateTag("!")

	assert.Error(t, err)
	assert.Equal(t, oldTag, profile.Tag())
}

func TestCanUpdateTagAt(t *testing.T) {
	profile := newProfile(t)

	expected := profile.tagUpdatedAt.Add(tagUpdateCooldown)

	assert.Equal(t, expected, profile.CanUpdateTagAt())
}
